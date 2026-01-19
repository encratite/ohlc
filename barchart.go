package ohlc

import (
	"fmt"
	"log"
	"path/filepath"
	"slices"

	"github.com/encratite/commons"
)

func ReadBarchart(symbol string, timeFrame TimeFrame, directory string) []Record {
	var suffix string
	switch timeFrame {
	case TimeFrameD1:
		suffix = "D1"
	case TimeFrameH1:
		suffix = "H1"
	case TimeFrameM15:
		suffix = "M15"
	default:
		commons.Fatalf("Unknown time frame: %d", timeFrame)
	}
	fileName := fmt.Sprintf("%s.%s.csv", symbol, suffix)
	path := filepath.Join(directory, fileName)
	records := []Record{}
	columns := []string{
		"time",
		"open",
		"high",
		"low",
		"close",
	}
	line := 2
	commons.ReadCSVColumns(path, columns, func(cells []string) {
		timestamp, err := commons.ParseTime(cells[0])
		if err != nil {
			commons.Fatalf("Invalid timestamp on line %d in file %s", line, path)
		}
		readFloat := func (name string, index int) float64 {
			value, err := commons.ParseFloat(cells[index])
			if err != nil {
				commons.Fatalf("Invalid %s value on line %d in file %s", name, line, path)
			}
			if value <= 0 {
				log.Fatalf("Invalid %s value of %.2f at %s in %s", name, value, commons.GetTimeString(timestamp), path)
			}
			return value
		}
		open := readFloat("open", 1)
		high := readFloat("high", 2)
		low := readFloat("low", 3)
		close := readFloat("close", 4)
		record := Record{
			Timestamp: timestamp,
			Open: open,
			High: high,
			Low: low,
			Close: close,
		}
		records = append(records, record)
		line++
	})
	slices.SortFunc(records, func (a, b Record) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
	return records
}