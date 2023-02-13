package main

const CANDLE_SYMBOL = "BTCUSDT"
const CANDLE_INTERVAL = "1m"
const TOTAL_MONEY_AMOUNT = 100
const COMMISSION = 0.15
const DATASETS_DIRECTORY = "datasets"

type Config struct {
	HighSellPercentage float64

	TrailingTopPercentage           float64
	TrailingUpdateTimesBeforeFinish int

	TotalRevenue float64
	Selection    float64
}

type ConfigRestriction struct {
	HighSellPercentage MinMaxFloat64

	TrailingTopPercentage           MinMaxFloat64
	TrailingUpdateTimesBeforeFinish MinMaxInt
}

type MinMaxInt struct {
	min int
	max int
}

type MinMaxFloat64 struct {
	min float64
	max float64
}
