package ohlc

import (
	"fmt"
	"math"
	"path/filepath"
	"slices"

	"github.com/encratite/commons"
)

func ReadBarchart(symbol string, directory string, timeFrame TimeFrame) ([]Record, error) {
	var suffix string
	switch timeFrame {
	case TimeFrameD1:
		suffix = "D1"
	case TimeFrameH1:
		suffix = "H1"
	case TimeFrameM30:
		suffix = "M30"
	case TimeFrameM15:
		suffix = "M15"
	default:
		return nil, fmt.Errorf("Unknown time frame: %d", timeFrame)
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
	var csvError error
	commons.ReadCSVColumns(path, columns, func(cells []string) {
		timestamp, err := commons.ParseTime(cells[0])
		if err != nil {
			if csvError != nil {
				csvError = fmt.Errorf("Invalid timestamp on line %d in file %s", line, path)
			}
			return
		}
		readFloat := func (name string, index int) float64 {
			value, err := commons.ParseFloat(cells[index])
			if err != nil {
				if csvError != nil {
					csvError = fmt.Errorf("Invalid %s value on line %d in file %s", name, line, path)
				}
				return math.NaN()
			}
			if value <= 0 {
				if csvError != nil {
					csvError = fmt.Errorf("Invalid %s value of %.2f at %s in %s", name, value, commons.GetTimeString(timestamp), path)
				}
				return math.NaN()
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
	if csvError != nil {
		return nil, csvError
	}
	slices.SortFunc(records, func (a, b Record) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
	return records, nil
}

func MustReadBarchart(symbol string, directory string, timeFrame TimeFrame) []Record {
	records, err := ReadBarchart(symbol, directory, timeFrame)
	if err != nil {
		commons.Fatalf("%v", err)
	}
	return records
}