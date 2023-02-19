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

func (bot *Bot) Kill() {
	bot.db.connect.Close()
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
		candle := bot.buffer.GetLastCandle()
		LogAndPrint(fmt.Sprintf("Buy signal, Created at: %s, ExchangeRate: %f", candle.CloseTime, bot.buffer.GetLastCandleClosePrice()))
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
		candle := bot.buffer.GetLastCandle()
		rev := bot.sell(buy)
		LogAndPrint(fmt.Sprintf("Sell signal, Created At: %s, ExchangeRate: %f: Revenue: %f", candle.CloseTime, bot.buffer.GetLastCandleClosePrice(), rev))
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

func (bot *Bot) sell(buy Buy) float64 {
	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()
	rev := calcRevenue(buy.Coins, exchangeRate)

	bot.db.AddSell(
		CANDLE_SYMBOL,
		buy.Coins,
		exchangeRate,
		rev,
		buy.Id,
		candle.CloseTime,
	)

	return rev
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

	buysCountIndicator := NewBuysCountIndicator(
		bot.Config,
		bot.buffer,
		&bot.db,
	)

	bot.BuyIndicators = []BuyIndicator{
		&backTrailingBuyIndicator,
		&buysCountIndicator,
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
