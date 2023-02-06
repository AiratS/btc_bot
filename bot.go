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
}

func (bot *Bot) runSellIndicators() {
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
