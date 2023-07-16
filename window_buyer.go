package main

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"math"
	"time"
)

type WindowBuyer struct {
	IsStarted      bool
	currentWindow  int
	lastWindowTime time.Time

	config              *Config
	buffer              *Buffer
	db                  *Database
	futuresOrderManager *FuturesOrderManager
}

func NewWindowBuyer(
	config *Config,
	buffer *Buffer,
	db *Database,
	futuresOrderManager *FuturesOrderManager,
) *WindowBuyer {
	return &WindowBuyer{
		IsStarted:     false,
		currentWindow: config.WindowWindowsCount,

		config:              config,
		buffer:              buffer,
		db:                  db,
		futuresOrderManager: futuresOrderManager,
	}
}

func (buyer *WindowBuyer) Start(usedMoney float64) {
	if buyer.IsStarted {
		return
	}
	buyer.IsStarted = true

	// Create order limit buy order
	buyer.createLimitBuyOrder(usedMoney)
}

func (buyer *WindowBuyer) Finish() {
	buyer.IsStarted = false
	buyer.currentWindow = buyer.config.WindowWindowsCount

	if hasValue, buyOrder := buyer.db.GetLastNewBuyOrder(); hasValue {
		buyer.db.UpdateBuyOrderStatus(buyOrder.Id, BuyOrderStatusFilled)
	}
}

func (buyer *WindowBuyer) MoveWindows(usedMoney float64) {
	if !buyer.IsStarted {
		return
	}

	// Move windows
	if buyer.currentWindow == 1 {
		return
	}

	// Check for period
	_, buyOrder := buyer.db.GetLastNewBuyOrder()
	if buyer.config.WindowWindowsCount == buyer.currentWindow {
		buyer.lastWindowTime = carbon.Parse(buyOrder.CreatedAt, carbon.Greenwich).ToStdTime()
	}

	currentCandle := buyer.buffer.GetLastCandle()
	currentPeriod := buyer.getCurrentWindowPeriod()
	diffInMinutes := carbon.FromStdTime(buyer.lastWindowTime).
		DiffInMinutes(carbon.Parse(currentCandle.CloseTime, carbon.Greenwich))

	if 0 > diffInMinutes {
		panic("Invalid minutes diff")
	}

	if currentPeriod > diffInMinutes {
		return
	}

	// Decrease window
	buyer.decreaseWindow()
	buyer.lastWindowTime = carbon.Parse(currentCandle.CloseTime, carbon.Greenwich).ToStdTime()

	// Create new buy order
	buyer.db.UpdateBuyOrderStatus(buyOrder.Id, BuyOrderStatusRejected)

	if IS_REAL_ENABLED && USE_REAL_MONEY {
		Log(fmt.Sprintf("CANCEL_LIMIT_BUY: orderId: %d", buyOrder.RealOrderId))
		buyer.futuresOrderManager.CancelOrder(CANDLE_SYMBOL, buyOrder.RealOrderId)
	}

	buyer.createLimitBuyOrder(usedMoney)
}

func (buyer *WindowBuyer) CheckForPercentage() bool {
	hasValue, buyOrder := buyer.db.GetLastNewBuyOrder()
	if !hasValue {
		return true
	}

	currentCandle := buyer.buffer.GetLastCandle()
	// Check for percentage
	percentage := CalcGrowth(buyOrder.ExchangeRate, currentCandle.GetPrice())
	if 0 <= percentage {
		return false
	}

	currentWindowPercentage := buyer.getCurrentWindowPercentage()

	return currentWindowPercentage <= math.Abs(percentage)
}

func (buyer *WindowBuyer) createLimitBuyOrder(usedMoney float64) {
	candle := buyer.buffer.GetLastCandle()
	var realOrderId int64 = 1
	buyPrice := CalcBottomPrice(candle.ClosePrice, buyer.getCurrentWindowPercentage())
	quantity := (usedMoney * float64(buyer.config.Leverage)) / buyPrice

	// Only for real money
	if IS_REAL_ENABLED && USE_REAL_MONEY {
		realOrderId, quantity, buyPrice = buyer.futuresOrderManager.
			CreateBuyOrder(CANDLE_SYMBOL, buyPrice, usedMoney)
	}

	buyer.db.AddNewBuyOrder(
		CANDLE_SYMBOL,
		usedMoney,
		quantity,
		buyPrice,
		buyPrice,
		realOrderId,
		candle.CloseTime,
	)
}

func (buyer *WindowBuyer) decreaseWindow() {
	currentCandle := buyer.buffer.GetLastCandle()
	Log(fmt.Sprintf(
		"WindowBuyer__DecreaseWindow\nCreatedAt: %s\nCurrentWindow: %d\nCurrentPercentage: %f",
		currentCandle.CloseTime,
		buyer.currentWindow,
		buyer.getCurrentWindowPercentage(),
	))

	buyer.currentWindow--

	if 1 > buyer.currentWindow {
		panic("Invalid currentWindow")
	}
}

func (buyer *WindowBuyer) getCurrentWindowPercentage() float64 {
	return buyer.config.WindowBasePercentage +
		buyer.config.WindowOffsetPercentage*(float64(buyer.currentWindow)-1)
}

func (buyer *WindowBuyer) getCurrentWindowPeriod() int64 {
	return int64(
		buyer.config.WindowBasePeriodMinutes +
			buyer.config.WindowOffsetPeriodMinutes*(buyer.currentWindow-1))
}
