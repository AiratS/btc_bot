package main

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
)

type Bot struct {
	Config         *Config
	BuyIndicators  []BuyIndicator
	SellIndicators []SellIndicator
	buffer         *Buffer
	db             *Database
	orderManager   *OrderManager
}

func NewBot(config *Config) Bot {
	buffer := NewBuffer(resolveBufferSize(config))
	db := NewDatabase()

	bot := Bot{
		Config: config,
		buffer: &buffer,
		db:     &db,
	}

	setupBuyIndicators(&bot)
	setupSellIndicators(&bot)

	return bot
}

func NewRealBot(config *Config, binanceClient *binance.Client) Bot {
	orderManager = NewOrderManager(binanceClient)
	bot := NewBot(config)
	bot.orderManager = &orderManager

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
		for _, indicator := range bot.BuyIndicators {
			indicator.Finish()
		}

		candle := bot.buffer.GetLastCandle()
		price := bot.buffer.GetLastCandleClosePrice()

		if !IS_REAL_ENABLED {
			LogAndPrint(fmt.Sprintf("Buy signal, Created at: %s, ExchangeRate: %f", candle.CloseTime, price))
		}

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
		if IS_REAL_ENABLED {
			if USE_REAL_MONEY && orderManager.IsBuySold(CANDLE_SYMBOL, buy.RealOrderId) {
				bot.sell(buy)
			}

			if !USE_REAL_MONEY {
				bot.sell(buy)
			}
		} else {
			rev := bot.sell(buy)
			candle := bot.buffer.GetLastCandle()
			LogAndPrint(fmt.Sprintf("Sell signal, Created At: %s, ExchangeRate: %f: Revenue: %f", candle.CloseTime, bot.buffer.GetLastCandleClosePrice(), rev))
		}
	}
}

func (bot *Bot) buy() {
	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()

	coinsCount := TOTAL_MONEY_AMOUNT / exchangeRate
	desiredPrice := bot.calcDesiredPrice(exchangeRate)

	if IS_REAL_ENABLED {
		rawPrice := candle.ClosePrice

		LogAndPrintAndSendTg(fmt.Sprintf("GOT_BUY_SIGNAL\nPrice: %f", rawPrice))

		if USE_REAL_MONEY &&
			(!bot.orderManager.HasEnoughMoneyForBuy() ||
				!bot.orderManager.CanBuyForPrice(CANDLE_SYMBOL, rawPrice)) {
			return
		}

		orderId, quantity, orderPrice := bot.orderManager.CreateMarketBuyOrder(candle.Symbol, rawPrice)

		if !USE_REAL_MONEY {
			quantity = coinsCount
		}

		buyInsertResult := bot.db.AddRealBuy(
			CANDLE_SYMBOL,
			coinsCount,
			orderPrice,
			desiredPrice,
			candle.CloseTime,
			orderId,
			quantity,
		)

		LogAndPrintAndSendTg(fmt.Sprintf("BUY\nPrice: %f\nQuantity: %f\nOrderId: %d", orderPrice, quantity, orderId))

		if USE_REAL_MONEY {
			upperPrice := CalcUpperPrice(orderPrice, bot.Config.HighSellPercentage)
			sellOrderId := orderManager.CreateSellOrder(candle.Symbol, upperPrice, quantity)

			buyId, _ := buyInsertResult.LastInsertId()
			bot.db.UpdateRealBuyOrderId(buyId, sellOrderId)

			LogAndPrintAndSendTg(fmt.Sprintf("SELL_ORDER\nOrderId: %d\nUpperPrice: %f", sellOrderId, upperPrice))
		}
	} else {
		bot.db.AddBuy(
			CANDLE_SYMBOL,
			coinsCount,
			exchangeRate,
			desiredPrice,
			candle.CloseTime,
		)
	}
}

func (bot *Bot) calcDesiredPrice(currentPrice float64) float64 {
	upperPrice := CalcUpperPrice(currentPrice, bot.Config.HighSellPercentage)
	if bot.Config.DesiredPriceCandles > len(bot.buffer.GetCandles()) {
		return upperPrice
	}

	medPrice := Median(GetClosePrices(bot.buffer.GetCandles()))
	if upperPrice < medPrice {
		return medPrice
	}

	return upperPrice
}

func (bot *Bot) sell(buy Buy) float64 {
	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()
	rev := calcRevenue(buy.Coins, exchangeRate)

	if IS_REAL_ENABLED {
		rev = calcRevenue(buy.RealQuantity, exchangeRate)
		//orderId := orderManager.CreateSellOrder(candle.Symbol, candle.ClosePrice, buy.RealQuantity)
		//orderId := orderManager.CreateMarketSellOrder(candle.Symbol, candle.ClosePrice, buy.RealQuantity)
		//bot.db.UpdateRealBuyOrderId(buy.Id, orderId)

		LogAndPrintAndSendTg(fmt.Sprintf("SELL\nPrice: %f - %f\nRevenue: %f", buy.ExchangeRate, candle.ClosePrice, rev))
	}

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
		config.BigFallCandlesCount,
		config.DesiredPriceCandles,
		config.GradientDescentCandles,
		config.GradientDescentPeriod,
	}) + 10
}

func setupBuyIndicators(bot *Bot) {
	//backTrailingBuyIndicator := NewBackTrailingBuyIndicator(
	//	bot.Config,
	//	bot.buffer,
	//	bot.db,
	//)

	buysCountIndicator := NewBuysCountIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	waitForPeriodIndicator := NewWaitForPeriodIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	bigFallIndicator := NewBigFallIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	gradientDescentIndicator := NewGradientDescentIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	bot.BuyIndicators = []BuyIndicator{
		//&backTrailingBuyIndicator,
		&buysCountIndicator,
		&waitForPeriodIndicator,
		&bigFallIndicator,
		&gradientDescentIndicator,
	}
}

func setupSellIndicators(bot *Bot) {
	//highPercentageSellIndicator := NewHighPercentageSellIndicator(
	//	bot.Config,
	//	bot.buffer,
	//	bot.db,
	//)

	desiredPriceSellIndicator := NewDesiredPriceSellIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	bot.SellIndicators = []SellIndicator{
		//&highPercentageSellIndicator,
		&desiredPriceSellIndicator,
	}
}
