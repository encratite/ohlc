package ohlc

import (
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

func ReadBinanceRecords(symbol string, directory string) []Record {
	paths := commons.GetFiles(directory, ".csv")
	records := []Record{}
	for _, path := range paths {
		fileName := filepath.Base(path)
		if !strings.Contains(fileName, symbol) {
			continue
		}
		commons.ReadCSV(path, func (row []string) {
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
			records = append(records, record)
		})
	}
	slices.SortFunc(records, func (a, b Record) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
	return records
}