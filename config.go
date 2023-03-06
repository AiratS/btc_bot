package main

const CANDLE_SYMBOL = "BTCUSDT"
const CANDLE_INTERVAL = "1m"
const TOTAL_MONEY_AMOUNT = 100
const COMMISSION = 0.15
const DATASETS_DIRECTORY = "datasets"
const UNSOLD_BUYS_COUNT = 10

type Config struct {
	HighSellPercentage float64

	TrailingTopPercentage           float64
	TrailingUpdateTimesBeforeFinish int

	WaitAfterLastBuyPeriod int

	BigFallCandlesCount int
	BigFallPercentage   float64

	DesiredPriceCandles int

	TotalRevenue float64
	Selection    float64
}
