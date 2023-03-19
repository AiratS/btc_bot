package main

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
}

func (balance *Balance) sell() {
	balance.buysCount--
	if balance.buysCount < 0 {
		balance.buysCount = 0
	}

	balance.inBalanceMoney += balance.config.TotalMoneyAmount
}
