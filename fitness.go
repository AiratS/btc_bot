package main

import (
	"fmt"
)

type BotRevenue struct {
	BotNumber        int
	Revenue          float64
	FinalBalance     float64
	TotalBuysCount   int
	UnsoldBuysCount  int
	LiquidationCount int
	AvgSellTime      float64

	ValidationRevenue          float64
	ValidationTotalBuysCount   int
	ValidationUnsoldBuysCount  int
	ValidationLiquidationCount int
	ValidationAvgSellTime      float64
}

func Fitness(
	botConfig Config,
	botNumber int,
	botRevenue chan BotRevenue,
	fitnessDatasets *[]Candle,
	validationDatasets *[]Candle,
) {
	totalRevenue, totalBuysCount, unsoldBuysCount, LiquidationCount, avgSellTime, FinalBalance := doBuysAndSells(fitnessDatasets, botConfig)

	// Validate bot
	Log(fmt.Sprintf("Validate bot: %d\n", botNumber))
	validationTotalRevenue, validationTotalBuysCount, ValidationUnsoldBuysCount, ValidationLiquidationCount, validationAvgSellTime := 0.0, 0, 0, 0, 0.0
	if !NO_VALIDATION {
		validationTotalRevenue, validationTotalBuysCount, ValidationUnsoldBuysCount, ValidationLiquidationCount, validationAvgSellTime, _ = doBuysAndSells(validationDatasets, botConfig)
	}

	botRevenue <- BotRevenue{
		BotNumber: botNumber,

		Revenue:          totalRevenue,
		FinalBalance:     FinalBalance,
		TotalBuysCount:   totalBuysCount,
		UnsoldBuysCount:  unsoldBuysCount,
		LiquidationCount: LiquidationCount,
		AvgSellTime:      avgSellTime,

		ValidationRevenue:          validationTotalRevenue,
		ValidationTotalBuysCount:   validationTotalBuysCount,
		ValidationLiquidationCount: ValidationLiquidationCount,
		ValidationAvgSellTime:      validationAvgSellTime,
		ValidationUnsoldBuysCount:  ValidationUnsoldBuysCount,
	}
}

func doBuysAndSells(fitnessDatasets *[]Candle, botConfig Config) (float64, int, int, int, float64, float64) {
	bot := NewBot(&botConfig)

	for _, candle := range *fitnessDatasets {
		bot.DoStuff(candle)
	}

	rev := 0.0
	liquidationsCount := bot.db.CountLiquidationBuys()
	rev = bot.db.GetFuturesTotalRevenue()
	if ENABLE_SHORT {
		rev = bot.db.GetShortFuturesTotalRevenue()
	}
	//rev -= float64(liquidationsCount) * bot.Config.TotalMoneyAmount

	//timeCancelRevenue := math.Abs(bot.db.GetTimeCancelTotalRevenue())
	//rev -= timeCancelRevenue

	if hasInvalidBuysCount(botConfig, liquidationsCount) {
		panic("Invalid liquidations count")
	}

	buyCount := bot.db.GetBuysCount()
	//commission := float64(buyCount) * COMMISSION
	commission := calcCommission(botConfig, buyCount)
	datasetRevenue := rev - commission
	unsold := bot.db.CountUnsoldBuys()
	//avgSellTime := bot.db.GetMedianSellTime()
	avgSellTime := bot.db.GetAvgSellTime()
	FinalBalance := bot.balance.inBalanceMoney

	fmt.Println(unsold)
	bot.Kill()

	fmt.Println(fmt.Sprintf(" DatasetRevenue: %f, TotalBuys: %d, UnsoldBuys: %d", datasetRevenue, buyCount, unsold))

	return datasetRevenue, buyCount, unsold, liquidationsCount, avgSellTime, FinalBalance
}

func hasInvalidBuysCount(botConfig Config, liquidationsCount int) bool {
	return (BALANCE_MONEY / botConfig.TotalMoneyAmount) < float64(liquidationsCount)
}

func calcCommission(botConfig Config, buyCount int) float64 {
	usedMoney := botConfig.TotalMoneyAmount
	if ENABLE_FUTURES {
		usedMoney = botConfig.TotalMoneyAmount * float64(botConfig.Leverage)
	}

	return (float64(buyCount) * usedMoney * COMMISSION) / 100
}
