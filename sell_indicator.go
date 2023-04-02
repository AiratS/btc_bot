package main

import "fmt"

type SellIndicator interface {
	HasSignal() (bool, []Buy)
	RunAfterBuy(buyId int64)
	Update()
	Finish(buyId int64)
}

type UpdatableIndicator interface {
}

type HighPercentageSellIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewHighPercentageSellIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) HighPercentageSellIndicator {
	return HighPercentageSellIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *HighPercentageSellIndicator) HasSignal() (bool, []Buy) {
	currentPrice := indicator.buffer.GetLastCandleClosePrice()

	buys := indicator.db.FetchUnsoldBuysByUpperPercentage(
		currentPrice,
		indicator.config.HighSellPercentage,
	)

	return len(buys) > 0, buys
}

func (indicator *HighPercentageSellIndicator) RunAfterBuy(buyId int64) {
}

func (indicator *HighPercentageSellIndicator) Update() {
}

func (indicator *HighPercentageSellIndicator) Finish(buyId int64) {
}

// ------------------------------------

type DesiredPriceSellIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewDesiredPriceSellIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) DesiredPriceSellIndicator {
	return DesiredPriceSellIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *DesiredPriceSellIndicator) HasSignal() (bool, []Buy) {
	candle := indicator.buffer.GetLastCandle()
	maxPrice := Max([]float64{candle.ClosePrice, candle.HighPrice})
	buys := indicator.db.FetchUnsoldBuysByDesiredPrice(maxPrice)

	return len(buys) > 0, buys
}

func (indicator *DesiredPriceSellIndicator) RunAfterBuy(buyId int64) {
}

func (indicator *DesiredPriceSellIndicator) Update() {
}

func (indicator *DesiredPriceSellIndicator) Finish(buyId int64) {
}

// ------------------------------------

type TrailingBuy struct {
	isActivated bool
	buyPrice    float64
	stopPrice   float64
}

type TrailingSellIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
	buys   map[int64]*TrailingBuy
}

func NewTrailingSellIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) TrailingSellIndicator {
	return TrailingSellIndicator{
		config: config,
		buffer: buffer,
		db:     db,
		buys:   map[int64]*TrailingBuy{},
	}
}

func (indicator *TrailingSellIndicator) RunAfterBuy(buyId int64) {
	if _, ok := indicator.buys[buyId]; !ok {
		currentPrice := indicator.buffer.GetLastCandleClosePrice()
		indicator.buys[buyId] = &TrailingBuy{
			isActivated: false,
			buyPrice:    currentPrice,
		}

		Log(fmt.Sprintf("TrailingSellIndicator__STARTED\nBuyId: %d", buyId))
	}
}

func (indicator *TrailingSellIndicator) Update() {
	for _, unsoldBuy := range indicator.db.FetchUnsoldBuys() {
		indicator.updateByBuyId(unsoldBuy.Id)
	}
}

func (indicator *TrailingSellIndicator) updateByBuyId(buyId int64) {
	if buyItem, ok := indicator.buys[buyId]; ok {
		currentPrice := indicator.buffer.GetLastCandleClosePrice()

		if buyItem.isActivated {
			newStopPrice := CalcBottomPrice(currentPrice, indicator.config.TrailingSellStopPercentage)

			if buyItem.stopPrice < newStopPrice {
				buyItem.stopPrice = newStopPrice
				Log(fmt.Sprintf(
					"TrailingSellIndicator__STOP_MOVED\nBuyId: %d\nStopPrice: %f",
					buyId,
					newStopPrice,
				))

				currentCandle := indicator.buffer.GetLastCandle()
				PlotAddTrailingSellPoint(buyId, currentCandle.CloseTime, newStopPrice)
			}

			return
		}

		// Check for activation condition
		activationPercentage := indicator.config.HighSellPercentage + indicator.config.TrailingSellActivationAdditionPercentage
		activationPrice := CalcUpperPrice(buyItem.buyPrice, activationPercentage)
		if !buyItem.isActivated && activationPrice <= currentPrice {
			indicator.activate(buyId)
			Log(fmt.Sprintf("TrailingSellIndicator__ACTIVATED\nBuyId: %d", buyId))
		}
	}
}

func (indicator *TrailingSellIndicator) activate(buyId int64) {
	if buyItem, ok := indicator.buys[buyId]; ok {
		buyItem.isActivated = true
		buyItem.stopPrice = CalcUpperPrice(
			indicator.buys[buyId].buyPrice,
			indicator.config.HighSellPercentage,
		)

		currentCandle := indicator.buffer.GetLastCandle()
		PlotAddTrailingSellPoint(buyId, currentCandle.CloseTime, buyItem.stopPrice)
	}
}

func (indicator *TrailingSellIndicator) Finish(buyId int64) {
	if _, ok := indicator.buys[buyId]; ok {
		delete(indicator.buys, buyId)
		Log(fmt.Sprintf("TrailingSellIndicator__FINISHED\nBuyId: %d", buyId))
	}
}

func (indicator *TrailingSellIndicator) HasSignal() (bool, []Buy) {
	var unsoldBuyIds []int64
	currentPrice := indicator.buffer.GetLastCandleClosePrice()

	for buyId, buyItem := range indicator.buys {
		if buyItem.isActivated && buyItem.stopPrice >= currentPrice {
			unsoldBuyIds = append(unsoldBuyIds, buyId)
		}
	}

	if 0 == len(unsoldBuyIds) {
		return false, []Buy{}
	}

	buys := indicator.db.FetchUnsoldBuysById(unsoldBuyIds)
	return len(buys) > 0, buys
}

func (indicator *TrailingSellIndicator) GetBuyItemByBuyId(buyId int64) (bool, *TrailingBuy) {
	if buyItem, ok := indicator.buys[buyId]; ok {
		return true, buyItem
	}

	return false, &TrailingBuy{}
}

func (indicator *TrailingSellIndicator) calculateStopPrice(closePrice, percentage float64) float64 {
	return closePrice - ((closePrice * percentage) / 100)
}

// ------------------------------------

type LeverageSellIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewLeverageSellIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) LeverageSellIndicator {
	return LeverageSellIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *LeverageSellIndicator) HasSignal() (bool, []Buy) {
	var resultingBuys []Buy
	candle := indicator.buffer.GetLastCandle()

	// Upper buys
	upperBuys := indicator.db.FetchUnsoldBuysByUpperPercentage(
		candle.GetPrice(),
		indicator.config.HighSellPercentage,
	)
	indicator.appendBuyIfNotExists(&resultingBuys, upperBuys)

	// Liquidation buys
	liquidationBuys := indicator.db.FetchUnsoldBuysByLowerPercentage(
		candle.LowPrice,
		GetLeverageLiquidationPercentage(indicator.config.Leverage),
	)
	indicator.appendBuyIfNotExists(&resultingBuys, liquidationBuys)

	// Time cancel buys
	timeCancelBuys := indicator.db.FetchTimeCancelBuys(
		candle.CloseTime,
		indicator.config.FuturesAvgSellTimeMinutes,
	)
	indicator.appendBuyIfNotExists(&resultingBuys, timeCancelBuys)

	// Mark timeCancel
	indicator.markBuys(&resultingBuys, timeCancelBuys, TimeCancel)

	// Mark liquidation buys
	indicator.markBuys(&resultingBuys, liquidationBuys, Liquidation)

	return len(resultingBuys) > 0, resultingBuys
}

func (indicator *LeverageSellIndicator) appendBuyIfNotExists(saveList *[]Buy, newList []Buy) {
	for _, buy := range newList {
		if !indicator.hasSuchBuy(buy.Id, *saveList) {
			*saveList = append(*saveList, buy)
		}
	}
}

func (indicator *LeverageSellIndicator) markBuys(
	targetList *[]Buy,
	markList []Buy,
	buyType BuyType,
) {
	for idx, targetBuy := range *targetList {
		if indicator.hasSuchBuy(targetBuy.Id, markList) {
			(*targetList)[idx].BuyType = buyType
		}
	}

	fmt.Println()
}

func (indicator *LeverageSellIndicator) hasSuchBuy(buyId int64, buys []Buy) bool {
	for _, buy := range buys {
		if buy.Id == buyId {
			return true
		}
	}

	return false
}

func (indicator *LeverageSellIndicator) RunAfterBuy(buyId int64) {
}

func (indicator *LeverageSellIndicator) Update() {
}

func (indicator *LeverageSellIndicator) Finish(buyId int64) {
}
