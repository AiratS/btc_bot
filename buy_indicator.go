package main

import (
	"fmt"
	"github.com/markcheno/go-talib"
	"math"
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

	if indicator.db.CountUnsoldBuys() > 0 {
		return true
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

	if IS_REAL_ENABLED {
		Log(fmt.Sprintf(
			"BigFallIndicator__smoothedPricesCount: %d",
			len(smoothedPrices),
		))
	}

	if 4 > smoothedLen {
		return false
	}

	firstPrice := smoothedPrices[0]
	lastPrice := smoothedPrices[smoothedLen-1]
	fallPercentage := -1 * CalcGrowth(firstPrice, lastPrice)

	if IS_REAL_ENABLED {
		Log(fmt.Sprintf(
			"BigFallIndicator__fallPercentage: %f",
			CalcGrowth(firstPrice, lastPrice),
		))
	}
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

	if indicator.db.CountUnsoldBuys() > 0 {
		return true
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

type MoreThanPreviousBuyIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewMoreThanPreviousBuyIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) MoreThanPreviousBuyIndicator {
	return MoreThanPreviousBuyIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *MoreThanPreviousBuyIndicator) HasSignal() bool {
	hasValue, buy := indicator.db.GetLastUnsoldBuy()
	if !hasValue {
		return true
	}

	percentage := CalcGrowth(buy.ExchangeRate, indicator.buffer.GetLastCandleClosePrice())

	return percentage >= indicator.config.MoreThanPreviousBuyPercentage
}

func (indicator *MoreThanPreviousBuyIndicator) IsStarted() bool {
	return true
}

func (indicator *MoreThanPreviousBuyIndicator) Start() {
}

func (indicator *MoreThanPreviousBuyIndicator) Update() {
}

func (indicator *MoreThanPreviousBuyIndicator) Finish() {
}

// ---------------------------------------

// ---------------------------------------

type BoostBuyIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewBoostBuyIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) BoostBuyIndicator {
	return BoostBuyIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *BoostBuyIndicator) HasSignal() bool {
	hasFirstBuy, firstBuy := indicator.db.FindFirstUnsoldBuy()
	if !hasFirstBuy {
		return false
	}

	currentCandle := indicator.buffer.GetLastCandle()

	// Check for percentage
	fallPercentage := -1 * CalcGrowth(firstBuy.ExchangeRate, currentCandle.GetPrice())
	if indicator.config.BoostBuyFallPercentage > fallPercentage {
		return false
	}

	// Check for period
	hasLastBuy, lastBuy := indicator.db.GetLastUnsoldBuy()
	if !hasLastBuy {
		return false
	}

	currentTime := ConvertDateStringToTime(currentCandle.CloseTime)
	buyTime := ConvertDatabaseDateStringToTime(lastBuy.CreatedAt)
	diff := currentTime.Sub(buyTime)

	return float64(indicator.config.BoostBuyPeriodMinutes) <= diff.Minutes()
}

func (indicator *BoostBuyIndicator) IsStarted() bool {
	return true
}

func (indicator *BoostBuyIndicator) Start() {
}

func (indicator *BoostBuyIndicator) Update() {
}

func (indicator *BoostBuyIndicator) Finish() {
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

// -----------------------------------------------------

const (
	LinearRegressionLimit = 30
	LinearRegressionKStep = 0.1
)

type LinearRegressionIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewLinearRegressionIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) LinearRegressionIndicator {
	return LinearRegressionIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *LinearRegressionIndicator) HasSignal() bool {
	count := len(indicator.buffer.GetCandles())
	if (indicator.config.LinearRegressionCandles + 1) > count {
		return false
	}

	if indicator.db.CountUnsoldBuys() > 0 {
		return true
	}

	closePrices := GetValuesFromSlice(
		GetClosePrices(indicator.buffer.GetCandles()),
		indicator.config.LinearRegressionCandles,
	)

	lastIdx := len(closePrices) - 1
	if closePrices[lastIdx-1] >= closePrices[lastIdx] {
		return false
	}

	smoothedPrices := FilterZeroPrices(talib.Sma(closePrices, indicator.getPeriod()))
	smoothedLen := len(smoothedPrices)
	if 4 > smoothedLen {
		return false
	}

	// Calc K coefficient
	mse, k := findBestLineCoefficient(smoothedPrices)
	if mse > indicator.config.LinearRegressionMse {
		return false
	}

	return indicator.config.LinearRegressionK >= k
}

func findBestLineCoefficient(priceValues []float64) (mse, kResulting float64) {
	mse = calcLineMse(
		getLinePoints(-LinearRegressionLimit, priceValues[0], len(priceValues)),
		priceValues,
	)

	k := -LinearRegressionLimit + LinearRegressionKStep
	kResulting = k

	for k < LinearRegressionLimit {
		newMse := calcLineMse(
			getLinePoints(k, priceValues[0], len(priceValues)),
			priceValues,
		)

		if mse > newMse {
			mse = newMse
			kResulting = k
		}

		k += LinearRegressionKStep
	}

	return mse, kResulting
}

func calcLineMse(lineValues, priceValues []float64) float64 {
	var errors []float64

	for i := range lineValues {
		diff := priceValues[i] - lineValues[i]
		errors = append(errors, math.Pow(diff, 2))
	}

	return Sum(errors) / float64(len(lineValues))
}

func getLinePoints(k, b float64, n int) []float64 {
	var points []float64

	for x := 0; x < n; x++ {
		fx := k*float64(x) + b
		points = append(points, fx)
	}

	return points
}

func (indicator *LinearRegressionIndicator) IsStarted() bool {
	return true
}

func (indicator *LinearRegressionIndicator) Start() {
}

func (indicator *LinearRegressionIndicator) Update() {
}

func (indicator *LinearRegressionIndicator) Finish() {
}

func (indicator *LinearRegressionIndicator) getPeriod() int {
	period := indicator.config.LinearRegressionCandles - indicator.config.LinearRegressionPeriod
	if period <= 0 {
		period = 1
	}

	return period
}
