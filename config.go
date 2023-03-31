package main

const CANDLE_SYMBOL = "BTCBUSD"
const CANDLE_INTERVAL = "30m"
const BALANCE_MONEY = 1000.0
const COMMISSION = 0.06
const DATASETS_DIRECTORY = "datasets"
const UNSOLD_BUYS_COUNT = 20

type Config struct {
	HighSellPercentage float64

	TrailingTopPercentage           float64
	TrailingUpdateTimesBeforeFinish int

	WaitAfterLastBuyPeriod int

	BigFallCandlesCount int
	BigFallSmoothPeriod int
	BigFallPercentage   float64

	DesiredPriceCandles int

	GradientDescentCandles  int
	GradientDescentPeriod   int
	GradientDescentGradient float64

	TrailingSellActivationAdditionPercentage float64
	TrailingSellStopPercentage               float64

	TotalMoneyAmount float64
	Leverage         int

	TotalRevenue   float64
	TotalBuysCount int
	AvgSellTime    float64
	Selection      float64
}
