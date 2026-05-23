package ohlc

import (
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"time"

	"github.com/encratite/commons"
)

type Futures struct {
	activeSymbols map[time.Time]string
	recordsMap map[string]*futuresRecords
}

type futuresRecords struct {
	records []Record
	indexes map[time.Time]int
}

func ReadFutures(symbol, directory, futuresDirectory string, timeFrame TimeFrame) (Futures, error) {
	dailySuffix, err := getSuffix(TimeFrameD1)
	if err != nil {
		return Futures{}, err
	}
	dailyFileName := fmt.Sprintf("%s.%s.csv", symbol, dailySuffix)
	suffix, err := getSuffix(timeFrame)
	if err != nil {
		return Futures{}, err
	}
	dailyPath := filepath.Join(directory, dailyFileName)
	columns := []string{
		"symbol",
		"time",
	}
	activeSymbols := map[time.Time]string{}
	var csvErr error
	readErr := commons.ReadCSVColumns(dailyPath, columns, func(cells []string) {
		symbol := cells[0]
		timestamp, err := commons.ParseTime(cells[1])
		if err != nil {
			if csvErr != nil {
				csvErr = err
			}
			return
		}
		activeSymbols[timestamp] = symbol
	})
	if readErr != nil {
		return Futures{}, readErr
	}
	if csvErr != nil {
		return Futures{}, csvErr
	}
	fileName := fmt.Sprintf("%s.%s.csv", symbol, suffix)
	futuresPath := filepath.Join(futuresDirectory, fileName)
	recordsMap := map[string]*futuresRecords{}
	ohlcColumns := []string{
		"symbol",
		"time",
		"open",
		"high",
		"low",
		"close",
	}
	line := 2
	readErr = commons.ReadCSVColumns(futuresPath, ohlcColumns, func(cells []string) {
		symbol := cells[0]
		timestamp, err := commons.ParseTime(cells[1])
		if err != nil {
			if csvErr != nil {
				csvErr = fmt.Errorf("Invalid timestamp on line %d in file %s", line, dailyPath)
			}
			return
		}
		readFloat := func (name string, index int) float64 {
			value, err := commons.ParseFloat(cells[index])
			if err != nil {
				if csvErr != nil {
					csvErr = fmt.Errorf("Invalid %s value on line %d in file %s", name, line, dailyPath)
				}
				return math.NaN()
			}
			if value <= 0 {
				if csvErr != nil {
					csvErr = fmt.Errorf("Invalid %s value of %.2f at %s in %s", name, value, commons.GetTimeString(timestamp), dailyPath)
				}
				return math.NaN()
			}
			return value
		}
		open := readFloat("open", 2)
		high := readFloat("high", 3)
		low := readFloat("low", 4)
		close := readFloat("close", 5)
		record := Record{
			Timestamp: timestamp,
			Open: open,
			High: high,
			Low: low,
			Close: close,
		}
		fRecords, exists := recordsMap[symbol]
		if !exists {
			f := futuresRecords{
				records: []Record{},
				indexes: map[time.Time]int{},
			}
			fRecords = &f
			recordsMap[symbol] = fRecords
		}
		fRecords.records = append(fRecords.records, record)
		line++
	})
	for key := range recordsMap {
		fRecords := recordsMap[key]
		slices.SortFunc(fRecords.records, func (a, b Record) int {
			return a.Timestamp.Compare(b.Timestamp)
		})
		for i, record := range fRecords.records {
			fRecords.indexes[record.Timestamp] = i
		}
	}
	if readErr != nil {
		return Futures{}, readErr
	}
	if csvErr != nil {
		return Futures{}, csvErr
	}
	futures := Futures{
		activeSymbols: activeSymbols,
		recordsMap: recordsMap,
	}
	return futures, nil
}

func MustReadFutures(symbol, directory, futuresDirectory string, timeFrame TimeFrame) Futures {
	futures, err := ReadFutures(symbol, directory, futuresDirectory, timeFrame)
	if err != nil {
		commons.Fatalf("%v", err)
	}
	return futures
}

func (f *Futures) GetRecords(timestamp time.Time) ([]Record, int, bool, string) {
	date := commons.GetDate(timestamp)
	symbol, exists := f.activeSymbols[date]
	if !exists {
		return nil, 0, false, ""
	}
	fRecords, exists := f.recordsMap[symbol]
	if !exists {
		return nil, 0, false, ""
	}
	index, exists := fRecords.indexes[timestamp]
	if !exists {
		return nil, 0, false, ""
	}
	return fRecords.records, index, true, symbol
}