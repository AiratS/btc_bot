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

	Log(fmt.Sprintf("BackTrailingBuyIndicator__STARTED: %s\nExchangeRate: %f", candle.CloseTime, candle.ClosePrice))

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
		Log(fmt.Sprintf("BackTrailingBuyIndicator__STOP_MOVED: %s\nStopPrice: %f", candle.CloseTime, newUpperStopPrice))

		indicator.upperStopPrice = newUpperStopPrice
	}

	indicator.resolveSignal(currentPrice)
	indicator.lastPrice = currentPrice
}

func (indicator *BackTrailingBuyIndicator) Finish() {
	candle := indicator.buffer.GetLastCandle()
	Log(fmt.Sprintf("BackTrailingBuyIndicator__FINISHED: %s\nExchangeRate: %f", candle.CloseTime, candle.ClosePrice))

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

	//firstCandle := indicator.buffer.GetBackCandle(indicator.config.BigFallCandlesCount).ClosePrice
	//lastCandle := indicator.buffer.GetLastCandle().ClosePrice
	//fallPercentage := -1 * CalcGrowth(firstCandle, lastCandle)

	// ----------------------
	closePrices := GetValuesFromSlice(
		GetClosePrices(indicator.buffer.GetCandles()),
		indicator.config.BigFallCandlesCount,
	)
	smoothedPrices := FilterZeroPrices(talib.Sma(closePrices, indicator.getPeriod()))
	smoothedLen := len(smoothedPrices)

	//Log(fmt.Sprintf(
	//	"BigFallIndicator__smoothedPricesCount: %d",
	//	len(smoothedPrices),
	//))

	if 4 > smoothedLen {
		return false
	}

	firstPrice := smoothedPrices[0]
	lastPrice := smoothedPrices[smoothedLen-1]
	fallPercentage := -1 * CalcGrowth(firstPrice, lastPrice)

	//Log(fmt.Sprintf(
	//	"BigFallIndicator__fallPercentage: %f",
	//	CalcGrowth(firstPrice, lastPrice),
	//))
	// ----------------------

	return fallPercentage >= indicator.config.BigFallPercentage
}

func (indicator *BigFallIndicator) getPeriod() int {
	period := indicator.config.BigFallCandlesCount - indicator.config.BigFallSmoothPeriod
	if period <= 0 {
		period = 1
	}

	return period
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

	closePrices := GetValuesFromSlice(
		GetClosePrices(indicator.buffer.GetCandles()),
		indicator.config.GradientDescentCandles,
	)
	smoothedPrices := FilterZeroPrices(talib.Sma(closePrices, indicator.getPeriod()))
	smoothedLen := len(smoothedPrices)
	if 4 > smoothedLen {
		return false
	}

	//// Проверяем, что весь период падали
	//fallingPercentage := indicator.calcFallingPercentage(smoothedPrices)
	//if fallingPercentage < 80 {
	//	return false
	//}
	//
	////  Проверяем, что мы находимся на дне падающего участка
	//x := float64(indicator.config.GradientDescentCandles)
	//y := smoothedPrices[smoothedLen-2] - smoothedPrices[smoothedLen-1]
	//currentGradient := y / x
	//
	//return -indicator.config.GradientDescentGradient <= currentGradient &&
	//	currentGradient <= indicator.config.GradientDescentGradient

	x := float64(indicator.config.GradientDescentCandles)
	y := smoothedPrices[0] - smoothedPrices[smoothedLen-1]
	gradient := y / x

	return -indicator.config.GradientDescentGradient <= gradient &&
		gradient <= indicator.config.GradientDescentGradient
}

func (indicator *GradientDescentIndicator) getPeriod() int {
	period := indicator.config.GradientDescentCandles - indicator.config.GradientDescentPeriod
	if period <= 0 {
		period = 1
	}

	return period
}

func (indicator *GradientDescentIndicator) mapGradients(smoothedPrices []float64) []float64 {
	var gradients []float64

	for index, price := range smoothedPrices {
		if index == 0 {
			continue
		}
		gradients = append(gradients, smoothedPrices[index-1]-price)
	}

	return gradients
}

func (indicator *GradientDescentIndicator) calcFallingPercentage(smoothedPrices []float64) float64 {
	fallingCount := 0
	gradients := indicator.mapGradients(smoothedPrices)

	for _, gradient := range gradients {
		if gradient > 0 {
			fallingCount++
		}
	}

	total := len(smoothedPrices)

	return (float64(fallingCount) * 100.0) / float64(total)
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

// ---------------------------------------

type LessThanPreviousBuyIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewLessThanPreviousBuyIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) LessThanPreviousBuyIndicator {
	return LessThanPreviousBuyIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *LessThanPreviousBuyIndicator) HasSignal() bool {
	hasValue, buy := indicator.db.GetLastUnsoldBuy()
	if !hasValue {
		return true
	}

	percentage := CalcGrowth(buy.ExchangeRate, indicator.buffer.GetLastCandleClosePrice())

	return percentage <= indicator.config.LessThanPreviousBuyPercentage
}

func (indicator *LessThanPreviousBuyIndicator) IsStarted() bool {
	return true
}

func (indicator *LessThanPreviousBuyIndicator) Start() {
}

func (indicator *LessThanPreviousBuyIndicator) Update() {
}

func (indicator *LessThanPreviousBuyIndicator) Finish() {
}

// ---------------------------------------

type StopAfterUnsuccessfullySellIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewStopAfterUnsuccessfullySellIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) StopAfterUnsuccessfullySellIndicator {
	return StopAfterUnsuccessfullySellIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *StopAfterUnsuccessfullySellIndicator) HasSignal() bool {
	hasSell, sell := indicator.db.FindLastLiquidationSell()
	if !hasSell {
		return true
	}

	currentCandle := indicator.buffer.GetLastCandle()
	currentTime := ConvertDateStringToTime(currentCandle.CloseTime)
	sellTime := ConvertDatabaseDateStringToTime(sell.CreatedAt)
	diff := currentTime.Sub(sellTime)

	return float64(indicator.config.StopAfterUnsuccessfullySellMinutes) <= diff.Minutes()
}

func (indicator *StopAfterUnsuccessfullySellIndicator) IsStarted() bool {
	return true
}

func (indicator *StopAfterUnsuccessfullySellIndicator) Start() {
}

func (indicator *StopAfterUnsuccessfullySellIndicator) Update() {
}

func (indicator *StopAfterUnsuccessfullySellIndicator) Finish() {
}
