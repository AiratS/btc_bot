package main

const OPEN_TIME = 0
const OPEN_PRICE = 1
const HIGH_PRICE = 2
const LOW_PRICE = 3
const CLOSE_PRICE = 4
const VOLUME = 5
const CLOSE_TIME = 6
const QUOTE_ASSET_VOLUME = 7
const NUMBER_OF_TRADES = 8
const TAKER_BUY_BASE_ASSET_VOLUME = 9
const TAKER_BUY_QUOTE_ASSET_VOLUME = 10
const IGNORE = 11

type Candle struct {
	Symbol                   string
	OpenTime                 string
	CloseTime                string
	OpenPrice                float64
	HighPrice                float64
	LowPrice                 float64
	ClosePrice               float64
	Volume                   float64
	QuoteAssetVolume         float64
	NumberOfTrades           int
	TakerBuyBaseAssetVolume  float64
	TakerBuyQuoteAssetVolume float64
	Ignore                   int
}

func (candle *Candle) GetPrice() float64 {
	return candle.ClosePrice
}
