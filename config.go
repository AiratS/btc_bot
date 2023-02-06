package main

const CANDLE_SYMBOL = "BTCUSDT"
const CANDLE_INTERVAL = "1m"
const TOTAL_MONEY_AMOUNT = 100
const DATASETS_DIRECTORY = "datasets"

type Config struct {
	HighSellPercentage float64
	LowSellPercentage  float64

	PriceFallCandles    int
	PriceFallPercentage float64

	TrailingTopPercentage float64
	TrailingLowPercentage float64
}
