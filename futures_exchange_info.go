package main

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

type FuturesExchangeInfo struct {
	symbolInfoMap map[string]ExchangeInfoContainer
}

func NewFuturesExchangeInfo(res *futures.ExchangeInfo) FuturesExchangeInfo {
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

	r := infoMap["BTCUSDT"]
	fmt.Println(r)

	return FuturesExchangeInfo{infoMap}
}

func (info *FuturesExchangeInfo) GetInfoForSymbol(symbol string) (ExchangeInfoContainer, bool) {
	if container, ok := info.symbolInfoMap[symbol]; ok {
		return container, true
	}
	return ExchangeInfoContainer{}, false
}
