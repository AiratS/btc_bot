package main

import "time"

type Ticker struct {
	IntervalMinutes int
	lastMinute      int
}

func NewTicker(intervalMinutes int) *Ticker {
	return &Ticker{
		IntervalMinutes: intervalMinutes,
		lastMinute:      time.Now().Minute(),
	}
}

func (ticker *Ticker) tick() bool {
	currentMinute := time.Now().Minute()

	isChanged := ticker.lastMinute != currentMinute
	if isChanged {
		ticker.lastMinute = currentMinute
	}

	return isChanged
}
