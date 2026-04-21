package ohlc

import (
	"github.com/encratite/commons"
)

func ReadMEXC(symbol string, directory string, timeFrame TimeFrame) ([]Record, error) {
	return ReadBarchart(symbol, directory, timeFrame)
}

func MustReadMEXC(symbol string, directory string, timeFrame TimeFrame) []Record {
	records, err := ReadMEXC(symbol, directory, timeFrame)
	if err != nil {
		commons.Fatalf("%v", err)
	}
	return records
}