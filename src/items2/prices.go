package main

import (
	"fmt"
	"os"
	"bufio"
	"runtime"
)

func init() {
	var err error
	priceDataOut, err = os.Create("/cs/icore/amitlavon/prices.txt")
	if err != nil { panic(err) }
	priceDataOutBuf = bufio.NewWriter(priceDataOut)
	
	go func() {
		for prices := range priceDataChan {
			reportPriceData(prices)
		}
		priceDatadone <- 0
	}()
}

var priceDataChan = make(chan []*priceData, runtime.NumCPU())
var priceDatadone = make(chan int, 1)

func priceDataFinalize() {
	close(priceDataChan)
	<-priceDatadone
	priceDataOutBuf.Flush()
	priceDataOut.Close()
}

var priceDataOut *os.File
var priceDataOutBuf *bufio.Writer

type priceData struct {
	timestamp int64
	itemId int
	storeId int
	price string
	unitOfMeasurePrice string
	unitOfMeasure string
	quantity string
};

func (p *priceData) hash() int {
	return hash(
		p.price,
		p.unitOfMeasurePrice,
		p.unitOfMeasure,
		p.quantity,
	)
}

func (p *priceData) id() string {
	return fmt.Sprintf("%s,%s", p.itemId, p.storeId)
}

// Maps itemId,storeId to hash.
var priceDataMap = map[string]int {}

func reportPriceData(ps []*priceData) {
	for i := range ps {
		h := ps[i].hash()
		last := priceDataMap[ps[i].id()]
		if h != last {
			priceDataMap[ps[i].id()] = h
			fmt.Fprintf(priceDataOutBuf, "%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
					ps[i].timestamp,
					ps[i].itemId,
					ps[i].storeId,
					ps[i].price,
					ps[i].unitOfMeasurePrice,
					ps[i].unitOfMeasure,
					ps[i].quantity,
			)
		}
	}
}


