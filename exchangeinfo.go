package main

import "github.com/adshao/go-binance/v2"

type ExchangeInfo struct {
	symbolInfoMap map[string]ExchangeInfoContainer
}

type ExchangeInfoContainer struct {
	LotSize     LotSize
	PriceFilter PriceFilter
}

type LotSize struct {
	minQty   float64
	maxQty   float64
	stepSize float64
}

type PriceFilter struct {
	minPrice float64
	maxPrice float64
	tickSize float64
}

func NewExchangeInfo(res *binance.ExchangeInfo) ExchangeInfo {
	infoMap := map[string]ExchangeInfoContainer{}

	for _, info := range res.Symbols {
		symbolInfoContainer := ExchangeInfoContainer{}

		for _, filter := range info.Filters {
			filterType := filter["filterType"]

			if filterType == string(binance.SymbolFilterTypeLotSize) {
				symbolInfoContainer.LotSize = LotSize{
					minQty:   convertToFloat64(filter["minQty"]),
					maxQty:   convertToFloat64(filter["maxQty"]),
					stepSize: convertToFloat64(filter["stepSize"]),
				}
			}

			if filterType == string(binance.SymbolFilterTypePriceFilter) {
				symbolInfoContainer.PriceFilter = PriceFilter{
					minPrice: convertToFloat64(filter["minPrice"]),
					maxPrice: convertToFloat64(filter["maxPrice"]),
					tickSize: convertToFloat64(filter["tickSize"]),
				}
			}
		}

		infoMap[info.Symbol] = symbolInfoContainer
	}

	return ExchangeInfo{infoMap}
}

func (info *ExchangeInfo) GetInfoForSymbol(symbol string) (ExchangeInfoContainer, bool) {
	if container, ok := info.symbolInfoMap[symbol]; ok {
		return container, true
	}
	return ExchangeInfoContainer{}, false
}
