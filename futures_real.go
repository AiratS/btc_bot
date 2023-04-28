package main

import (
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

const TickerIntervalMinutes = 1

var ticker *Ticker

func RunFuturesRealTime() {
	client := futures.NewClient(BINANCE_API_KEY, BINANCE_SECRET_KEY)
	tgBot, _ = tgbotapi.NewBotAPI(TG_API_KEY)

	candleConverter = NewSecToMinCandleConverter()
	config := GetRealBotConfig()
	realBot = NewFuturesRealBot(&config, client)
	ticker = NewTicker(TickerIntervalMinutes)

	errHandler := func(err error) {
		fmt.Println(err)
	}

	for {
		fmt.Println("Connect to binance...")
		symbols := map[string]string{CANDLE_SYMBOL: CANDLE_INTERVAL}
		doneC, _, err := futures.WsCombinedKlineServe(symbols, KlineEventHandlerFutures, errHandler)
		if err != nil {
			fmt.Println(err)
			continue
		}
		<-doneC

		fmt.Println("Disconnected, Reconnect in 3 seconds")
		time.Sleep(time.Second * 3)
	}
}

func KlineEventHandlerFutures(event *futures.WsKlineEvent) {
	secCandle := WebSocketCandleToKlineCandleFutures(event.Kline)
	fmt.Println(fmt.Sprintf("FUTURES: %s - Coin: %s, Price: %f", secCandle.CloseTime, secCandle.Symbol, secCandle.ClosePrice))

	if convertedCandle, ok := candleConverter.Convert(secCandle); ok {
		realBot.DoStuff(convertedCandle)
	}

	//if ticker.tick() {
	//	fmt.Println("__TICKER__")
	//	realBot.CheckBuyOrders()
	//}
}
