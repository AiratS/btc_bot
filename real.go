package main

import (
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"time"
)

// Amazing
//const BINANCE_API_KEY = "vWNmAhOl6ad99Ytgfb2NElNBN5jzTYkToxv6kxjv6ddGQY8sMuPSv9Eq9lNoL4UB"
//const BINANCE_SECRET_KEY = "rVUQ6YeLm6yFRwBvnpzDdKpjEEZ9bOqRLHDKaNLUERL4Kcckl08prdnEqciksWNj"

// Lucky
//const BINANCE_API_KEY = "VMWeJjjUvQNrkZxNd4IG6goKpmAb8aooRWtKAFR7xG7dhrtgNYLvScMFUz7vz5E9"    // 3 акк лакки
//const BINANCE_SECRET_KEY = "kGuxVBS4NOrmkGlELNhUaCCoLdovcmNrk9bNGL4MTZwsOTsZocDxz8PsKY4npPXl" // 3 акк лакки

// Ayaz
const BINANCE_API_KEY = "blXKwv9BVFS52FqA6PePWGhNrFgolagVlyJahr94GkASHkxvD7sDBJImPAWiICHx"    // второй акк аяз
const BINANCE_SECRET_KEY = "fxak06hMNtaSxyvTh6v8JO0uC7y8JgiRqFcNcnVfy37merNuBCY432ouKBwAqPsk" // второй акк аяз

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
		HighSellPercentage:         0.1,
		FirstBuyHighSellPercentage: 0.1,

		TrailingTopPercentage:           0.1,
		TrailingUpdateTimesBeforeFinish: 1,

		WaitAfterLastBuyPeriod: 1,

		BigFallCandlesCount: 1,
		BigFallSmoothPeriod: 1,
		BigFallPercentage:   0.1,

		DesiredPriceCandles: 1,

		GradientDescentCandles:  1,
		GradientDescentPeriod:   1,
		GradientDescentGradient: 1,

		TrailingSellActivationAdditionPercentage: 1,
		TrailingSellStopPercentage:               1,

		LinearRegressionCandles:   1,
		LinearRegressionPeriod:    1,
		LinearRegressionMse:       1,
		LinearRegressionK:         1,
		LinearRegressionDeviation: 1,

		GradientSwingIndicatorCandles:   1,
		GradientSwingIndicatorPeriod:    1,
		GradientSwingIndicatorSwingType: 1,

		CatchingFallingKnifeCandles:   1,
		CatchingFallingKnifeSellPercentage:   1,
		CatchingFallingKnifeAdditionalBuyPercentage:   1,

		TotalMoneyAmount:                1000,
		TotalMoneyIncreasePercentage:    10,
		FirstBuyMoneyIncreasePercentage: 1,
		StopIncreaseMoneyAfterBuysCount: 1,
		Leverage:                        10,

		LessThanPreviousBuyPercentage: 1,

		BoostBuyFallPercentage:          1,
		BoostBuyPeriodMinutes:           1,
		BoostBuyMoneyIncreasePercentage: 1,

		StopAfterUnsuccessfullySellMinutes: 1,
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


	if IS_REAL_ENABLED {
		Log(fmt.Sprintf("%f", secCandle.GetPrice()))

		realBot.runBuyIndicators(secCandle)
	}

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
