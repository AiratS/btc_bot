package main

type BuyIndicator interface {
	HasSignal() bool
}

type BackTrailingBuyIndicator struct {
	config *Config
	db     *Database
}

func NewBackTrailingBuyIndicator() BackTrailingBuyIndicator {
	return BackTrailingBuyIndicator{}
}

func (indicator *BackTrailingBuyIndicator) HasSignal() bool {
	return false
}
