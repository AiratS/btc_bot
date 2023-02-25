package main

import (
	"context"
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	"math"
)

const ORDER_MONEY = 13

type OrderManager struct {
	binanceClient *binance.Client
	isEnabled     bool
	exchangeInfo  *ExchangeInfo
}

func NewOrderManager(binanceClient *binance.Client, isEnabled bool) OrderManager {
	res, err := binanceClient.NewExchangeInfoService().
		Symbols("BTCUSDT").
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	info := NewExchangeInfo(res)

	return OrderManager{
		binanceClient: binanceClient,
		isEnabled:     false,
		exchangeInfo:  &info,
	}
}

func (manager *OrderManager) CanBuyForPrice(symbol string, price float64) bool {
	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, ORDER_MONEY), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		return info.LotSize.minQty <= quantityLotSize && quantityLotSize <= info.LotSize.maxQty &&
			info.PriceFilter.minPrice <= priceConverted && priceConverted <= info.PriceFilter.maxPrice
	}

	return false
}

func (manager *OrderManager) IsBuySold(symbol string, orderId int64) bool {
	res, err := manager.binanceClient.NewGetOrderService().
		OrderID(orderId).
		Symbol(symbol).
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(res.Status)

	return res.Status == "FILLED"
}

func (manager *OrderManager) GetSoldPrice(symbol string, orderId int64) (float64, bool) {
	res, err := manager.binanceClient.NewGetOrderService().
		OrderID(orderId).
		Symbol(symbol).
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if res.Status == "FILLED" {
		return 0.0, false
	}

	price := convertStringToFloat64(res.Price)

	return price, true
}

func (manager *OrderManager) HasEnoughMoneyForBuy() bool {
	res, err := manager.binanceClient.NewGetAccountService().
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		return false
	}

	for _, balance := range res.Balances {
		if balance.Asset == "USDT" {
			freeMoney := convertToFloat64(balance.Free)

			return freeMoney >= ORDER_MONEY
		}
	}

	return false
}

func (manager *OrderManager) CreateBuyOrder(symbol string, price float64) (int64, float64) {
	if !manager.isEnabled {
		return 0, 0.0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, ORDER_MONEY), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateBuyOrder: %f, %f", priceConverted, quantityLotSize))

		order, err := manager.binanceClient.
			NewCreateOrderService().
			Symbol(symbol).
			Side(binance.SideTypeBuy).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(floatToBinancePrice(quantityLotSize)).
			Price(floatToBinancePrice(priceConverted)).
			Do(context.Background())

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		//fmt.Println(order)
		//fmt.Println("orderId", order.OrderID)

		return order.OrderID, quantityLotSize
	}

	return 0, 0.0
}

func (manager *OrderManager) CreateMarketBuyOrder(symbol string, price float64) (int64, float64) {
	if !manager.isEnabled {
		return 0, 0.0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, ORDER_MONEY), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateBuyOrder: %f, %f", priceConverted, quantityLotSize))

		order, err := manager.binanceClient.
			NewCreateOrderService().
			Symbol(symbol).
			Side(binance.SideTypeBuy).
			Type(binance.OrderTypeMarket).
			Quantity(floatToBinancePrice(quantityLotSize)).
			Do(context.Background())

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		return order.OrderID, quantityLotSize
	}

	return 0, 0.0
}

func (manager *OrderManager) MoveStopPrice(symbol string, orderId int64, stopPrice, quantity float64) int64 {
	if !manager.isEnabled {
		return 0
	}

	manager.CancelOrder(symbol, orderId)

	return manager.CreateSellOrder(symbol, stopPrice, quantity)
}

func (manager *OrderManager) CreateSellOrder(symbol string, stopPrice, quantity float64) int64 {
	if !manager.isEnabled {
		return 0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		priceConverted := valueToPriceSize(stopPrice, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateSellOrder: %f, %f, %f", priceConverted, stopPrice, quantity))

		order, err := manager.binanceClient.
			NewCreateOrderService().
			Symbol(symbol).
			Side(binance.SideTypeSell).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(floatToBinancePrice(quantity)).
			Price(floatToBinancePrice(priceConverted)).
			Do(context.Background())

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		return order.OrderID
	}

	return 0
}

func (manager *OrderManager) CreateMarketSellOrder(symbol string, stopPrice, quantity float64) int64 {
	if !manager.isEnabled {
		return 0
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		priceConverted := valueToPriceSize(stopPrice, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateSellOrder: %f, %f, %f", priceConverted, stopPrice, quantity))

		order, err := manager.binanceClient.
			NewCreateOrderService().
			Symbol(symbol).
			Side(binance.SideTypeSell).
			Type(binance.OrderTypeMarket).
			Quantity(floatToBinancePrice(quantity)).
			Do(context.Background())

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		return order.OrderID
	}

	return 0
}

func (manager *OrderManager) CancelOrder(symbol string, orderId int64) int64 {
	if !manager.isEnabled {
		return 0
	}

	order, err := manager.binanceClient.
		NewCancelOrderService().
		Symbol(symbol).
		OrderID(orderId).
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	//fmt.Println(order)
	//fmt.Println("orderId", order.OrderID)

	return order.OrderID
}

func floatToBinancePrice(price float64) string {
	return fmt.Sprintf("%.6f", price)
}

func calcQuantity(exchangeRate, orderMoney float64) float64 {
	return orderMoney / exchangeRate
}

func valueToLotSize(value, step float64) float64 {
	return math.Round(value/step) * step
}

func valueToPriceSize(value, tickSize float64) float64 {
	return math.Round(value/tickSize) * tickSize
}
