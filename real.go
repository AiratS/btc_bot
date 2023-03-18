package main

import (
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"time"
)

//const BINANCE_API_KEY = "vYnGaa6tvn1JNFe3NSGOGen2Y5hAoBL7n4p59Etiu9nm9X4vZyaejmwec7pFuHM6"
//const BINANCE_SECRET_KEY = "Bb0oU4WdHCmF6HRLZbBvr0EVaLSOn2HMgAlOo5C9shu7Xzp2FK7cHSMjntwSL9gP"

const BINANCE_API_KEY = "vWNmAhOl6ad99Ytgfb2NElNBN5jzTYkToxv6kxjv6ddGQY8sMuPSv9Eq9lNoL4UB"
const BINANCE_SECRET_KEY = "rVUQ6YeLm6yFRwBvnpzDdKpjEEZ9bOqRLHDKaNLUERL4Kcckl08prdnEqciksWNj"

//const TG_API_KEY = "5055344139:AAGZrvVouQPWdU3Jn6p_4ipvCwSDOubVn-4" // телеграм фаболос//
//const TG_API_KEY = "2083210132:AAHHp9h2dziqJbB1J9ySmv3vBoG8FCDLVag" // телеграм лакки//
//const TG_API_KEY = "5421904898:AAGda6XvZlYZIUFD4ZStPo2U9kGH7_2CTng" // телеграм Ayaz
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

		TrailingSellActivationAdditionPercentage: 1,
		TrailingSellStopPercentage:               1,
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
	// 324037113, 135933418
	for _, chatID := range getChatIDs() {
		msg := tgbotapi.NewMessage(chatID, msg)
		tgBot.Send(msg)
	}
}

func SendImage(fileName string) {
	photoBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}
	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: photoBytes,
	}

	for _, chatID := range getChatIDs() {
		_, err2 := tgBot.Send(tgbotapi.NewPhoto(chatID, photoFileBytes))
		if err2 != nil {
			fmt.Println(err2)
		}
	}
}

func getChatIDs() []int64 {
	return []int64{324037113, 135933418}
}
