package main

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"strconv"
)

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
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  float64
	TakerBuyQuoteAssetVolume float64
	Ignore                   int64
	IsClosed                 bool
}

func (candle *Candle) GetPrice() float64 {
	return candle.ClosePrice
}

// Candle source converter
func WebSocketCandleToKlineCandle(wsKline binance.WsKline) Candle {
	openPrice, openPriceErr := strconv.ParseFloat(wsKline.Open, 64)
	closePrice, closePriceErr := strconv.ParseFloat(wsKline.Close, 64)
	highPrice, highPriceErr := strconv.ParseFloat(wsKline.High, 64)
	lowPrice, lowPriceErr := strconv.ParseFloat(wsKline.Low, 64)
	volume, volumeErr := strconv.ParseFloat(wsKline.Volume, 64)

	if openPriceErr != nil ||
		closePriceErr != nil ||
		highPriceErr != nil ||
		lowPriceErr != nil ||
		volumeErr != nil {

		panic("Can not convert Websocket candle.")
	}

	return Candle{
		Symbol:     wsKline.Symbol,
		OpenTime:   FormatTimestamp(wsKline.StartTime),
		CloseTime:  FormatTimestamp(wsKline.EndTime),
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		ClosePrice: closePrice,
		Volume:     volume,
		//QuoteAssetVolume:         wsKline.QuoteVolume,
		NumberOfTrades: wsKline.TradeNum,
		//TakerBuyBaseAssetVolume:  wsKline.ActiveBuyVolume,
		//TakerBuyQuoteAssetVolume: wsKline.ActiveBuyQuoteVolume,
		IsClosed: wsKline.IsFinal,
	}
}

func WebSocketCandleToKlineCandleFutures(wsKline futures.WsKline) Candle {
	openPrice, openPriceErr := strconv.ParseFloat(wsKline.Open, 64)
	closePrice, closePriceErr := strconv.ParseFloat(wsKline.Close, 64)
	highPrice, highPriceErr := strconv.ParseFloat(wsKline.High, 64)
	lowPrice, lowPriceErr := strconv.ParseFloat(wsKline.Low, 64)
	volume, volumeErr := strconv.ParseFloat(wsKline.Volume, 64)

	if openPriceErr != nil ||
		closePriceErr != nil ||
		highPriceErr != nil ||
		lowPriceErr != nil ||
		volumeErr != nil {

		panic("Can not convert Websocket candle.")
	}

	return Candle{
		Symbol:     wsKline.Symbol,
		OpenTime:   FormatTimestamp(wsKline.StartTime),
		CloseTime:  FormatTimestamp(wsKline.EndTime),
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		ClosePrice: closePrice,
		Volume:     volume,
		//QuoteAssetVolume:         wsKline.QuoteVolume,
		NumberOfTrades: wsKline.TradeNum,
		//TakerBuyBaseAssetVolume:  wsKline.ActiveBuyVolume,
		//TakerBuyQuoteAssetVolume: wsKline.ActiveBuyQuoteVolume,
		IsClosed: wsKline.IsFinal,
	}
}

type SecToMinCandleIntervalConverter struct {
}

func NewSecToMinCandleConverter() SecToMinCandleIntervalConverter {
	return SecToMinCandleIntervalConverter{}
}

func (converter *SecToMinCandleIntervalConverter) Convert(candle Candle) (Candle, bool) {
	if !candle.IsClosed {
		return Candle{}, false
	}

	return candle, true
}
