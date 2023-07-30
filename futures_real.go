package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
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

	wg := sync.WaitGroup{}
	wg.Add(2)
	go runWsCombinedKlineServe(client)
	go runWsUserDataServe(client, &realBot)
	wg.Wait()
}

func runWsCombinedKlineServe(client *futures.Client) {
	for {
		fmt.Println("Connect to binance...")
		symbols := map[string]string{CANDLE_SYMBOL: CANDLE_INTERVAL}
		doneC, _, err := futures.WsCombinedKlineServe(symbols, KlineEventHandlerFutures, func(err error) {
			fmt.Println(err)
		})
		if err != nil {
			fmt.Println(err)
		}
		<-doneC

		fmt.Println("Disconnected, Reconnect in 3 seconds")
		time.Sleep(time.Second * 3)
	}
}

func runWsUserDataServe(client *futures.Client, bot *Bot) {
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		panic(err)
		return
	}

	for {
		doneA, _, _ := futures.WsUserDataServe(listenKey, func(event *futures.WsUserDataEvent) {
			orderData := event.OrderTradeUpdate

			if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
				orderData.Side == futures.SideTypeBuy &&
				orderData.Type == futures.OrderTypeLimit &&
				orderData.Status == futures.OrderStatusTypeFilled {

				fmt.Println("OrderTradeUpdate", event.OrderTradeUpdate)
				fmt.Printf("OrderData: %+v\n", orderData)
				bot.OnLimitBuyFilled(orderData.ID)
			}
		}, func(err error) {
			fmt.Println(err)
		})
		<-doneA

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
