package main

import (
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

//const BINANCE_API_KEY = "vYnGaa6tvn1JNFe3NSGOGen2Y5hAoBL7n4p59Etiu9nm9X4vZyaejmwec7pFuHM6"
//const BINANCE_SECRET_KEY = "Bb0oU4WdHCmF6HRLZbBvr0EVaLSOn2HMgAlOo5C9shu7Xzp2FK7cHSMjntwSL9gP"

const BINANCE_API_KEY = "vWNmAhOl6ad99Ytgfb2NElNBN5jzTYkToxv6kxjv6ddGQY8sMuPSv9Eq9lNoL4UB"
const BINANCE_SECRET_KEY = "rVUQ6YeLm6yFRwBvnpzDdKpjEEZ9bOqRLHDKaNLUERL4Kcckl08prdnEqciksWNj"

const TG_API_KEY = "5174010399:AAGfCYIHJ6ToTIovKhgemgJXWQCi3A4CxJg" // телеграм амайзинг

var tgBot *tgbotapi.BotAPI
var candleConverter SecToMinCandleIntervalConverter
var orderManager OrderManager
var realBot Bot

func GetRealBotConfig() Config {
	return Config{
		HighSellPercentage: 0.1,

		TrailingTopPercentage:           0.1,
		TrailingUpdateTimesBeforeFinish: 1,

		WaitAfterLastBuyPeriod: 1,

		BigFallCandlesCount: 1,
		BigFallPercentage:   0.1,

		DesiredPriceCandles: 1,

		GradientDescentCandles:  1,
		GradientDescentPeriod:   1,
		GradientDescentGradient: 1,
	}
}

func RunRealTime() {
	client := binance.NewClient(BINANCE_API_KEY, BINANCE_SECRET_KEY)
	tgBot, _ = tgbotapi.NewBotAPI(TG_API_KEY)

	candleConverter = NewSecToMinCandleConverter()
	config := GetRealBotConfig()
	realBot = NewRealBot(&config, client)

	errHandler := func(err error) {
		fmt.Println(err)
	}

	for {
		fmt.Println("Connect to binance...")
		symbols := map[string]string{"BTCUSDT": CANDLE_INTERVAL}
		doneC, _, err := binance.WsCombinedKlineServe(symbols, KlineEventHandler, errHandler)
		if err != nil {
			fmt.Println(err)
			continue
		}
		<-doneC

		fmt.Println("Disconnected, Reconnect in 3 seconds")
		time.Sleep(time.Second * 3)
	}
}

func KlineEventHandler(event *binance.WsKlineEvent) {
	secCandle := WebSocketCandleToKlineCandle(event.Kline)
	fmt.Println(fmt.Sprintf("%s - Coin: %s, Price: %f", secCandle.CloseTime, secCandle.Symbol, secCandle.ClosePrice))

	if convertedCandle, ok := candleConverter.Convert(secCandle); ok {
		realBot.DoStuff(convertedCandle)
	}
}

func SendTgBotMessage(msg string) {
	// 324037113
	chats := []int64{135933418, 324037113}
	for _, chatID := range chats {
		msg := tgbotapi.NewMessage(chatID, msg)
		tgBot.Send(msg)
	}
}
