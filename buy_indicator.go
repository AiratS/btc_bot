package main

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"math"
	"time"

	"github.com/markcheno/go-talib"
	"github.com/montanaflynn/stats"
)

type BuyIndicator interface {
	HasSignal(candle Candle) bool
	IsStarted() bool
	Start()
	Update()
	Finish()
}

type StableMarketIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewStableMarketIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) StableMarketIndicator {
	return StableMarketIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *StableMarketIndicator) HasSignal(candle Candle) bool {
	count := len(indicator.buffer.GetCandles())
	if (indicator.config.StableTradeIndicatorCandles + 1) > count {
		return false
	}

	if indicator.db.CountUnsoldBuys() == 0 {
		return true
	}

	// Начинаем работу только ниже определенного процента
	if !indicator.shouldStart() {
		return false
	}

	// Если цена слишком сильно упала, но не поймали стабилизацию
	if indicator.hasGuaranteedSignal() {
		return true
	}

	closePrices := GetValuesFromSlice(
		GetClosePrices(indicator.buffer.GetCandles()),
		indicator.config.StableTradeIndicatorCandles,
	)
	smoothedPrices := FilterZeroPrices(talib.Sma(closePrices, indicator.getPeriod()))
	smoothedLen := len(smoothedPrices)
	if 4 > smoothedLen {
		return false
	}

	topPrice := CalcUpperPrice(smoothedPrices[0], indicator.config.StableTradeIndicatorPercentage)
	bottomPrice := CalcBottomPrice(smoothedPrices[0], indicator.config.StableTradeIndicatorPercentage)

	lastPrice := smoothedPrices[len(smoothedPrices)-1]

	return bottomPrice <= lastPrice && lastPrice <= topPrice
}

func (indicator *StableMarketIndicator) shouldStart() bool {
	unsoldBuys := indicator.db.FetchUnsoldBuys()
	if len(unsoldBuys) == 0 {
		return false
	}

	avgPrice := CalcFuturesAvgPrice(unsoldBuys)
	lastRealPrice := indicator.buffer.GetLastCandleClosePrice()
	fallPercentage := -1 * CalcGrowth(avgPrice, lastRealPrice)

	return fallPercentage >= indicator.config.StableTradeMinStartPercentage
}

func (indicator *StableMarketIndicator) hasGuaranteedSignal() bool {
	unsoldBuys := indicator.db.FetchUnsoldBuys()
	if len(unsoldBuys) == 0 {
		return false
	}

	avgPrice := CalcFuturesAvgPrice(unsoldBuys)
	lastRealPrice := indicator.buffer.GetLastCandleClosePrice()
	fallPercentage := -1 * CalcGrowth(avgPrice, lastRealPrice)

	return fallPercentage >= indicator.config.StableTradeGuaranteedSignalPercentage
}

func (indicator *StableMarketIndicator) getPeriod() int {
	period := indicator.config.StableTradeIndicatorCandles - indicator.config.StableTradeIndicatorSmoothPeriod
	if period <= 0 {
		period = 1
	}

	return period
}

func (indicator *StableMarketIndicator) IsStarted() bool {
	return false
}

func (indicator *StableMarketIndicator) Start() {
}

func (indicator *StableMarketIndicator) Update() {
}

func (indicator *StableMarketIndicator) Finish() {
}

// BackTrailingBuyIndicator
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

func (indicator *BackTrailingBuyIndicator) HasSignal(candle Candle) bool {
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

func (indicator *BuysCountIndicator) HasSignal(candle Candle) bool {
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

func (indicator *WaitForPeriodIndicator) HasSignal(candle Candle) bool {
	candle1 := indicator.buffer.GetLastCandle()
	return indicator.db.CanBuyInGivenPeriod(candle1.CloseTime, indicator.config.WaitAfterLastBuyPeriod)
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

func (indicator *BigFallIndicator) HasSignal(candle Candle) bool {
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
	closePrices = append(closePrices, candle.GetPrice())
	//smoothedPrices := FilterZeroPrices(talib.Sma(closePrices, indicator.getPeriod()))
	smoothedPrices := closePrices
	smoothedLen := len(smoothedPrices)

	if IS_REAL_ENABLED {
		// Log(fmt.Sprintf(
		// 	"BigFallIndicator__smoothedPricesCount: %d",
		// 	len(smoothedPrices),
		// ))
	}

	//if 4 > smoothedLen {
	//	return false
	//}

	firstPrice := smoothedPrices[0]
	lastPrice := smoothedPrices[smoothedLen-1]
	fallPercentage := -1 * CalcGrowth(firstPrice, lastPrice)

	if IS_REAL_ENABLED {
		// Log(fmt.Sprintf(
		// 	"BigFallIndicator__fallPercentage: %f",
		// 	CalcGrowth(firstPrice, lastPrice),
		// ))
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

func (indicator *GradientDescentIndicator) HasSignal(candle Candle) bool {
	count := len(indicator.buffer.GetCandles())
	if (indicator.config.GradientDescentCandles + 1) > count {
		//Log(fmt.Sprintf("GradientDescentIndicator: not enough candles"))
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
		//Log(fmt.Sprintf(
		//	"GradientDescentIndicator: not enough smothed candles total: %d, smoothed: %d",
		//	count,
		//	smoothedLen,
		//))
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

func (indicator *LessThanPreviousBuyIndicator) HasSignal(candle Candle) bool {
	hasValue, buy := indicator.db.GetLastUnsoldBuy()
	if !hasValue {
		return true
	}

	percentage := CalcGrowth(buy.ExchangeRate, candle.GetPrice())

	return indicator.config.LessThanPreviousBuyPercentage >= percentage
}

func (indicator *LessThanPreviousBuyIndicator) getLessThanPercentage() float64 {
	if !ENABLE_DYNAMIC_NEXT_BUY_PERCENTAGE {
		return indicator.config.LessThanPreviousBuyPercentage
	}

	unsoldCount := indicator.db.CountUnsoldBuys()
	previousPercentage := math.Abs(indicator.config.LessThanPreviousBuyPercentage)

	for i := 0; i < unsoldCount; i++ {
		increasePercentage := math.Pow(float64(i), 2) / indicator.config.ParabolaDivider
		previousPercentage = CalcUpperPrice(previousPercentage, increasePercentage)
	}

	return -1 * previousPercentage
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

type LessThanPreviousAverageIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewLessThanPreviousAverageIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) LessThanPreviousAverageIndicator {
	return LessThanPreviousAverageIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *LessThanPreviousAverageIndicator) HasSignal(candle Candle) bool {
	unsoldBuys := indicator.db.FetchUnsoldBuys()
	if 0 == len(unsoldBuys) {
		return true
	}

	avgFuturesPrice := CalcFuturesAvgPrice(unsoldBuys)
	percentage := CalcGrowth(avgFuturesPrice, indicator.buffer.GetLastCandleClosePrice())

	return indicator.config.LessThanPreviousBuyPercentage >= percentage
}

func (indicator *LessThanPreviousAverageIndicator) IsStarted() bool {
	return true
}

func (indicator *LessThanPreviousAverageIndicator) Start() {
}

func (indicator *LessThanPreviousAverageIndicator) Update() {
}

func (indicator *LessThanPreviousAverageIndicator) Finish() {
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

func (indicator *MoreThanPreviousBuyIndicator) HasSignal(candle Candle) bool {
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

type MoreThanPreviousAverageIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewMoreThanPreviousAverageIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) MoreThanPreviousAverageIndicator {
	return MoreThanPreviousAverageIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *MoreThanPreviousAverageIndicator) HasSignal(candle Candle) bool {
	unsoldBuys := indicator.db.FetchUnsoldBuys()
	if 0 == len(unsoldBuys) {
		return true
	}

	avgFuturesPrice := CalcFuturesAvgPrice(unsoldBuys)
	percentage := CalcGrowth(avgFuturesPrice, indicator.buffer.GetLastCandleClosePrice())

	return indicator.config.MoreThanPreviousBuyPercentage <= percentage
}

func (indicator *MoreThanPreviousAverageIndicator) IsStarted() bool {
	return true
}

func (indicator *MoreThanPreviousAverageIndicator) Start() {
}

func (indicator *MoreThanPreviousAverageIndicator) Update() {
}

func (indicator *MoreThanPreviousAverageIndicator) Finish() {
}

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

func (indicator *BoostBuyIndicator) HasSignal(candle Candle) bool {
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

func (indicator *StopAfterUnsuccessfullySellIndicator) HasSignal(candle Candle) bool {
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

func (indicator *LinearRegressionIndicator) HasSignal(candle Candle) bool {
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
	_, k := findBestLineCoefficient(smoothedPrices)
	//if mse > indicator.config.LinearRegressionMse {
	//	return false
	//}

	deviation, err := stats.StandardDeviation(stats.LoadRawData(smoothedPrices))
	if err != nil {
		panic(err)
	}

	if deviation > indicator.config.LinearRegressionDeviation {
		return false
	}

	return indicator.config.LinearRegressionK >= math.Abs(k)
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

// ---------------------------------------

type GradientSwingIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewGradientSwingIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) GradientSwingIndicator {
	return GradientSwingIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

const SwingTypeGrowth = 0
const SwingTypeFall = 1
const SwingTypeAny = 2

func (indicator *GradientSwingIndicator) HasSignal(candle Candle) bool {
	count := len(indicator.buffer.GetCandles())
	if (indicator.config.GradientSwingIndicatorCandles + 1) > count {
		return false
	}

	if indicator.db.CountUnsoldBuys() > 0 {
		return true
	}

	closePrices := GetValuesFromSlice(
		GetClosePrices(indicator.buffer.GetCandles()),
		indicator.config.GradientSwingIndicatorCandles,
	)
	smoothedPrices := FilterZeroPrices(talib.Sma(closePrices, indicator.getPeriod()))
	smoothedLen := len(smoothedPrices)
	if 4 > smoothedLen {
		return false
	}

	// Main part
	lastIdx := smoothedLen - 1
	penultPrice := smoothedPrices[lastIdx-2]
	prevPrice := smoothedPrices[lastIdx-1]
	currentCandle := smoothedPrices[lastIdx]

	gradient1 := prevPrice - penultPrice
	gradient2 := currentCandle - prevPrice

	// Checks
	isGradient1NegativeOrZero := 0 >= gradient1
	isGradient2Positive := 0 < gradient2

	isGradient1PlusOrZero := 0 <= gradient1
	isGradient2Negative := 0 > gradient2

	switch indicator.config.GradientSwingIndicatorSwingType {
	case SwingTypeGrowth:
		return isGradient1NegativeOrZero && isGradient2Positive
	case SwingTypeFall:
		return isGradient1PlusOrZero && isGradient2Negative
	case SwingTypeAny:
		return isGradient1NegativeOrZero && isGradient2Positive ||
			isGradient1PlusOrZero && isGradient2Negative
	}

	return false
}

func (indicator *GradientSwingIndicator) getPeriod() int {
	period := indicator.config.GradientSwingIndicatorCandles - indicator.config.GradientSwingIndicatorPeriod
	if period <= 0 {
		period = 1
	}

	return period
}

func (indicator *GradientSwingIndicator) IsStarted() bool {
	return true
}

func (indicator *GradientSwingIndicator) Start() {
}

func (indicator *GradientSwingIndicator) Update() {
}

func (indicator *GradientSwingIndicator) Finish() {
}

// ------------------------------------------------------

type WindowLongIndicator struct {
	config         *Config
	buffer         *Buffer
	db             *Database
	currentWindow  int
	lastWindowTime time.Time
}

func NewWindowLongIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) WindowLongIndicator {
	return WindowLongIndicator{
		config:        config,
		buffer:        buffer,
		db:            db,
		currentWindow: config.WindowWindowsCount,
	}
}

func (indicator *WindowLongIndicator) HasSignal(candle Candle) bool {
	if indicator.checkForPercentage() {
		indicator.ResetWindow()
		return true
	}

	// If the last window and no signal, wait until percentage reached
	if 1 == indicator.currentWindow {
		return false
	}

	// Check for period
	_, buy := indicator.db.GetLastUnsoldBuy()
	if indicator.config.WindowWindowsCount == indicator.currentWindow {
		indicator.lastWindowTime = carbon.Parse(buy.CreatedAt, carbon.Greenwich).ToStdTime()
	}

	currentCandle := indicator.buffer.GetLastCandle()
	currentPeriod := indicator.getCurrentWindowPeriod()
	diffInMinutes := carbon.FromStdTime(indicator.lastWindowTime).
		DiffInMinutes(carbon.Parse(currentCandle.CloseTime, carbon.Greenwich))

	if 0 > diffInMinutes {
		panic("Invalid minutes diff")
	}

	if currentPeriod > diffInMinutes {
		return false
	}

	// Decrease window
	indicator.DecreaseWindow()
	indicator.lastWindowTime = carbon.Parse(currentCandle.CloseTime, carbon.Greenwich).ToStdTime()

	// Check for percentage
	hasSignal := indicator.checkForPercentage()
	if hasSignal {
		indicator.ResetWindow()
	}

	return hasSignal
}

func (indicator *WindowLongIndicator) getCurrentWindowPeriod() int64 {
	return int64(
		indicator.config.WindowBasePeriodMinutes +
			indicator.config.WindowOffsetPeriodMinutes*(indicator.currentWindow-1))
}

func (indicator *WindowLongIndicator) getCurrentWindowPercentage() float64 {
	return indicator.config.WindowBasePercentage +
		indicator.config.WindowOffsetPercentage*(float64(indicator.currentWindow)-1)
}

func (indicator *WindowLongIndicator) checkForPercentage() bool {
	hasValue, buy := indicator.db.GetLastUnsoldBuy()
	if !hasValue {
		return true
	}

	currentCandle := indicator.buffer.GetLastCandle()
	// Check for percentage
	percentage := CalcGrowth(buy.ExchangeRate, currentCandle.GetPrice())
	if 0 <= percentage {
		return false
	}

	currentWindowPercentage := indicator.getCurrentWindowPercentage()

	return currentWindowPercentage <= math.Abs(percentage)
}

func (indicator *WindowLongIndicator) ResetWindow() {
	currentCandle := indicator.buffer.GetLastCandle()
	Log(fmt.Sprintf(
		"WindowLongIndicator__ResetWindow\nCreatedAt: %s\nCurrentWindow: %d\nCurrentPercentage: %f",
		currentCandle.CloseTime,
		indicator.currentWindow,
		indicator.getCurrentWindowPercentage(),
	))

	indicator.currentWindow = indicator.config.WindowWindowsCount
}

func (indicator *WindowLongIndicator) DecreaseWindow() {
	currentCandle := indicator.buffer.GetLastCandle()
	Log(fmt.Sprintf(
		"WindowLongIndicator__DecreaseWindow\nCreatedAt: %s\nCurrentWindow: %d\nCurrentPercentage: %f",
		currentCandle.CloseTime,
		indicator.currentWindow,
		indicator.getCurrentWindowPercentage(),
	))

	indicator.currentWindow--

	if 1 > indicator.currentWindow {
		panic("Invalid currentWindow")
	}
}

func (indicator *WindowLongIndicator) IsStarted() bool {
	return true
}

func (indicator *WindowLongIndicator) Start() {
}

func (indicator *WindowLongIndicator) Update() {
}

func (indicator *WindowLongIndicator) Finish() {
}

// ------------------------------------------------------

type WindowShortIndicator struct {
	config         *Config
	buffer         *Buffer
	db             *Database
	currentWindow  int
	lastWindowTime time.Time
}

func NewWindowShortIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) WindowShortIndicator {
	return WindowShortIndicator{
		config:        config,
		buffer:        buffer,
		db:            db,
		currentWindow: config.WindowWindowsCount,
	}
}

func (indicator *WindowShortIndicator) HasSignal(candle Candle) bool {
	if indicator.checkForPercentage() {
		indicator.ResetWindow()
		return true
	}

	// If the last window and no signal, wait until percentage reached
	if 1 == indicator.currentWindow {
		return false
	}

	// Check for period
	_, buy := indicator.db.GetLastUnsoldBuy()
	if indicator.config.WindowWindowsCount == indicator.currentWindow {
		indicator.lastWindowTime = carbon.Parse(buy.CreatedAt, carbon.Greenwich).ToStdTime()
	}

	currentCandle := indicator.buffer.GetLastCandle()
	currentPeriod := indicator.getCurrentWindowPeriod()
	diffInMinutes := carbon.FromStdTime(indicator.lastWindowTime).
		DiffInMinutes(carbon.Parse(currentCandle.CloseTime, carbon.Greenwich))

	if currentPeriod > diffInMinutes {
		return false
	}

	// Decrease window
	indicator.DecreaseWindow()
	indicator.lastWindowTime = carbon.Parse(currentCandle.CloseTime, carbon.Greenwich).ToStdTime()

	// Check for percentage
	hasSignal := indicator.checkForPercentage()
	if hasSignal {
		indicator.ResetWindow()
	}

	return hasSignal
}

func (indicator *WindowShortIndicator) getCurrentWindowPeriod() int64 {
	return int64(
		indicator.config.WindowBasePeriodMinutes +
			indicator.config.WindowOffsetPeriodMinutes*(indicator.currentWindow-1))
}

func (indicator *WindowShortIndicator) getCurrentWindowPercentage() float64 {
	return indicator.config.WindowBasePercentage +
		indicator.config.WindowOffsetPercentage*(float64(indicator.currentWindow)-1)
}

func (indicator *WindowShortIndicator) checkForPercentage() bool {
	hasValue, buy := indicator.db.GetLastUnsoldBuy()
	if !hasValue {
		return true
	}

	currentCandle := indicator.buffer.GetLastCandle()
	// Check for percentage
	percentage := CalcGrowth(buy.ExchangeRate, currentCandle.GetPrice())
	if 0 > percentage {
		return false
	}

	currentWindowPercentage := indicator.getCurrentWindowPercentage()

	return currentWindowPercentage <= math.Abs(percentage)
}

func (indicator *WindowShortIndicator) ResetWindow() {
	currentCandle := indicator.buffer.GetLastCandle()
	Log(fmt.Sprintf(
		"WindowShortIndicator__ResetWindow\nCreatedAt: %s\nCurrentWindow: %d\nCurrentPercentage: %f",
		currentCandle.CloseTime,
		indicator.currentWindow,
		indicator.getCurrentWindowPercentage(),
	))

	indicator.currentWindow = indicator.config.WindowWindowsCount
}

func (indicator *WindowShortIndicator) DecreaseWindow() {
	currentCandle := indicator.buffer.GetLastCandle()
	Log(fmt.Sprintf(
		"WindowShortIndicator__DecreaseWindow\nCreatedAt: %s\nCurrentWindow: %d\nCurrentPercentage: %f",
		currentCandle.CloseTime,
		indicator.currentWindow,
		indicator.getCurrentWindowPercentage(),
	))

	indicator.currentWindow--

	if 1 > indicator.currentWindow {
		panic("Invalid currentWindow")
	}
}

func (indicator *WindowShortIndicator) IsStarted() bool {
	return true
}

func (indicator *WindowShortIndicator) Start() {
}

func (indicator *WindowShortIndicator) Update() {
}

func (indicator *WindowShortIndicator) Finish() {
}

// ------------------------------------------------------

type CatchingFallingKnifeIndicator struct {
	config         *Config
	buffer         *Buffer
	db             *Database
	currentWindow  int
	lastWindowTime time.Time
}

func NewCatchingFallingKnifeIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) CatchingFallingKnifeIndicator {
	return CatchingFallingKnifeIndicator{
		config:        config,
		buffer:        buffer,
		db:            db,
		currentWindow: config.CatchingFallingKnifeCandles,
	}
}

func (indicator *CatchingFallingKnifeIndicator) HasSignal(candle Candle) bool {
	count := len(indicator.buffer.GetCandles())
	if (indicator.config.CatchingFallingKnifeCandles + 1) > count {
		return false
	}

	if indicator.db.CountUnsoldBuys() > 0 {
		return true
	}

	closePrices := GetValuesFromSlice(
		GetClosePrices(indicator.buffer.GetCandles()),
		indicator.config.CatchingFallingKnifeCandles+1,
	)
	givenPeriodMinPrice := Min(closePrices[:len(closePrices)-1])
	currentPrice := indicator.buffer.GetLastCandle().ClosePrice

	return givenPeriodMinPrice > currentPrice
}

func (indicator *CatchingFallingKnifeIndicator) IsStarted() bool {
	return true
}

func (indicator *CatchingFallingKnifeIndicator) Start() {
}

func (indicator *CatchingFallingKnifeIndicator) Update() {
}

func (indicator *CatchingFallingKnifeIndicator) Finish() {
}
