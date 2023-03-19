package main

const CANDLE_SYMBOL = "BTCUSDT"
const CANDLE_INTERVAL = "1m"
const BALANCE_MONEY = 1000.0
const COMMISSION = 0.15
const DATASETS_DIRECTORY = "datasets"
const UNSOLD_BUYS_COUNT = 20

type Config struct {
	HighSellPercentage float64

	TrailingTopPercentage           float64
	TrailingUpdateTimesBeforeFinish int

	WaitAfterLastBuyPeriod int

	BigFallCandlesCount int
	BigFallPercentage   float64

	DesiredPriceCandles int

	GradientDescentCandles  int
	GradientDescentPeriod   int
	GradientDescentGradient float64

	TrailingSellActivationAdditionPercentage float64
	TrailingSellStopPercentage               float64

	TotalMoneyAmount float64

	TotalRevenue float64
	Selection    float64
}
