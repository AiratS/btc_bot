package main

import (
	"github.com/adshao/go-binance/v2/futures"
)

func RunFuturesRealTime() {
	client := futures.NewClient(BINANCE_API_KEY, BINANCE_SECRET_KEY)
	futuresClient := NewFuturesOrderManager(client)

	futuresClient.IsBuySold(CANDLE_SYMBOL, 32332357883)

	//futuresClient.CreateFuturesMarketByOrder(CANDLE_SYMBOL, 28370)
	//futuresClient.CreateSellOrder(CANDLE_SYMBOL, 28400, 0.001)
	//futuresClient.HasEnoughMoneyForBuy()
}
