package ohlc

import (
	"fmt"
	"regexp"
	"time"

	"github.com/encratite/commons"
)

const (
	tradingDaysFirstYear = 1998
	tradingDaysLastYear = 2026
)

type TradingDays struct {
	days map[time.Time]TradingDayIndex
}

type TradingDayIndex struct {
	Positive int
	Negative int
}

func LoadTradingDays(path string) (TradingDays, error) {
	columns := []string{
		"TradeDate",
	}
	pattern := regexp.MustCompile("^(\\d{4})(\\d{2})(\\d{2})$")
	var err error
	holidays := map[time.Time]struct{}{}
	commons.ReadCSVColumns(path, columns, func (cells []string) {
		line := cells[0]
		match := pattern.FindStringSubmatch(line)
		if match == nil {
			err = fmt.Errorf("Failed to parse line: \"%s\"", line)
			return
		}
		year := commons.MustParseInt(match[1])
		month := commons.MustParseInt(match[2])
		day := commons.MustParseInt(match[3])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		holidays[date] = struct{}{}
	})
	if err != nil {
		return TradingDays{}, err
	}
	days := map[time.Time]TradingDayIndex{}
	date := time.Date(tradingDaysFirstYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	hasPreviousDate := false
	var previousDate time.Time
	index := 0
	tradingDates := []time.Time{}
	lastDay := time.Date(tradingDaysLastYear + 1, time.January, 1, 0, 0, 0, 0, time.UTC)
	for date.Before(lastDay) {
		_, exists := holidays[date]
		weekday := date.Weekday()
		isTradingDay := weekday != time.Saturday && weekday != time.Sunday && !exists
		if isTradingDay {
			days[date] = TradingDayIndex{
				Positive: index,
				Negative: -1,
			}
			tradingDates = append(tradingDates, date)
			if hasPreviousDate && date.Month() != previousDate.Month() {
				index = 0
			} else {
				index++
			}
			previousDate = date
			hasPreviousDate = true
		}
		date = date.AddDate(0, 0, 1)
	}
	finalIndex := len(tradingDates) - 1
	negativeIndex := -1
	for i := 0; i < finalIndex; i++ {
		offset := finalIndex - i
		date1 := tradingDates[offset]
		date2 := tradingDates[offset - 1]
		indexData1, exists := days[date1]
		if !exists {
			continue
		}
		indexData2, exists := days[date2]
		if !exists {
			continue
		}
		if indexData1.Positive == 0 {
			negativeIndex = -1
		} else {
			negativeIndex--
		}
		indexData2.Negative = negativeIndex
		days[date2] = indexData2
	}
	tradingDays := TradingDays{
		days: days,
	}
	return tradingDays, nil
}

func MustLoadTradingDays(path string) TradingDays {
	tradingDays, err := LoadTradingDays(path)
	if err != nil {
		commons.Fatalf("Failed to load trading days: %v", err)
	}
	return tradingDays
}

func (t *TradingDays) GetIndex(date time.Time) (TradingDayIndex, bool) {
	index, exists := t.days[date]
	return index, exists
}