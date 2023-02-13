package main

import (
	"fmt"
)

type BotRevenue struct {
	BotNumber      int
	Revenue        float64
	TotalBuysCount int
}

func Fitness(botConfig Config, botNumber int, botRevenue chan BotRevenue, fitnessDatasets *[]Candle) {
	totalRevenue, totalBuysCount := doBuysAndSells(fitnessDatasets, botConfig)

	botRevenue <- BotRevenue{
		BotNumber:      botNumber,
		Revenue:        totalRevenue,
		TotalBuysCount: totalBuysCount,
	}
}

func doBuysAndSells(fitnessDatasets *[]Candle, botConfig Config) (float64, int) {
	bot := NewBot(&botConfig)

	for _, candle := range *fitnessDatasets {
		bot.DoStuff(candle)
	}

	rev := bot.db.GetTotalRevenue()
	buyCount := bot.db.GetBuysCount()
	commission := float64(buyCount) * COMMISSION
	datasetRevenue := rev - commission
	bot.Kill()

	fmt.Println(fmt.Sprintf(" DatasetRevenue: %f, TotalBuys: %d", datasetRevenue, buyCount))

	return datasetRevenue, buyCount
}
