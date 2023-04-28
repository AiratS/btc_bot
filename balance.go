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

func (balance *Balance) HasEnoughMoneyForBuy(requiredMoney float64) bool {
	return balance.inBalanceMoney >= requiredMoney
}

func (balance *Balance) buy(usedMoney float64) {
	balance.inBalanceMoney -= usedMoney
	if balance.inBalanceMoney < 0 {
		panic("Balance is minus")
		balance.inBalanceMoney = 0
	}

	//balance.buysCount++

	//Log(fmt.Sprintf(
	//	"Balance__BUY\nBuysCount: %d\nInBalanceMoney:%f\nUsedMoney: %f",
	//	balance.buysCount,
	//	balance.inBalanceMoney,
	//	usedMoney,
	//))
}

func (balance *Balance) sell(returnMoney float64) {
	//balance.buysCount--
	//if balance.buysCount < 0 {
	//	balance.buysCount = 0
	//}

	//if returnMoney > balance.config.TotalMoneyAmount {
	//	returnMoney = balance.config.TotalMoneyAmount
	//}

	balance.inBalanceMoney += returnMoney
	if balance.inBalanceMoney > BALANCE_MONEY+1 {
		panic(fmt.Sprintf("Sell: balance is higher, %f", balance.inBalanceMoney))
		balance.inBalanceMoney = BALANCE_MONEY
	}

	//Log(fmt.Sprintf(
	//	"Balance__SELL\nBuysCount: %d\nInBalanceMoney:%f\nReturnedMoney: %f",
	//	balance.buysCount,
	//	balance.inBalanceMoney,
	//	returnMoney,
	//))
}
