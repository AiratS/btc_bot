package main

type SellIndicator interface {
	HasSignal() (bool, []Buy)
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
