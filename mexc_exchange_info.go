package main

func NewMexcExchangeInfo(res *MexcExchangeInfoResponse) FuturesExchangeInfo {
	infoMap := map[string]ExchangeInfoContainer{}

	for _, info := range res.Symbols {
		symbolInfoContainer := ExchangeInfoContainer{}

		symbolInfoContainer.LotSize = LotSize{
			minQty:   convertBinanceToFloat64(info.BaseSizePrecision),
			maxQty:   1000000000,
			stepSize: convertBinanceToFloat64(info.BaseSizePrecision),
		}

		symbolInfoContainer.PriceFilter = PriceFilter{
			minPrice: convertBinanceToFloat64(info.BaseSizePrecision),
			maxPrice: 1000000000,
			tickSize: 0.0001,
		}

		infoMap[info.Symbol] = symbolInfoContainer
	}

	return FuturesExchangeInfo{infoMap}
}
