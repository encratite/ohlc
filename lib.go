package ohlc

import (
	"time"
)

type TimeFrame int

const (
	TimeFrameD1 TimeFrame = iota
	TimeFrameH1
	TimeFrameM15
)

type Record struct {
	Timestamp time.Time `yaml:"timestamp"`
	Open float64 `yaml:"open"`
	High float64 `yaml:"high"`
	Low float64 `yaml:"low"`
	Close float64 `yaml:"close"`
}