package ohlc

import (
	"github.com/encratite/commons"
)

func ReadHyperliquid(symbol string, directory string, timeFrame TimeFrame) ([]Record, error) {
	return ReadBarchart(symbol, directory, timeFrame)
}

func MustReadHyperliquid(symbol string, directory string, timeFrame TimeFrame) []Record {
	records, err := ReadHyperliquid(symbol, directory, timeFrame)
	if err != nil {
		commons.Fatalf("%v", err)
	}
	return records
}