package main

import "fmt"

type Balance struct {
	config         Config
	inBalanceMoney float64
	buysCount      int64
}

func NewBalance(config Config) Balance {
	return Balance{
		config:         config,
		inBalanceMoney: BALANCE_MONEY,
		buysCount:      0,
	}
}

func (balance *Balance) HasEnoughMoneyForBuy() bool {
	return balance.inBalanceMoney >= balance.config.TotalMoneyAmount
}

func (balance *Balance) buy() {
	balance.inBalanceMoney -= balance.config.TotalMoneyAmount
	if balance.inBalanceMoney < 0 {
		balance.inBalanceMoney = 0
	}

	balance.buysCount++

	Log(fmt.Sprintf(
		"Balance__BUY\nBuysCount: %d\nInBalanceMoney:%f",
		balance.buysCount,
		balance.inBalanceMoney,
	))
}

func (balance *Balance) sell() {
	balance.buysCount--
	if balance.buysCount < 0 {
		balance.buysCount = 0
	}

	balance.inBalanceMoney += balance.config.TotalMoneyAmount
	if balance.inBalanceMoney > BALANCE_MONEY {
		balance.inBalanceMoney = BALANCE_MONEY
	}

	Log(fmt.Sprintf(
		"Balance__SELL\nBuysCount: %d\nInBalanceMoney:%f",
		balance.buysCount,
		balance.inBalanceMoney,
	))
}
