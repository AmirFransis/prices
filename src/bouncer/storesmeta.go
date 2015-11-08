package bouncer

// Handles reporting & bouncing of store metadata.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	storeMetaOut    *os.File          // Output file.
	storeMetaOutBuf *bufio.Writer     // Output buffer.
	storeMetaChan   chan []*StoreMeta // Used for reporting store-metas.
	storeMetaDone   chan int          // Indicates when meta reporting is
	                                  // finished.
	storeMetaMap    map[int]int       // Maps StoreId to hash.
)

// Initializes the 'stores_meta' table bouncer.
func initStoresMeta() {
	storeMetaChan = make(chan []*StoreMeta, runtime.NumCPU())
	storeMetaDone = make(chan int, 1)
	storeMetaMap = map[int]int{}

	var err error
	storeMetaOut, err = os.Create(filepath.Join(outDir, "stores_meta.txt"))
	if err != nil {
		panic(err)
	}
	storeMetaOutBuf = bufio.NewWriter(storeMetaOut)

	go func() {
		for metas := range storeMetaChan {
			reportStoreMetas(metas)
		}
		storeMetaDone <- 0
	}()
}

// Finalizes the 'stores_meta' table bouncer.
func finalizeStoresMeta() {
	close(storeMetaChan)
	<-storeMetaDone
	storeMetaOutBuf.Flush()
	storeMetaOut.Close()
}

// A single entry in the 'stores_meta' table.
type StoreMeta struct {
	Timestamp      int64
	StoreId        int
	BikoretNo      string
	StoreType      string
	ChainName      string
	SubchainName   string
	StoreName      string
	Address        string
	City           string
	ZipCode        string
	LastUpdateDate string
	LastUpdateTime string
}

// Returns the hash of an store-meta entry.
func (s *StoreMeta) hash() int {
	return hash(
		s.BikoretNo,
		s.StoreType,
		s.ChainName,
		s.SubchainName,
		s.StoreName,
		s.Address,
		s.City,
		s.ZipCode,
		s.LastUpdateDate,
		s.LastUpdateTime,
	)
}

// Reports the given metas.
func ReportStoreMetas(ss []*StoreMeta) {
	storeMetaChan <- ss
}

// Reports the given metas. Called by the goroutine that listens on the channel.
func reportStoreMetas(ss []*StoreMeta) {
	for i := range ss {
		h := ss[i].hash()
		last := storeMetaMap[ss[i].StoreId]
		if h != last {
			storeMetaMap[ss[i].StoreId] = h
			fmt.Fprintf(storeMetaOutBuf,
				"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
				ss[i].Timestamp,
				ss[i].StoreId,
				ss[i].BikoretNo,
				ss[i].StoreType,
				ss[i].ChainName,
				ss[i].SubchainName,
				ss[i].StoreName,
				ss[i].Address,
				ss[i].City,
				ss[i].ZipCode,
				ss[i].LastUpdateDate,
				ss[i].LastUpdateTime,
			)
		}
	}
}

