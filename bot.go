package main

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"reflect"
)

type Bot struct {
	Config                         *Config
	BuyIndicators                  []BuyIndicator
	SellIndicators                 []SellIndicator
	buffer                         *Buffer
	db                             *Database
	orderManager                   *OrderManager
	balance                        *Balance
	IsTrailingSellIndicatorEnabled bool
	trailingSellIndicator          *TrailingSellIndicator
}

func NewBot(config *Config) Bot {
	buffer := NewBuffer(resolveBufferSize(config))
	db := NewDatabase(*config)
	balance := NewBalance(*config)

	bot := Bot{
		Config:                         config,
		buffer:                         &buffer,
		db:                             &db,
		balance:                        &balance,
		IsTrailingSellIndicatorEnabled: false,
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
			Log(fmt.Sprintf("BUY: %s\nExchangeRate: %f", candle.CloseTime, price))
		}

		bot.buy()
	}
}

func (bot *Bot) runSellIndicators() {
	var eachIndicatorBuys [][]Buy

	for _, indicator := range bot.SellIndicators {
		indicator.Update()
		hasSignal, buys := indicator.HasSignal()
		if !hasSignal {
			return
		}

		eachIndicatorBuys = append(eachIndicatorBuys, buys)
	}

	// Sell
	for _, buy := range getIntersectedBuys(eachIndicatorBuys) {
		if IS_REAL_ENABLED {
			if USE_REAL_MONEY && !bot.IsTrailingSellIndicatorEnabled && orderManager.IsBuySold(CANDLE_SYMBOL, buy.RealOrderId) {
				bot.sell(buy)
				bot.finishSellIndicators(buy)
			}

			// has sell order
			if USE_REAL_MONEY && bot.IsTrailingSellIndicatorEnabled {
				if buy.HasSellOrder == 0 {
					bot.createRealMoneySellOrder(buy)
					bot.finishSellIndicators(buy)
				} else {
					if orderManager.IsBuySold(CANDLE_SYMBOL, buy.RealOrderId) {
						bot.sell(buy)
					}
				}
			}

			if !USE_REAL_MONEY {
				bot.sell(buy)
				bot.finishSellIndicators(buy)
			}
		} else {
			rev := bot.sell(buy)
			bot.finishSellIndicators(buy)
			candle := bot.buffer.GetLastCandle()
			Log(fmt.Sprintf(
				"SELL: %s\nExchangeRate: %f\nRevenue: %f",
				candle.CloseTime,
				bot.buffer.GetLastCandleClosePrice(),
				rev,
			))
		}
	}
}

func (bot *Bot) finishSellIndicators(buy Buy) {
	for _, indicator := range bot.SellIndicators {
		indicator.Finish(buy.Id)
	}
}

func (bot *Bot) buy() {
	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()

	coinsCount := bot.Config.TotalMoneyAmount / exchangeRate
	desiredPrice := bot.calcDesiredPrice(exchangeRate)

	if IS_REAL_ENABLED {
		rawPrice := candle.ClosePrice

		Log(fmt.Sprintf("GOT_BUY_SIGNAL\nPrice: %f", rawPrice))

		if USE_REAL_MONEY &&
			(!bot.orderManager.HasEnoughMoneyForBuy() ||
				!bot.orderManager.CanBuyForPrice(CANDLE_SYMBOL, rawPrice)) {
			return
		}

		if !USE_REAL_MONEY && !bot.balance.HasEnoughMoneyForBuy() {
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
		bot.balance.buy()

		buyId, _ := buyInsertResult.LastInsertId()
		bot.runAfterBuySellIndicators(buyId)

		Log(fmt.Sprintf("BUY\nPrice: %f\nQuantity: %f\nOrderId: %d", orderPrice, quantity, orderId))

		if USE_REAL_MONEY && !bot.IsTrailingSellIndicatorEnabled {
			upperPrice := CalcUpperPrice(orderPrice, bot.Config.HighSellPercentage)
			bot.createSellOrder(buyId, upperPrice, quantity)
		}
	} else {
		buyInsertResult := bot.db.AddBuy(
			CANDLE_SYMBOL,
			coinsCount,
			exchangeRate,
			desiredPrice,
			candle.CloseTime,
		)
		bot.balance.buy()

		buyId, _ := buyInsertResult.LastInsertId()
		bot.runAfterBuySellIndicators(buyId)
		PlotAddBuy(buyId, candle.CloseTime)
	}
}

func (bot *Bot) runAfterBuySellIndicators(buyId int64) {
	for _, indicator := range bot.SellIndicators {
		indicator.RunAfterBuy(buyId)
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

func (bot *Bot) createRealMoneySellOrder(buy Buy) {
	if !IS_REAL_ENABLED || !USE_REAL_MONEY || !bot.IsTrailingSellIndicatorEnabled {
		return
	}

	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()
	rev := calcRevenue(buy.RealQuantity, exchangeRate)

	if ok, buyItem := bot.trailingSellIndicator.GetBuyItemByBuyId(buy.Id); ok {
		orderId := bot.createSellOrder(buy.Id, buyItem.stopPrice, buy.RealQuantity)
		bot.db.UpdateRealBuyOrderId(buy.Id, orderId)

		Log(fmt.Sprintf("CREATE_SELL_ORDER\nPrice: %f - %f\nCalcedRevenue: %f", buy.ExchangeRate, candle.ClosePrice, rev))
	}
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

		Log(fmt.Sprintf("SELL\nPrice: %f - %f\nRevenue: %f", buy.ExchangeRate, candle.ClosePrice, rev))
	}

	bot.db.AddSell(
		CANDLE_SYMBOL,
		buy.Coins,
		exchangeRate,
		rev,
		buy.Id,
		candle.CloseTime,
	)
	bot.balance.sell()

	PlotAddSell(buy.Id, candle.CloseTime)

	return rev
}

func (bot *Bot) createSellOrder(buyId int64, sellPrice, quantity float64) int64 {
	//orderId := orderManager.CreateMarketSellOrder(CANDLE_SYMBOL, sellPrice, quantity)
	sellOrderId := orderManager.CreateSellOrder(CANDLE_SYMBOL, sellPrice, quantity)
	bot.db.UpdateRealBuyOrderId(buyId, sellOrderId)

	Log(fmt.Sprintf("SELL_ORDER\nOrderId: %d\nUpperPrice: %f", sellOrderId, sellPrice))

	return sellOrderId
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

	//buysCountIndicator := NewBuysCountIndicator(
	//	bot.Config,
	//	bot.buffer,
	//	bot.db,
	//)

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
		//&buysCountIndicator,
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

	//desiredPriceSellIndicator := NewDesiredPriceSellIndicator(
	//	bot.Config,
	//	bot.buffer,
	//	bot.db,
	//)

	trailingSellIndicator := NewTrailingSellIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	bot.SellIndicators = []SellIndicator{
		//&highPercentageSellIndicator,
		//&desiredPriceSellIndicator,
		&trailingSellIndicator,
	}

	bot.setIsTrailingSellIndicatorEnabled()
	if bot.IsTrailingSellIndicatorEnabled {
		bot.trailingSellIndicator = &trailingSellIndicator
	}
}

func (bot *Bot) setIsTrailingSellIndicatorEnabled() {
	for _, indicator := range bot.SellIndicators {
		if "*main.TrailingSellIndicator" == reflect.TypeOf(indicator).String() {
			bot.IsTrailingSellIndicatorEnabled = true
			return
		}
	}
}
