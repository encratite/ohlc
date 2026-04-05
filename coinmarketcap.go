package ohlc

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/encratite/commons"
)

type CoinMarketCap struct {
	Data CoinMarketCapData `json:"data"`
}

type CoinMarketCapData struct {
	Points []CoinMarketCapPoint `json:"points"`
}

type CoinMarketCapPoint struct {
	Timestamp string `json:"s"`
	Values []float64 `json:"v"`
}

func ReadCoinMarketCap(symbol string, directory string, timeFrame TimeFrame) ([]Record, error) {
	suffix, err := getSuffix(timeFrame)
	if err != nil {
		return nil, err
	}
	fileName := fmt.Sprintf("%s.%s.json", symbol, suffix)
	path := filepath.Join(directory, fileName)
	data := commons.ReadJSON[CoinMarketCap](path)
	records := []Record{}
	for _, point := range data.Data.Points {
		unixSeconds, err := commons.ParseInt64(point.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse UNIX timestamp: %s", point.Timestamp)
		}
		timestamp := time.Unix(unixSeconds, 0)
		if len(point.Values) != 3 {
			return nil, fmt.Errorf("Invalid point values: %v", point.Values)
		}
		close := point.Values[0]
		var open float64
		if len(records) > 0 {
			previous := records[len(records) - 1]
			open = previous.Close
		} else {
			open = close
		}
		high := max(open, close)
		low := min(open, close)
		record := Record{
			Timestamp: timestamp,
			Open: open,
			High: high,
			Low: low,
			Close: close,
		}
		records = append(records, record)
	}
	return records, nil
}

func MustReadCoinMarketCap(symbol string, directory string, timeFrame TimeFrame) []Record {
	records, err := ReadCoinMarketCap(symbol, directory, timeFrame)
	if err != nil {
		commons.Fatalf("%v", err)
	}
	return records
}