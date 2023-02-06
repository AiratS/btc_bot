package main

const BTC_SYMBOL = "BTCUSDT"

type Config struct {
	HighSellPercentage float64
	LowSellPercentage  float64

	PriceFallCandles    int
	PriceFallPercentage float64
}
