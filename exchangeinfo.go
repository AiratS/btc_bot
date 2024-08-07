package main

import (
	"github.com/adshao/go-binance/v2"
)

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
				//fmt.Println(filter)
				symbolInfoContainer.LotSize = LotSize{
					minQty:   convertBinanceToFloat64(filter["minQty"]),
					maxQty:   convertBinanceToFloat64(filter["maxQty"]),
					stepSize: convertBinanceToFloat64(filter["stepSize"]),
				}
			}

			if filterType == string(binance.SymbolFilterTypePriceFilter) {
				//fmt.Println(convertBinanceToFloat64(filter["minPrice"]))
				symbolInfoContainer.PriceFilter = PriceFilter{
					minPrice: convertBinanceToFloat64(filter["minPrice"]),
					maxPrice: convertBinanceToFloat64(filter["maxPrice"]),
					tickSize: convertBinanceToFloat64(filter["tickSize"]),
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
