package main

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"math"
	"reflect"
)

type Bot struct {
	Config                         *Config
	BuyIndicators                  []BuyIndicator
	SellIndicators                 []SellIndicator
	buffer                         *Buffer
	db                             *Database
	orderManager                   *OrderManager
	futuresOrderManager            *FuturesOrderManager
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

func NewFuturesRealBot(config *Config, futuresClient *futures.Client) Bot {
	futuresOrderManager := NewFuturesOrderManager(futuresClient)
	bot := NewBot(config)
	bot.futuresOrderManager = &futuresOrderManager

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

		if !USE_REAL_MONEY && !bot.balance.HasEnoughMoneyForBuy() {
			return
		}

		if !IS_REAL_ENABLED {
			Log(fmt.Sprintf("BUY: %s\nExchangeRate: %f", candle.CloseTime, price))
		}

		if USE_REAL_MONEY && BUY_ORDER_REDUCTION_ENABLED {
			bot.createLimitBuyOrder(candle)
		} else {
			bot.buy(candle.GetPrice(), candle.CloseTime, &BuyOrder{})
		}
	}

	if USE_REAL_MONEY && BUY_ORDER_REDUCTION_ENABLED {
		bot.checkBuyOrders()
	}
}

func (bot *Bot) createLimitBuyOrder(candle Candle) {
	Log(fmt.Sprintf("GOT_BUY_SIGNAL\nPrice: %f", candle.ClosePrice))

	buyPrice := CalcBottomPrice(candle.ClosePrice, BUY_ORDER_REDUCTION_PERCENTAGE)
	if USE_REAL_MONEY &&
		(!bot.HasEnoughMoneyForBuy() || !bot.CanBuyForPrice(CANDLE_SYMBOL, buyPrice)) {
		return
	}

	orderId, quantity, realBuyPrice :=
		bot.futuresOrderManager.CreateBuyOrder(CANDLE_SYMBOL, buyPrice)

	Log(fmt.Sprintf(
		"CREATE_LIMIT_BUY_ORDER\nOrderId: %d\nQuantity: %f\nBuyPrice: %f\nExchangeRate: %f",
		orderId,
		quantity,
		realBuyPrice,
		candle.ClosePrice,
	))

	bot.db.AddNewBuyOrder(
		CANDLE_SYMBOL,
		quantity,
		candle.ClosePrice,
		realBuyPrice,
		orderId,
		candle.CloseTime,
	)
}

func (bot *Bot) checkBuyOrders() {
	currentCandle := bot.buffer.GetLastCandle()

	newBuyOrders := bot.db.FetchNewBuyOrders(CANDLE_SYMBOL)
	rejectBuyOrders := bot.db.FetchRejectBuyOrders(CANDLE_SYMBOL, currentCandle.CloseTime)

	for _, buyOrder := range newBuyOrders {
		// Save buy
		if bot.IsOrderFilled(CANDLE_SYMBOL, buyOrder.RealOrderId) {
			bot.db.UpdateBuyOrderStatus(buyOrder.Id, BuyOrderStatusFilled)
			bot.buy(buyOrder.BuyPrice, currentCandle.CloseTime, &buyOrder)
			continue
		}

		// Reject buy orders
		if bot.isRejectionBuyOrder(buyOrder.Id, &rejectBuyOrders) {
			bot.futuresOrderManager.CancelOrder(CANDLE_SYMBOL, buyOrder.RealOrderId)
			bot.db.UpdateBuyOrderStatus(buyOrder.Id, BuyOrderStatusRejected)
		}
	}
}

func (bot *Bot) isRejectionBuyOrder(buyOrderId int64, rejectionBuyOrders *[]BuyOrder) bool {
	for _, buyOrder := range *rejectionBuyOrders {
		if buyOrder.Id == buyOrderId {
			return true
		}
	}

	return false
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
			if !USE_REAL_MONEY {
				bot.sell(buy)
				bot.finishSellIndicators(buy)
				continue
			}

			if ENABLE_FUTURES && ENABLE_TIME_CANCEL && buy.BuyType == TimeCancel {
				bot.sell(buy)
				bot.finishSellIndicators(buy)
				continue
			}

			if bot.IsOrderFilled(CANDLE_SYMBOL, buy.RealOrderId) {
				Log(fmt.Sprintf("IS_BUY_SOLD: YES\nOrderId: %d", buy.RealOrderId))

				bot.sell(buy)
				bot.finishSellIndicators(buy)
				continue
			} else {
				Log(fmt.Sprintf("IS_BUY_SOLD: NO\nOrderId: %d", buy.RealOrderId))
				continue
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

func (bot *Bot) IsOrderFilled(symbol string, realOrderId int64) bool {
	if ENABLE_FUTURES {
		return bot.futuresOrderManager.IsBuySold(symbol, realOrderId)
	}

	return bot.orderManager.IsBuySold(symbol, realOrderId)
}

func (bot *Bot) finishSellIndicators(buy Buy) {
	for _, indicator := range bot.SellIndicators {
		indicator.Finish(buy.Id)
	}
}

func (bot *Bot) buy(exchangeRate float64, closeTime string, buyOrder *BuyOrder) {
	//candle := bot.buffer.GetLastCandle()
	//exchangeRate := candle.GetPrice()

	desiredPrice := bot.calcDesiredPrice(exchangeRate)

	if !USE_REAL_MONEY && !bot.balance.HasEnoughMoneyForBuy() {
		return
	}

	if IS_REAL_ENABLED {
		coinsCount := bot.Config.TotalMoneyAmount / exchangeRate
		if ENABLE_FUTURES {
			coinsCount = (bot.Config.TotalMoneyAmount * LEVERAGE) / exchangeRate
		}
		rawPrice := exchangeRate

		Log(fmt.Sprintf("GOT_BUY_SIGNAL\nPrice: %f", rawPrice))

		// Buy order reduction ENABLED
		if USE_REAL_MONEY && BUY_ORDER_REDUCTION_ENABLED {
			buyInsertResult := bot.db.AddRealBuy(
				CANDLE_SYMBOL,
				buyOrder.Coins,
				buyOrder.BuyPrice,
				desiredPrice,
				closeTime,
				buyOrder.RealOrderId,
				buyOrder.Coins,
			)
			bot.balance.buy()

			buyId, _ := buyInsertResult.LastInsertId()
			Log(fmt.Sprintf(
				"BUY_FILLED\nPrice: %f\nQuantity: %f\nOrderId: %d",
				buyOrder.BuyPrice,
				buyOrder.Coins,
				buyOrder.RealOrderId,
			))
			bot.runAfterBuy(buyId)
			return
		}

		// Buy order reduction DISABLED
		if USE_REAL_MONEY &&
			(!bot.HasEnoughMoneyForBuy() || !bot.CanBuyForPrice(CANDLE_SYMBOL, rawPrice)) {
			return
		}

		orderId, quantity, orderPrice := bot.CreateMarketBuyOrder(CANDLE_SYMBOL, rawPrice)

		if !USE_REAL_MONEY {
			quantity = coinsCount
		}

		buyInsertResult := bot.db.AddRealBuy(
			CANDLE_SYMBOL,
			coinsCount,
			orderPrice,
			desiredPrice,
			closeTime,
			orderId,
			quantity,
		)
		bot.balance.buy()

		buyId, _ := buyInsertResult.LastInsertId()
		Log(fmt.Sprintf("BUY\nPrice: %f\nQuantity: %f\nOrderId: %d", orderPrice, quantity, orderId))
		bot.runAfterBuy(buyId)

		//if USE_REAL_MONEY && !bot.IsTrailingSellIndicatorEnabled {
		//	upperPrice := CalcUpperPrice(orderPrice, bot.Config.HighSellPercentage)
		//	bot.createAndUpdateSellOrder(buyId, upperPrice, quantity)
		//}
	} else {
		coinsCount := bot.Config.TotalMoneyAmount / exchangeRate
		if ENABLE_FUTURES {
			coinsCount = (bot.Config.TotalMoneyAmount * float64(bot.Config.Leverage)) / exchangeRate
		}

		buyInsertResult := bot.db.AddBuy(
			CANDLE_SYMBOL,
			coinsCount,
			exchangeRate,
			desiredPrice,
			closeTime,
		)
		bot.balance.buy()

		buyId, _ := buyInsertResult.LastInsertId()
		bot.runAfterBuy(buyId)
		PlotAddBuy(buyId, closeTime)
	}
}

func (bot *Bot) HasEnoughMoneyForBuy() bool {
	if ENABLE_FUTURES {
		return bot.futuresOrderManager.HasEnoughMoneyForBuy()
	}

	return bot.orderManager.HasEnoughMoneyForBuy()
}

func (bot *Bot) CanBuyForPrice(symbol string, rawPrice float64) bool {
	if ENABLE_FUTURES {
		return bot.futuresOrderManager.CanBuyForPrice(symbol, rawPrice)
	}

	return bot.orderManager.CanBuyForPrice(symbol, rawPrice)
}

func (bot *Bot) CreateMarketBuyOrder(symbol string, rawPrice float64) (int64, float64, float64) {
	if ENABLE_FUTURES {
		return bot.futuresOrderManager.CreateMarketBuyOrder(symbol, rawPrice)
	}

	return bot.orderManager.CreateMarketBuyOrder(symbol, rawPrice)
}

func (bot *Bot) runAfterBuy(buyId int64) {
	// Run sell indicators hooks
	for _, indicator := range bot.SellIndicators {
		indicator.RunAfterBuy(buyId)
	}

	// Recalculate AVG futures prices
	unsoldBuys := bot.db.FetchUnsoldBuys()
	avgFuturesPrice := CalcFuturesAvgPrice(unsoldBuys)
	desiredSellPrice := CalcUpperPrice(avgFuturesPrice, bot.Config.HighSellPercentage)

	liquidationPrice := CalcBottomPrice(avgFuturesPrice, GetLeverageLiquidationPercentage(bot.Config.Leverage))
	Log(fmt.Sprintf(
		"AVG_PRICE\nAvgPrice: %f\nDesirecSellPrice: %f\nLiquidationPrice: %f",
		avgFuturesPrice,
		desiredSellPrice,
		liquidationPrice,
	))

	lastIdx := len(unsoldBuys) - 1
	for idx, buy := range unsoldBuys {
		if !USE_REAL_MONEY {
			bot.db.UpdateDesiredPriceByBuyId(buy.Id, desiredSellPrice)
			continue
		}

		if !IS_REAL_ENABLED {
			panic("Disable real time!")
		}

		if idx != lastIdx {
			Log(fmt.Sprintf("CANCEL_ORDER\nOrderId: %d\n", buy.RealOrderId))
			bot.futuresOrderManager.CancelOrder(CANDLE_SYMBOL, buy.RealOrderId)
		}

		bot.createAndUpdateSellOrder(buy.Id, desiredSellPrice, buy.RealQuantity)
	}
}

func (bot *Bot) calcDesiredPrice(currentPrice float64) float64 {
	upperPrice := CalcUpperPrice(currentPrice, bot.Config.HighSellPercentage)
	return upperPrice

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
	rev := bot.calcRevenue(buy.RealQuantity, bot.Config.HighSellPercentage, buy.ExchangeRate)

	if ok, buyItem := bot.trailingSellIndicator.GetBuyItemByBuyId(buy.Id); ok {
		orderId := bot.createAndUpdateSellOrder(buy.Id, buyItem.stopPrice, buy.RealQuantity)
		bot.db.UpdateRealBuyOrderId(buy.Id, orderId)

		Log(fmt.Sprintf("CREATE_SELL_ORDER\nPrice: %f - %f\nCalcedRevenue: %f", buy.ExchangeRate, candle.ClosePrice, rev))
	}
}

func (bot *Bot) sell(buy Buy) float64 {
	candle := bot.buffer.GetLastCandle()
	exchangeRate := candle.GetPrice()
	//rev := bot.calcRevenue(buy.Coins, bot.Config.HighSellPercentage, buy.ExchangeRate)
	rev := bot.calcFuturesRevenue(buy)

	if IS_REAL_ENABLED {
		//rev = bot.calcRevenue(buy.RealQuantity, bot.Config.HighSellPercentage, buy.ExchangeRate)
		//orderId := orderManager.CreateSellOrder(candle.Symbol, candle.ClosePrice, buy.RealQuantity)
		//orderId := orderManager.CreateMarketSellOrder(candle.Symbol, candle.ClosePrice, buy.RealQuantity)
		//bot.db.UpdateRealBuyOrderId(buy.Id, orderId)

		Log(fmt.Sprintf("SELL\nPrice: %f - %f\nRevenue: %f", buy.ExchangeRate, candle.ClosePrice, rev))
	}

	returnMoney := bot.Config.TotalMoneyAmount

	if ENABLE_FUTURES {
		if buy.BuyType == Liquidation {
			rev = 0
			returnMoney = 0

			Log(fmt.Sprintf("GOT_LIQUIDATION\nOrderId: %d\n", buy.RealOrderId))
		} else if buy.BuyType == TimeCancel {
			rev = bot.calcFuturesTimeCancelRevenue(buy.Coins, buy.ExchangeRate, exchangeRate)

			Log(fmt.Sprintf("GOT_TIME_CANCEL\nOrderId: %d\n", buy.RealOrderId))

			if IS_REAL_ENABLED && USE_REAL_MONEY {
				Log(fmt.Sprintf("CANCEL_ORDER\nOrderId: %d\n", buy.RealOrderId))
				bot.futuresOrderManager.CancelOrder(CANDLE_SYMBOL, buy.RealOrderId)
				bot.createAndUpdateSellOrder(buy.Id, exchangeRate, buy.RealQuantity)
			}

			if rev > bot.Config.TotalMoneyAmount {
				returnMoney = bot.Config.TotalMoneyAmount
			} else {
				returnMoney = bot.Config.TotalMoneyAmount - math.Abs(rev)
			}
		} else if buy.BuyType == Default {
			//if ADD_REVENUE_TO_BALANCE {
			//	returnMoney = bot.Config.TotalMoneyAmount +
			//		(rev - (bot.Config.TotalMoneyAmount * float64(bot.Config.Leverage)))
			//}
		}
	}

	Log(fmt.Sprintf("JUST_ADD_SELL\nOrderId: %d\n", buy.RealOrderId))
	bot.db.AddSell(
		CANDLE_SYMBOL,
		buy.Coins,
		exchangeRate,
		rev,
		buy.Id,
		candle.CloseTime,
	)

	bot.balance.sell(returnMoney)

	PlotAddSell(buy.Id, candle.CloseTime)

	return rev
}

func (bot *Bot) calcFuturesTimeCancelRevenue(coinsCount, buyPrice, currentPrice float64) float64 {
	percentage := CalcGrowth(buyPrice, currentPrice)

	if percentage < 0 {
		//leverage := float64(bot.Config.Leverage)
		totalMoney := bot.Config.TotalMoneyAmount // * leverage

		liqPercentage := GetLeverageLiquidationPercentage(bot.Config.Leverage)
		lose := buyPrice - currentPrice
		lPrice := CalcBottomPrice(buyPrice, liqPercentage)
		total := buyPrice - lPrice
		minusPercentage := (lose * 100) / total

		minus := CalcValuePercentage(
			totalMoney,
			minusPercentage,
		)

		return -minus // -(totalMoney - minus)
	}

	return bot.calcRevenue(coinsCount, percentage, buyPrice)
}

func (bot *Bot) createAndUpdateSellOrder(buyId int64, sellPrice, quantity float64) int64 {
	//orderId := orderManager.CreateMarketSellOrder(CANDLE_SYMBOL, sellPrice, quantity)
	sellOrderId := bot.CreateSellOrder(CANDLE_SYMBOL, sellPrice, quantity)
	bot.db.UpdateRealBuyOrderId(buyId, sellOrderId)

	Log(fmt.Sprintf("SELL_ORDER\nOrderId: %d\nUpperPrice: %f", sellOrderId, sellPrice))

	return sellOrderId
}

func (bot *Bot) CreateSellOrder(symbol string, sellPrice, quantity float64) int64 {
	if ENABLE_FUTURES {
		return bot.futuresOrderManager.CreateSellOrder(symbol, sellPrice, quantity)
	}

	return bot.orderManager.CreateSellOrder(symbol, sellPrice, quantity)
}

func (bot *Bot) calcRevenue(coinsCounts, upperPercentage, buyExchangeRate float64) float64 {
	additionalPrice := (buyExchangeRate * upperPercentage) / 100
	sellPrice := buyExchangeRate + additionalPrice

	return coinsCounts * sellPrice
}

func (bot *Bot) calcFuturesRevenue(buy Buy) float64 {
	if IS_REAL_ENABLED {
		return buy.RealQuantity * buy.DesiredPrice
	}

	return buy.Coins * buy.DesiredPrice
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
	}) + 1
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

	//waitForPeriodIndicator := NewWaitForPeriodIndicator(
	//	bot.Config,
	//	bot.buffer,
	//	bot.db,
	//)

	bigFallIndicator := NewBigFallIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	lessThanPreviousBuyIndicator := NewLessThanPreviousBuyIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	stopAfterUnsuccessfullySellIndicator := NewStopAfterUnsuccessfullySellIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	//gradientDescentIndicator := NewGradientDescentIndicator(
	//	bot.Config,
	//	bot.buffer,
	//	bot.db,
	//)

	bot.BuyIndicators = []BuyIndicator{
		//&backTrailingBuyIndicator,
		//&buysCountIndicator,
		//&waitForPeriodIndicator,
		&bigFallIndicator,
		&lessThanPreviousBuyIndicator,
		&stopAfterUnsuccessfullySellIndicator,
		//&gradientDescentIndicator,
	}
}

func setupSellIndicators(bot *Bot) {
	if ENABLE_FUTURES {
		leverageSellIndicator := NewLeverageSellIndicator(
			bot.Config,
			bot.buffer,
			bot.db,
		)

		bot.SellIndicators = []SellIndicator{
			//&highPercentageSellIndicator,
			&leverageSellIndicator,
			//&trailingSellIndicator,
		}
		return
	}

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

	trailingSellIndicator := NewTrailingSellIndicator(
		bot.Config,
		bot.buffer,
		bot.db,
	)

	bot.SellIndicators = []SellIndicator{
		//&highPercentageSellIndicator,
		&desiredPriceSellIndicator,
		//&trailingSellIndicator,
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
