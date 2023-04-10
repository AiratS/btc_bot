package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2/futures"
)

const LEVERAGE = 10
const MARGIN_TYPE = "isolated"

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

func (manager *FuturesOrderManager) CanBuyForPrice(symbol string, price float64) bool {
	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney()), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		return info.LotSize.minQty <= quantityLotSize && quantityLotSize <= info.LotSize.maxQty &&
			info.PriceFilter.minPrice <= priceConverted && priceConverted <= info.PriceFilter.maxPrice
	}

	return false
}

func (manager *FuturesOrderManager) HasEnoughMoneyForBuy() bool {
	res, err := manager.futuresClient.NewGetBalanceService().
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		return false
	}

	for _, balance := range res {
		if balance.Asset == "BUSD" {
			freeMoney := convertBinanceToFloat64(balance.Balance)

			return freeMoney >= ORDER_MONEY
		}
	}

	return false
}

func (manager *FuturesOrderManager) IsBuySold(symbol string, orderId int64) bool {
	res, err := manager.futuresClient.NewGetOrderService().
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

func (manager *FuturesOrderManager) CreateMarketBuyOrder(symbol string, price float64) (int64, float64, float64) {
	if !manager.isEnabled {
		return 0, 0.0, price
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney()), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateBuyOrder: %f, %f", priceConverted, quantityLotSize))

		order, err := manager.futuresClient.
			NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideTypeBuy).
			Type(futures.OrderTypeMarket).
			Quantity(floatToBinancePrice(quantityLotSize)).
			Do(context.Background())

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		realBuyPrice := manager.getRealBuyPrice(priceConverted, order)
		realQuantity := manager.getRealBuyQuantity(quantityLotSize, order)

		return order.OrderID, realQuantity, realBuyPrice
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

		order, err := manager.futuresClient.
			NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideTypeSell).
			Type(futures.OrderTypeLimit).
			PositionSide(futures.PositionSideTypeLong).
			TimeInForce(futures.TimeInForceTypeGTC).
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

func (manager *FuturesOrderManager) CancelOrder(symbol string, orderId int64) int64 {
	if !manager.isEnabled {
		return 0
	}

	order, err := manager.futuresClient.
		NewCancelOrderService().
		Symbol(symbol).
		OrderID(orderId).
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return order.OrderID
}

func (manager *FuturesOrderManager) getOrderMoney() float64 {
	return ORDER_MONEY * LEVERAGE
}

func (manager *FuturesOrderManager) getRealBuyPrice(rawPrice float64, order *futures.CreateOrderResponse) float64 {
	//if len(order.Fills) > 0 {
	//	return convertBinanceToFloat64(order.Fills[len(order.Fills)-1].Price)
	//}

	return rawPrice
}

func (manager *FuturesOrderManager) getRealBuyQuantity(rawQuantity float64, order *futures.CreateOrderResponse) float64 {
	return rawQuantity
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
