package main

type Bot struct {
	Config         *Config
	BuyIndicators  []BuyIndicator
	SellIndicators []SellIndicator
	buffer         Buffer
	db             Database
}

func NewBot(config *Config) Bot {
	bot := Bot{
		Config: config,
		buffer: NewBuffer(resolveBufferSize(config)),
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
	for _, indicator := range bot.BuyIndicators {
		if !indicator.HasSignal() {
			return
		}
	}

	bot.buy()
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
	if 1 == count {
		return eachIndicatorBuys[0]
	}

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
		config.PriceFallCandles,
	})
}

func setupBuyIndicators(bot *Bot) {
	backTrailingBuyIndicator := NewBackTrailingBuyIndicator()

	bot.BuyIndicators = []BuyIndicator{
		&backTrailingBuyIndicator,
	}
}

func setupSellIndicators(bot *Bot) {
	highPercentageSellIndicator := NewHighPercentageSellIndicator()

	bot.SellIndicators = []SellIndicator{
		&highPercentageSellIndicator,
	}
}
