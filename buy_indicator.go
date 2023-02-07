package main

type BuyIndicator interface {
	HasSignal() bool
}

type BackTrailingBuyIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database

	lastPrice      float64
	upperStopPrice float64
	hasSignal      bool
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
	}
}

func (indicator *BackTrailingBuyIndicator) Start() {
	price := indicator.buffer.GetLastCandleClosePrice()
	indicator.upperStopPrice = indicator.calculateStopPrice(
		price,
		indicator.config.TrailingTopPercentage,
	)
	indicator.lastPrice = price
}

func (indicator *BackTrailingBuyIndicator) Update() {
	currentPrice := indicator.buffer.GetLastCandleClosePrice()
	isGrowing := indicator.isGrowing()
	indicator.lastPrice = currentPrice

	if isGrowing {
		if indicator.upperStopPrice <= currentPrice {
			indicator.hasSignal = true
		}
		return
	}

	newUpperStopPrice := indicator.calculateStopPrice(
		currentPrice,
		indicator.config.TrailingTopPercentage,
	)

	if newUpperStopPrice < indicator.upperStopPrice {
		indicator.upperStopPrice = newUpperStopPrice
	}
}

func (indicator *BackTrailingBuyIndicator) Finish() {
	indicator.hasSignal = false
}

func (indicator *BackTrailingBuyIndicator) HasSignal() bool {
	return indicator.hasSignal
}

func (indicator *BackTrailingBuyIndicator) isGrowing() bool {
	currentPrice := indicator.buffer.GetLastCandleClosePrice()

	return currentPrice > indicator.lastPrice
}

func (indicator *BackTrailingBuyIndicator) calculateStopPrice(closePrice, percentage float64) float64 {
	return closePrice - ((closePrice * percentage) / 100)
}
