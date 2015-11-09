// Handles data reporting and bouncing. Reports data in TSV files, while
// bouncing off repeating data entries.
//
// Prior to calls to this package, one should call the Initialize() function.
// After finishing, one should call Finalize() to ensure all buffers are
// flushed and all resources are released. Multiple rounds of Initialize and
// Finalize are allowed.
//
// Input data to this package should be sorted ascendingly by time, to allow
// proper bouncing.
package bouncer

import (
	"fmt"
	"hash/crc64"
)

// Root path for data output.
var outDir string

// Initializes the bouncing mechanism. Should be called prior to any call to the
// package.
func Initialize(dir string) {
	outDir = dir
	initItems()
	initItemsMeta()
	initPrices()
	initStores()
	initStoresMeta()
	initPromos()
}

// Flushes and closes all streams used by this package, and terminates all
// goroutines. Should be called after using the package to ensure all data were
// written and no goroutines leak.
func Finalize() {
	finalizeItems()
	finalizeItemsMeta()
	finalizePrices()
	finalizeStores()
	finalizeStoresMeta()
	finalizePromos()
}

// Used for hashing stuff.
var crcTable = crc64.MakeTable(crc64.ECMA)

// Returns the hash generated by printing the given arguments one after another.
func hash(a interface{}, b ...interface{}) int {
	crc := crc64.New(crcTable)

	fmt.Fprint(crc, a)
	for _, element := range b {
		fmt.Fprint(crc, element)
	}

	return int(crc.Sum64())
}

