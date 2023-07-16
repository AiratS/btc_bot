package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
	"time"
)

const LEVERAGE = 10
const MARGIN_TYPE = "isolated"
const RETRIES_COUNT = 5
const RETRY_DELAY = 300

type FuturesOrderManager struct {
	futuresClient *futures.Client
	isEnabled     bool
	exchangeInfo  *FuturesExchangeInfo
}

type PositionInfo struct {
	MarginType string
	Leverage   int
}

func NewFuturesOrderManager(futuresClient *futures.Client) FuturesOrderManager {
	service := FuturesOrderManager{
		futuresClient: futuresClient,
		isEnabled:     USE_REAL_MONEY,
	}

	if USE_REAL_MONEY {
		res, err := futuresClient.NewExchangeInfoService().
			Do(context.Background())

		if err != nil {
			panic(err)
		}

		balance := service.getBalance()
		if 10.0 >= balance {
			panic(fmt.Sprintf("Not enough money in balance: %f", balance))
		}

		info := NewFuturesExchangeInfo(res)
		service.exchangeInfo = &info

		service.adjustConfiguration()
	}

	return service
}

func (manager *FuturesOrderManager) adjustConfiguration() {
	info := manager.getPositionInfo()

	if info.MarginType != MARGIN_TYPE {
		err := manager.futuresClient.NewChangeMarginTypeService().
			Symbol(CANDLE_SYMBOL).
			MarginType(futures.MarginTypeIsolated).
			Do(context.Background())

		if err != nil {
			panic(err)
		}
	}

	if info.Leverage != LEVERAGE {
		res, err := manager.futuresClient.NewChangeLeverageService().
			Symbol(CANDLE_SYMBOL).
			Leverage(LEVERAGE).
			Do(context.Background())

		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}

	// Update position mode if required
	if ENABLE_SHORT && !manager.isDualSidePositionMode() {
		manager.changePositionMode(true)
	} else if !ENABLE_SHORT && manager.isDualSidePositionMode() {
		manager.changePositionMode(false)
	}
}

func (manager *FuturesOrderManager) getPositionInfo() PositionInfo {
	res, err := manager.futuresClient.NewGetPositionRiskService().
		Symbol(CANDLE_SYMBOL).
		Do(context.Background())

	if err != nil {
		panic(err)
	}

	for _, info := range res {
		if info.Symbol == CANDLE_SYMBOL {
			return PositionInfo{
				MarginType: info.MarginType,
				Leverage:   convertBinanceToInt(info.Leverage),
			}
		}
	}

	panic("No futures position info.")
}

func (manager *FuturesOrderManager) isDualSidePositionMode() bool {
	res, err := manager.futuresClient.NewGetPositionModeService().
		Do(context.Background())

	if err != nil {
		panic(err)
	}

	return res.DualSidePosition
}

func (manager *FuturesOrderManager) changePositionMode(isDualSide bool) {
	err := manager.futuresClient.NewChangePositionModeService().
		DualSide(isDualSide).
		Do(context.Background())

	if err != nil {
		panic(err)
	}
}

func (manager *FuturesOrderManager) CanBuyForPrice(symbol string, price, usedMoney float64) bool {
	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney(usedMoney)), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		return info.LotSize.minQty <= quantityLotSize && quantityLotSize <= info.LotSize.maxQty &&
			info.PriceFilter.minPrice <= priceConverted && priceConverted <= info.PriceFilter.maxPrice
	}

	return false
}

func (manager *FuturesOrderManager) HasEnoughMoneyForBuy(usedMoney float64) bool {
	var errorMessage string

	for i := 0; i < RETRIES_COUNT; i++ {
		res, err := manager.futuresClient.NewGetBalanceService().
			Do(context.Background())

		if err != nil {
			errorMessage = err.Error()
			Log(errorMessage)
			time.Sleep(RETRY_DELAY * time.Millisecond)
			continue
		}

		for _, balance := range res {
			if balance.Asset == "BUSD" {
				freeMoney := convertBinanceToFloat64(balance.AvailableBalance)

				return freeMoney >= usedMoney
			}
		}

		return false
	}

	panic(errorMessage)
}

func (manager *FuturesOrderManager) getBalance() float64 {
	var errorMessage string

	for i := 0; i < RETRIES_COUNT; i++ {
		res, err := manager.futuresClient.NewGetBalanceService().
			Do(context.Background())

		if err != nil {
			errorMessage = err.Error()
			Log(errorMessage)
			time.Sleep(RETRY_DELAY * time.Millisecond)
			continue
		}

		for _, balance := range res {
			if balance.Asset == "BUSD" {
				return convertBinanceToFloat64(balance.AvailableBalance)
			}
		}

		return 0
	}

	panic(errorMessage)
}

func (manager *FuturesOrderManager) IsBuySold(symbol string, orderId int64) bool {
	var errorMessage string

	for i := 0; i < RETRIES_COUNT; i++ {
		res, err := manager.futuresClient.NewGetOrderService().
			OrderID(orderId).
			Symbol(symbol).
			Do(context.Background())

		if err != nil {
			errorMessage = err.Error()
			Log(errorMessage)
			time.Sleep(RETRY_DELAY * time.Millisecond)
			continue
		}

		fmt.Println(res.Status)

		return res.Status == "FILLED"
	}

	panic(errorMessage)
}

func (manager *FuturesOrderManager) GetOrderInfo(orderId int64) *futures.Order {
	var errorMessage string

	for i := 0; i < RETRIES_COUNT; i++ {
		res, err := manager.futuresClient.NewGetOrderService().
			OrderID(orderId).
			Symbol(CANDLE_SYMBOL).
			Do(context.Background())

		if err != nil {
			errorMessage = err.Error()
			Log(errorMessage)
			time.Sleep(RETRY_DELAY * time.Millisecond)
			continue
		}

		fmt.Println(res.Status)

		return res
	}

	panic(errorMessage)
}

func (manager *FuturesOrderManager) CreateBuyOrder(symbol string, price, usedMoney float64) (int64, float64, float64) {
	if !manager.isEnabled {
		return 0, 0.0, price
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney(usedMoney)), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateLimitBuyOrder: %f, %f", priceConverted, quantityLotSize))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := manager.futuresClient.
				NewCreateOrderService().
				Symbol(symbol).
				Side(futures.SideTypeBuy).
				Type(futures.OrderTypeLimit).
				//PositionSide(futures.PositionSideTypeLong).
				TimeInForce(futures.TimeInForceTypeGTC).
				Quantity(floatToBinancePrice(quantityLotSize)).
				Price(floatToBinancePrice(priceConverted)).
				Do(context.Background())

			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			//realBuyPrice, realQuantity := manager.getRealBuyPriceAndQuantity(priceConverted, quantityLotSize, order)

			return order.OrderID, quantityLotSize, priceConverted
		}

		panic(errorMessage)
	}

	return 0, 0.0, 0.0
}

func (manager *FuturesOrderManager) CreateShortLimitBuyOrder(symbol string, price, quantity float64) int64 {
	if !manager.isEnabled {
		return 0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(quantity, info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateLimitBuyOrder: %f, %f", priceConverted, quantityLotSize))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := manager.futuresClient.
				NewCreateOrderService().
				Symbol(symbol).
				Side(futures.SideTypeBuy).
				Type(futures.OrderTypeLimit).
				PositionSide(futures.PositionSideTypeShort).
				TimeInForce(futures.TimeInForceTypeGTC).
				Quantity(floatToBinancePrice(quantityLotSize)).
				Price(floatToBinancePrice(priceConverted)).
				Do(context.Background())

			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			//realBuyPrice := manager.getRealBuyPrice(priceConverted, order)
			//realQuantity := manager.getRealBuyQuantity(quantity, order)

			return order.OrderID
		}

		panic(errorMessage)
	}

	return 0
}

func (manager *FuturesOrderManager) CreateMarketBuyOrder(symbol string, price, usedMoney float64) (int64, float64, float64) {
	if !manager.isEnabled {
		return 0, 0.0, price
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney(usedMoney)), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateBuyOrder: %f, %f", priceConverted, quantityLotSize))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := manager.futuresClient.
				NewCreateOrderService().
				Symbol(symbol).
				Side(futures.SideTypeBuy).
				Type(futures.OrderTypeMarket).
				//PositionSide(futures.PositionSideTypeLong).
				Quantity(floatToBinancePrice(quantityLotSize)).
				Do(context.Background())

			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			realBuyPrice, realQuantity := manager.getRealBuyPriceAndQuantity(priceConverted, quantityLotSize, order)

			return order.OrderID, realQuantity, realBuyPrice
		}

		panic(errorMessage)
	}

	return 0, 0.0, 0.0
}

func (manager *FuturesOrderManager) CreateShortMarketBuyOrder(symbol string, price, usedMoney float64) (int64, float64, float64) {
	if !manager.isEnabled {
		return 0, 0.0, price
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney(usedMoney)), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateBuyOrder: %f, %f", priceConverted, quantityLotSize))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := manager.futuresClient.
				NewCreateOrderService().
				Symbol(symbol).
				Side(futures.SideTypeBuy).
				Type(futures.OrderTypeMarket).
				PositionSide(futures.PositionSideTypeShort).
				Quantity(floatToBinancePrice(quantityLotSize)).
				Do(context.Background())

			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			realBuyPrice, realQuantity := manager.getRealBuyPriceAndQuantity(priceConverted, quantityLotSize, order)

			return order.OrderID, realQuantity, realBuyPrice
		}

		panic(errorMessage)
	}

	return 0, 0.0, 0.0
}

func (manager *FuturesOrderManager) CreateShortMarketSellOrder(
	symbol string,
	stopPrice,
	quantity float64,
) (int64, float64, float64) {
	if !manager.isEnabled {
		return 0, 0.0, 0.0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(quantity, info.LotSize.stepSize)
		priceConverted := valueToPriceSize(stopPrice, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateSellOrder: %f, %f, %f", priceConverted, stopPrice, quantityLotSize))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := manager.futuresClient.
				NewCreateOrderService().
				Symbol(symbol).
				Side(futures.SideTypeSell).
				Type(futures.OrderTypeMarket).
				PositionSide(futures.PositionSideTypeShort).
				Quantity(floatToBinancePrice(quantityLotSize)).
				Do(context.Background())

			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			return order.OrderID,
				convertBinanceToFloat64(order.OrigQuantity),
				priceConverted
		}

		panic(errorMessage)
	}

	return 0, 0.0, 0.0
}

func (manager *FuturesOrderManager) CreateSellOrder(symbol string, stopPrice, quantity float64) int64 {
	if !manager.isEnabled {
		return 0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		priceConverted := valueToPriceSize(stopPrice, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateSellOrder: %f, %f, %f", priceConverted, stopPrice, quantity))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := manager.futuresClient.
				NewCreateOrderService().
				Symbol(symbol).
				Side(futures.SideTypeSell).
				Type(futures.OrderTypeLimit).
				//PositionSide(futures.PositionSideTypeLong).
				TimeInForce(futures.TimeInForceTypeGTC).
				Quantity(floatToBinancePrice(quantity)).
				Price(floatToBinancePrice(priceConverted)).
				Do(context.Background())

			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			return order.OrderID
		}

		panic(errorMessage)
	}

	return 0
}

func (manager *FuturesOrderManager) CreateShortSellOrder(symbol string, stopPrice, quantity float64) int64 {
	if !manager.isEnabled {
		return 0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		priceConverted := valueToPriceSize(stopPrice, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateSellOrder: %f, %f, %f", priceConverted, stopPrice, quantity))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := manager.futuresClient.
				NewCreateOrderService().
				Symbol(symbol).
				Side(futures.SideTypeSell).
				Type(futures.OrderTypeLimit).
				PositionSide(futures.PositionSideTypeShort).
				TimeInForce(futures.TimeInForceTypeGTC).
				Quantity(floatToBinancePrice(quantity)).
				Price(floatToBinancePrice(priceConverted)).
				Do(context.Background())

			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			return order.OrderID
		}

		panic(errorMessage)
	}

	return 0
}

func (manager *FuturesOrderManager) CancelOrder(symbol string, orderId int64) int64 {
	if !manager.isEnabled {
		return 0
	}

	var errorMessage string
	for i := 0; i < RETRIES_COUNT; i++ {
		order, err := manager.futuresClient.
			NewCancelOrderService().
			Symbol(symbol).
			OrderID(orderId).
			Do(context.Background())

		if err != nil {
			errorMessage = err.Error()
			Log(errorMessage)
			time.Sleep(RETRY_DELAY * time.Millisecond)
			continue
		}

		return order.OrderID
	}

	panic(errorMessage)
}

func (manager *FuturesOrderManager) getOrderMoney(usedMoney float64) float64 {
	return usedMoney * LEVERAGE
}

func (manager *FuturesOrderManager) getRealBuyPriceAndQuantity(rawPrice, rawQuantity float64, order *futures.CreateOrderResponse) (float64, float64) {
	realBuyPrice := rawPrice
	realQuantity := rawQuantity

	for j := 0; j < 10; j++ {
		time.Sleep(RETRY_DELAY * time.Millisecond)

		orderInfo := manager.GetOrderInfo(order.OrderID)
		realBuyPrice = convertBinanceToFloat64(orderInfo.AvgPrice)
		realQuantity = convertBinanceToFloat64(orderInfo.ExecutedQuantity)

		Log(fmt.Sprintf("realBuyPrice: %f, ExecutedQuantity: %f", realBuyPrice, realQuantity))

		if realBuyPrice != 0 && realQuantity != 0 {
			break
		}
	}

	return realBuyPrice, realQuantity
}

func (manager *FuturesOrderManager) getRealBuyQuantity(rawQuantity float64, order *futures.CreateOrderResponse) float64 {
	return convertBinanceToFloat64(order.ExecutedQuantity)
	//if 0 == len(order.Fills) {
	//	return rawQuantity
	//}
	//
	//realQuantity := 0.0
	//for _, fill := range order.Fills {
	//	realQuantity += convertBinanceToFloat64(fill.Quantity)
	//}
	//
	//return realQuantity
}
