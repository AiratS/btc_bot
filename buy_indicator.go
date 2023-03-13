package main

import (
	"fmt"
	"github.com/markcheno/go-talib"
)

type BuyIndicator interface {
	HasSignal() bool
	IsStarted() bool
	Start()
	Update()
	Finish()
}

type BackTrailingBuyIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database

	isStarted                bool
	hasSignal                bool
	lastPrice                float64
	upperStopPrice           float64
	updatesCount             int
	updatedTimesBeforeFinish int
}

func NewBackTrailingBuyIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) BackTrailingBuyIndicator {
	return BackTrailingBuyIndicator{
		config: config,
		buffer: buffer,
		db:     db,

		isStarted:                false,
		hasSignal:                false,
		updatesCount:             0,
		updatedTimesBeforeFinish: 0,
	}
}

func (indicator *BackTrailingBuyIndicator) IsStarted() bool {
	return indicator.isStarted
}

func (indicator *BackTrailingBuyIndicator) Start() {
	candle := indicator.buffer.GetLastCandle()

	msg := fmt.Sprintf("trailing_STARTED: %s\nExchangeRate: %f", candle.CloseTime, candle.ClosePrice)
	if IS_REAL_ENABLED {
		LogAndPrintAndSendTg(msg)
	} else {
		LogAndPrint(msg)
	}

	price := indicator.buffer.GetLastCandleClosePrice()
	indicator.upperStopPrice = indicator.calculateStopPrice(
		price,
		indicator.config.TrailingTopPercentage,
	)

	indicator.isStarted = true
	indicator.lastPrice = price
	indicator.hasSignal = false
	indicator.updatesCount = 0
	indicator.updatedTimesBeforeFinish = 0
	indicator.updatesCount = 0
}

func (indicator *BackTrailingBuyIndicator) Update() {
	if !indicator.IsStarted() {
		indicator.Start()
	}

	indicator.updatesCount++
	currentPrice := indicator.buffer.GetLastCandleClosePrice()
	if indicator.updatesCount <= 1 {
		return
	}

	if indicator.isGrowing() {
		indicator.hasSignal = indicator.upperStopPrice <= currentPrice
	}

	newUpperStopPrice := indicator.calculateStopPrice(
		currentPrice,
		indicator.config.TrailingTopPercentage,
	)

	if newUpperStopPrice < indicator.upperStopPrice {
		candle := indicator.buffer.GetLastCandle()
		msg := fmt.Sprintf("trailing_MOVED: %s\nStopPrice: %f", candle.CloseTime, newUpperStopPrice)
		if IS_REAL_ENABLED {
			LogAndPrintAndSendTg(msg)
		} else {
			LogAndPrint(msg)
		}
		indicator.upperStopPrice = newUpperStopPrice
	}

	indicator.resolveSignal(currentPrice)
	indicator.lastPrice = currentPrice
}

func (indicator *BackTrailingBuyIndicator) Finish() {
	candle := indicator.buffer.GetLastCandle()

	msg := fmt.Sprintf("trailing_FINISHED: %s\nExchangeRate: %f", candle.CloseTime, candle.ClosePrice)
	if IS_REAL_ENABLED {
		LogAndPrintAndSendTg(msg)
	} else {
		LogAndPrint(msg)
	}

	indicator.isStarted = false
	indicator.updatesCount = 0
	indicator.updatedTimesBeforeFinish = 0
}

func (indicator *BackTrailingBuyIndicator) HasSignal() bool {
	return indicator.hasSignal
}

func (indicator *BackTrailingBuyIndicator) isGrowing() bool {
	currentPrice := indicator.buffer.GetLastCandleClosePrice()

	return currentPrice > indicator.lastPrice
}

func (indicator *BackTrailingBuyIndicator) resolveSignal(currentPrice float64) {
	indicator.hasSignal = currentPrice >= indicator.upperStopPrice
	if indicator.hasSignal || indicator.updatedTimesBeforeFinish > 0 {
		indicator.hasSignal = true
		if indicator.updatedTimesBeforeFinish == indicator.config.TrailingUpdateTimesBeforeFinish {
			indicator.Finish()
			return
		}
		indicator.updatedTimesBeforeFinish++
	}
}

func (indicator *BackTrailingBuyIndicator) calculateStopPrice(closePrice, percentage float64) float64 {
	return closePrice + ((closePrice * percentage) / 100)
}

// --------------------------------

type BuysCountIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewBuysCountIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) BuysCountIndicator {
	return BuysCountIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *BuysCountIndicator) HasSignal() bool {
	val := indicator.db.CountUnsoldBuys()
	return UNSOLD_BUYS_COUNT > val
}

func (indicator *BuysCountIndicator) IsStarted() bool {
	return true
}

func (indicator *BuysCountIndicator) Start() {
}

func (indicator *BuysCountIndicator) Update() {
}

func (indicator *BuysCountIndicator) Finish() {
}

// --------------------------------

type WaitForPeriodIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewWaitForPeriodIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) WaitForPeriodIndicator {
	return WaitForPeriodIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *WaitForPeriodIndicator) HasSignal() bool {
	candle := indicator.buffer.GetLastCandle()
	return indicator.db.CanBuyInGivenPeriod(candle.CloseTime, indicator.config.WaitAfterLastBuyPeriod)
}

func (indicator *WaitForPeriodIndicator) IsStarted() bool {
	return true
}

func (indicator *WaitForPeriodIndicator) Start() {
}

func (indicator *WaitForPeriodIndicator) Update() {
}

func (indicator *WaitForPeriodIndicator) Finish() {
}

// ---------------------------------------

type BigFallIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewBigFallIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) BigFallIndicator {
	return BigFallIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *BigFallIndicator) HasSignal() bool {
	count := len(indicator.buffer.GetCandles())
	if (indicator.config.BigFallCandlesCount + 1) > count {
		return false
	}

	firstCandle := indicator.buffer.GetBackCandle(indicator.config.BigFallCandlesCount).ClosePrice
	lastCandle := indicator.buffer.GetLastCandle().ClosePrice
	fallPercentage := -1 * CalcGrowth(firstCandle, lastCandle)

	return fallPercentage >= indicator.config.BigFallPercentage
}

func (indicator *BigFallIndicator) IsStarted() bool {
	return true
}

func (indicator *BigFallIndicator) Start() {
}

func (indicator *BigFallIndicator) Update() {
}

func (indicator *BigFallIndicator) Finish() {
}

// ---------------------------------------

type GradientDescentIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewGradientDescentIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) GradientDescentIndicator {
	return GradientDescentIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *GradientDescentIndicator) HasSignal() bool {
	count := len(indicator.buffer.GetCandles())
	if (indicator.config.GradientDescentCandles + 1) > count {
		return false
	}

	closePrices := GetClosePrices(indicator.buffer.GetCandles())
	smoothedPrices := FilterZeroPrices(talib.Sma(closePrices, indicator.config.GradientDescentPeriod))
	smoothedLen := len(smoothedPrices)
	if 4 > smoothedLen {
		return false
	}

	x := float64(indicator.config.GradientDescentCandles)
	y := smoothedPrices[0] - smoothedPrices[smoothedLen-1]
	gradient := y / x

	return -indicator.config.GradientDescentGradient <= gradient &&
		gradient <= indicator.config.GradientDescentGradient
}

func (indicator *GradientDescentIndicator) IsStarted() bool {
	return true
}

func (indicator *GradientDescentIndicator) Start() {
}

func (indicator *GradientDescentIndicator) Update() {
}

func (indicator *GradientDescentIndicator) Finish() {
}
