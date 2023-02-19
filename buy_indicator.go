package main

import "fmt"

type BuyIndicator interface {
	HasSignal() bool
	IsStarted() bool
	Start()
	Update()
	Finish()
}

type BackTrailingBuyIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database

	isStarted                bool
	hasSignal                bool
	lastPrice                float64
	upperStopPrice           float64
	updatesCount             int
	updatedTimesBeforeFinish int
}

func NewBackTrailingBuyIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) BackTrailingBuyIndicator {
	return BackTrailingBuyIndicator{
		config: config,
		buffer: buffer,
		db:     db,

		isStarted:                false,
		hasSignal:                false,
		updatesCount:             0,
		updatedTimesBeforeFinish: 0,
	}
}

func (indicator *BackTrailingBuyIndicator) IsStarted() bool {
	return indicator.isStarted
}

func (indicator *BackTrailingBuyIndicator) Start() {
	LogAndPrint("Trailing Started")
	price := indicator.buffer.GetLastCandleClosePrice()
	indicator.upperStopPrice = indicator.calculateStopPrice(
		price,
		indicator.config.TrailingTopPercentage,
	)

	indicator.isStarted = true
	indicator.lastPrice = price
	indicator.hasSignal = false
	indicator.updatesCount = 0
	indicator.updatedTimesBeforeFinish = 0
	indicator.updatesCount = 0
}

func (indicator *BackTrailingBuyIndicator) Update() {
	if !indicator.IsStarted() {
		indicator.Start()
	}

	LogAndPrint("Trailing Updated")
	indicator.updatesCount++
	currentPrice := indicator.buffer.GetLastCandleClosePrice()
	if indicator.updatesCount <= 1 {
		return
	}

	if indicator.isGrowing() {
		indicator.hasSignal = indicator.upperStopPrice <= currentPrice
	}

	newUpperStopPrice := indicator.calculateStopPrice(
		currentPrice,
		indicator.config.TrailingTopPercentage,
	)

	if newUpperStopPrice < indicator.upperStopPrice {
		LogAndPrint(fmt.Sprintf("Trailing UpperStopPrice has moved to Price: %f", newUpperStopPrice))
		indicator.upperStopPrice = newUpperStopPrice
	}

	indicator.resolveSignal(currentPrice)
	indicator.lastPrice = currentPrice
}

func (indicator *BackTrailingBuyIndicator) Finish() {
	LogAndPrint("Trailing Finished")
	indicator.isStarted = false
	indicator.updatesCount = 0
	indicator.updatedTimesBeforeFinish = 0
}

func (indicator *BackTrailingBuyIndicator) HasSignal() bool {
	return indicator.hasSignal
}

func (indicator *BackTrailingBuyIndicator) isGrowing() bool {
	currentPrice := indicator.buffer.GetLastCandleClosePrice()

	return currentPrice > indicator.lastPrice
}

func (indicator *BackTrailingBuyIndicator) resolveSignal(currentPrice float64) {
	indicator.hasSignal = currentPrice >= indicator.upperStopPrice
	if indicator.hasSignal || indicator.updatedTimesBeforeFinish > 0 {
		indicator.hasSignal = true
		if indicator.updatedTimesBeforeFinish == indicator.config.TrailingUpdateTimesBeforeFinish {
			indicator.Finish()
			return
		}
		indicator.updatedTimesBeforeFinish++
	}
}

func (indicator *BackTrailingBuyIndicator) calculateStopPrice(closePrice, percentage float64) float64 {
	return closePrice + ((closePrice * percentage) / 100)
}

type BuysCountIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewBuysCountIndicator(
	config *Config,
	buffer *Buffer,
	db *Database,
) BuysCountIndicator {
	return BuysCountIndicator{
		config: config,
		buffer: buffer,
		db:     db,
	}
}

func (indicator *BuysCountIndicator) HasSignal() bool {
	return UNSOLD_BUYS_COUNT > indicator.db.CountUnsoldBuys()
}

func (indicator *BuysCountIndicator) IsStarted() bool {
	return true
}

func (indicator *BuysCountIndicator) Start() {
}

func (indicator *BuysCountIndicator) Update() {
}

func (indicator *BuysCountIndicator) Finish() {
}
