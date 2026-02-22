package ohlc

import (
	"math"
	"time"

	"github.com/encratite/commons"
)

const (
	volatilityStep = time.Duration(12) * time.Hour
	volatilityMinSamples = 15
	forceInverseRisk = false
)

type TimeFrame int

const (
	TimeFrameD1 TimeFrame = iota
	TimeFrameH1
	TimeFrameM30
	TimeFrameM15
)

type Record struct {
	Timestamp time.Time `yaml:"timestamp"`
	Open float64 `yaml:"open"`
	High float64 `yaml:"high"`
	Low float64 `yaml:"low"`
	Close float64 `yaml:"close"`
}

func RiskParity(series [][]Record, weights []float64, step time.Duration, volatilityWindow time.Duration) []Record {
	if len(series) != len(weights) {
		commons.Fatalf("Number of record series and weights don't match")
	}
	var startTime, endTime time.Time
	seriesMap := []map[time.Time]float64{}
	for i, records := range series {
		if len(records) < 2 {
			commons.Fatalf("Not enough samples in record series")
		}
		recordStartTime := records[0].Timestamp
		recordEndTime := records[len(records) - 1].Timestamp
		if i > 0 {
			if recordStartTime.Before(startTime) {
				startTime = recordStartTime
			}
			if recordEndTime.After(endTime) {
				endTime = recordEndTime
			}
		} else {
			startTime = recordStartTime
			endTime = recordEndTime
		}
		weight := weights[i]
		returnsMap := map[time.Time]float64{}
		for j := 1; j < len(records); j++ {
			record := records[j]
			previousRecord := records[j - 1]
			var close1, close2 float64
			if weight > 0 {
				close1 = record.Close
				close2 = previousRecord.Close
			} else {
				close1 = previousRecord.Close
				close2 = record.Close
			}
			returns := math.Abs(weight) * (close1 / close2 - 1.0)
			// fmt.Printf("%s series = %d, weight = %.1f, record.Close = %.2f, previousRecord.Close = %.2f, returns = %.5f\n", commons.GetTimeString(record.Timestamp), i, weight, record.Close, previousRecord.Close, returns)
			returnsMap[record.Timestamp] = returns
		}
		seriesMap = append(seriesMap, returnsMap)
	}
	var seriesVolatility []float64
	price := 1.0
	output := []Record{}
	for timestamp := startTime; timestamp.Before(endTime); timestamp = timestamp.Add(step) {
		weekday := timestamp.Weekday()
		timeOfDay := commons.GetTimeOfDay(timestamp)
		if weekday == time.Monday && timeOfDay == 0 {
			seriesVolatility = []float64{}
			for _, returnsMap := range seriesMap {
				volatility := getVolatility(timestamp, returnsMap, volatilityWindow)
				if math.IsNaN(volatility) {
					seriesVolatility = []float64{}
					break
				}
				// fmt.Printf("%s New volatility for i = %d: %.5f\n", commons.GetTimeString(timestamp), i, volatility)
				seriesVolatility = append(seriesVolatility, volatility)
			}
		}
		useInverseRisk := forceInverseRisk || len(weights) > 2
		if len(seriesVolatility) > 0 {
			totalWeight := 0.0
			for _, volatility := range seriesVolatility {
				if useInverseRisk {
					totalWeight += 1.0 / volatility
				} else {
					totalWeight += volatility
				}
			}
			missingRecord := false
			syntheticReturns := 0.0
			for i, returnsMap := range seriesMap {
				returns, exists := returnsMap[timestamp]
				if !exists {
					missingRecord = true
					break
				}
				var weight float64
				if useInverseRisk {
					weight = 1.0 / seriesVolatility[i]
				} else {
					weight = seriesVolatility[len(seriesVolatility) - 1]
				}
				syntheticReturns += weight / totalWeight * returns
			}
			if missingRecord {
				continue
			}
			previousPrice := price
			price *= 1.0 + syntheticReturns
			record := Record{
				Timestamp: timestamp,
				Open: previousPrice,
				High: math.NaN(),
				Low: math.NaN(),
				Close: price,
			}
			output = append(output, record)
		}
	}
	return output
}

func getVolatility(timestamp time.Time, returnsMap map[time.Time]float64, volatilityWindow time.Duration) float64 {
	returns := []float64{}
	startTime := commons.GetDate(timestamp.Add(- volatilityWindow))
	endTime := commons.GetDate(timestamp)
	for timestamp := startTime; timestamp.Before(endTime); timestamp = timestamp.Add(volatilityStep) {
		r, exists := returnsMap[timestamp]
		if exists {
			returns = append(returns, r)
		}
	}
	if len(returns) < volatilityMinSamples {
		return math.NaN()
	}
	volatility := commons.StdDev(returns)
	return volatility
}