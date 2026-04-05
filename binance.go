package ohlc

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/encratite/commons"
)

const (
	secondsPerHour = 60 * 60
	binanceTimestampSwitch = 1735689600000000
)

func ReadBinance(symbol string, directory string, timeFrame TimeFrame) ([]Record, error) {
	symbolDirectory := filepath.Join(directory, symbol)
	paths, err := commons.GetFiles(symbolDirectory, ".zip")
	if err != nil {
		return nil, err
	}
	recordsMap := map[time.Time]Record{}
	var timeFrameString string
	switch timeFrame {
	case TimeFrameD1:
		timeFrameString = "1d"
	case TimeFrameH1:
		timeFrameString = "1h"
	case TimeFrameM30:
		timeFrameString = "30m"
	case TimeFrameM15:
		timeFrameString = "15m"
	case TimeFrameM1:
		timeFrameString = "1m"
	default:
		return nil, fmt.Errorf("Unknown time frame in ReadBinance: %d", timeFrame)
	}
	timeFrameFilter := fmt.Sprintf("-%s-", timeFrameString)
	for _, path := range paths {
		fileName := filepath.Base(path)
		if !strings.Contains(fileName, symbol) || !strings.Contains(fileName, timeFrameFilter) {
			continue
		}
		zipReader, err := zip.OpenReader(path)
		if err != nil {
			return nil, fmt.Errorf("Unable to read zip file %s: %v", path, err)
		}
		defer zipReader.Close()
		if len(zipReader.File) != 1 {
			return nil, fmt.Errorf("Unexpected number of files in zip file %s", path)
		}
		file := zipReader.File[0]
		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("Failed to open .csv file within zip file %s: %v", path, err)
		}
		defer fileReader.Close()
		commons.ReadCSVFile(fileReader, func (row []string) {
			unixTimestamp := commons.MustParseInt64(row[0])
			if unixTimestamp >= binanceTimestampSwitch {
				unixTimestamp /= 1000
			}
			seconds := unixTimestamp / 1000 + secondsPerHour
			timestamp := time.Unix(seconds, 0).UTC()
			open := commons.MustParseFloat(row[1])
			high := commons.MustParseFloat(row[2])
			low := commons.MustParseFloat(row[3])
			close := commons.MustParseFloat(row[4])
			record := Record{
				Timestamp: timestamp,
				Open: open,
				High: high,
				Low: low,
				Close: close,
			}
			recordsMap[record.Timestamp] = record
		})
	}
	records := []Record{}
	for _, record := range recordsMap {
		records = append(records, record)
	}
	slices.SortFunc(records, func (a, b Record) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
	if len(records) == 0 {
		return nil, fmt.Errorf("Failed to load any records for symbol %s", symbol)
	}
	return records, nil
}

func MustReadBinance(symbol string, directory string, timeFrame TimeFrame) []Record {
	records, err := ReadBinance(symbol, directory, timeFrame)
	if err != nil {
		commons.Fatalf("%v", err)
	}
	return records
}