package main

type SellIndicator interface {
	HasSignal() (bool, []Buy)
}

type HighPercentageSellIndicator struct {
	config *Config
	buffer *Buffer
	db     *Database
}

func NewHighPercentageSellIndicator() HighPercentageSellIndicator {
	return HighPercentageSellIndicator{}
}

func (indicator *HighPercentageSellIndicator) HasSignal() (bool, []Buy) {
	return false, []Buy{}
}
