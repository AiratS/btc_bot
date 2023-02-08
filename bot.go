package main

import "fmt"

type Bot struct {
	Config         *Config
	BuyIndicators  []BuyIndicator
	SellIndicators []SellIndicator
	buffer         *Buffer
	db             Database
}

func NewBot(config *Config) Bot {
	buffer := NewBuffer(resolveBufferSize(config))

	bot := Bot{
		Config: config,
		buffer: &buffer,
		db:     NewDatabase(),
	}

	setupBuyIndicators(&bot)
	setupSellIndicators(&bot)

	return bot
}

func (bot *Bot) DoStuff(candle Candle) {
	bot.buffer.AddCandle(candle)
	bot.runBuyIndicators()
	bot.runSellIndicators()
}

func (bot *Bot) runBuyIndicators() {
	signalsCount := 0

	for _, indicator := range bot.BuyIndicators {
		indicator.Update()
		if indicator.HasSignal() {
			signalsCount++
		}
	}

	if len(bot.BuyIndicators) == signalsCount {
		LogAndPrint(fmt.Sprintf("Buy signal, ExchangeRate: %f", bot.buffer.GetLastCandleClosePrice()))
		bot.buy()
	}
}

func (bot *Bot) runSellIndicators() {
	var eachIndicatorBuys [][]Buy

	for _, indicator := range bot.SellIndicators {
		hasSignal, buys := indicator.HasSignal()
		if !hasSignal {
			return
		}

		eachIndicatorBuys = append(eachIndicatorBuys, buys)
	}

	// Sell
	for _, buy := range getIntersectedBuys(eachIndicatorBuys) {
		LogAndPrint(fmt.Sprintf("Sell signal, ExchangeRate: %f", bot.buffer.GetLastCandleClosePrice()))
		bot.sell(buy)
	}
}

func (bot *Bot) buy() {
	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()

	coinsCount := TOTAL_MONEY_AMOUNT / exchangeRate
	bot.db.AddBuy(
		CANDLE_SYMBOL,
		coinsCount,
		exchangeRate,
		candle.CloseTime,
	)
}

func (bot *Bot) sell(buy Buy) {
	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()

	bot.db.AddSell(
		CANDLE_SYMBOL,
		buy.Coins,
		exchangeRate,
		calcRevenue(buy.Coins, exchangeRate),
		buy.Id,
		candle.CloseTime,
	)
}

func calcRevenue(coinsCounts, exchangeRate float64) float64 {
	return coinsCounts * exchangeRate
}

func getIntersectedBuys(eachIndicatorBuys [][]Buy) []Buy {
	count := len(eachIndicatorBuys)
	firstBuys := eachIndicatorBuys[0]

	for index, _ := range eachIndicatorBuys {
		if index == (count - 1) {
			break
		}

		secondBuys := eachIndicatorBuys[index+1]
		firstBuys = BuySliceIntersect(firstBuys, secondBuys)
	}

	return firstBuys
}

func resolveBufferSize(config *Config) int {
	return MaxInt([]int{
		// add your candles
		100,
	})
}

func setupBuyIndicators(bot *Bot) {
	backTrailingBuyIndicator := NewBackTrailingBuyIndicator(
		bot.Config,
		bot.buffer,
		&bot.db,
	)

	bot.BuyIndicators = []BuyIndicator{
		&backTrailingBuyIndicator,
	}
}

func setupSellIndicators(bot *Bot) {
	highPercentageSellIndicator := NewHighPercentageSellIndicator(
		bot.Config,
		bot.buffer,
		&bot.db,
	)

	bot.SellIndicators = []SellIndicator{
		&highPercentageSellIndicator,
	}
}
