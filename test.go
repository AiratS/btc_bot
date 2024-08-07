package main

import (
	"context"
	"fmt"
	"github.com/rocketlaunchr/dataframe-go"
	"github.com/rocketlaunchr/dataframe-go/exports"
	"math"
	"os"
)

func RunTest() {
	// Main
	LogAndPrint("Gen has started!")

	bots := GetInitialBots()
	//bots := GetInitialBotsFromFile("initial.csv")
	fitnessDatasets := ImportDatasets(GetDatasetDates())
	validationDatasets := &[]Candle{}
	if !NO_VALIDATION {
		validationDatasets = ImportDatasets(GetValidationDatasetDates())
	}

	for generation := 0; generation < GENERATION_COUNT; generation++ {
		var botRevenueChan = make(chan BotRevenue, 5)
		randValidationDataset := getRandomValidationDataset(validationDatasets)

		iterator := bots.ValuesIterator(dataframe.ValuesOptions{0, 1, true})
		for {
			botNumber, bot, _ := iterator()
			if botNumber == nil {
				break
			}

			if *botNumber < BEST_BOTS_FROM_PREV_GEN && generation > 0 {
				rev := convertToFloat64(bot["TotalRevenue"])
				totalBuysCount := convertToInt(bot["TotalBuysCount"])
				UnsoldBuysCount := convertToInt(bot["UnsoldBuysCount"])
				avgSellTime := convertToFloat64(bot["AvgSellTime"])
				//successPercentage := convertToFloat64(bot["SuccessPercentage"])

				fmt.Println(fmt.Sprintf("Gen: %d, Bot: %d", generation, *botNumber))
				fmt.Println(fmt.Sprintf("Gen: %d, Bot: %d, Revenue: %f, \n", generation, *botNumber, rev))
				SetBotTotalRevenue(
					bots,
					*botNumber,

					rev,
					totalBuysCount,
					UnsoldBuysCount,
					convertToInt(bot["LiquidationCount"]),
					avgSellTime,

					convertToFloat64(bot["ValidationTotalRevenue"]),
					convertToInt(bot["ValidationTotalBuysCount"]),
					convertToInt(bot["ValidationUnsoldBuysCount"]),
					convertToInt(bot["ValidationLiquidationCount"]),
					convertToFloat64(bot["ValidationAvgSellTime"]),
				)
				continue
			}

			fmt.Println(fmt.Sprintf("Gen: %d, Bot: %d", generation, *botNumber))
			botConfig := ConvertDataFrameToBotConfig(bot)
			go Fitness(botConfig, *botNumber, botRevenueChan, fitnessDatasets, &randValidationDataset)
		}

		channelsCount := bots.NRows()
		if generation > 0 {
			channelsCount = channelsCount - BEST_BOTS_FROM_PREV_GEN
		}

		for i := 0; i < channelsCount; i++ {
			botRevenue := <-botRevenueChan
			rev := fixRevenue(botRevenue.Revenue)
			SetBotTotalRevenue(
				bots,
				botRevenue.BotNumber,

				rev,
				botRevenue.TotalBuysCount,
				botRevenue.UnsoldBuysCount,
				botRevenue.LiquidationCount,
				botRevenue.AvgSellTime,

				fixRevenue(botRevenue.ValidationRevenue),
				botRevenue.ValidationTotalBuysCount,
				botRevenue.ValidationUnsoldBuysCount,
				botRevenue.ValidationLiquidationCount,
				botRevenue.ValidationAvgSellTime,
			)
			fmt.Println(fmt.Sprintf(
				"Gen: %d, Bot: %d, Buys Count: %d, Revenue: %f\n",
				generation,
				botRevenue.BotNumber,
				botRevenue.TotalBuysCount,
				rev,
			))
		}
		close(botRevenueChan)

		parentBots := SortBestBots(bots)
		botsCsvFile, _ := os.Create(fmt.Sprintf("generation_%d.csv", generation))
		exports.ExportToCSV(context.Background(), botsCsvFile, parentBots)

		bestBots := SelectNBots(BEST_BOTS_COUNT, parentBots)
		childBots := MakeChildren(bestBots)

		bots = CombineParentAndChildBots(
			SelectNBots(BEST_BOTS_FROM_PREV_GEN, bestBots),
			SelectNBots(BOTS_COUNT-BEST_BOTS_FROM_PREV_GEN, childBots),
		)
	}

	if canPlot() {
		PlotToJson("data.json")
		fmt.Println("Build plots")
		BuildPlots()
	}
}

func getRandomValidationDataset(validationDatasets *[]Candle) []Candle {
	if NO_VALIDATION {
		return []Candle{}
	}

	count := len(*validationDatasets)

	// After half of slice
	half := int(math.Round(float64(count) / 2))
	start := GetRandInt(0, half)
	end := count - 1

	// Before half of slice
	if GetRandInt(0, 1) == 1 {
		start = 0
		end = GetRandInt(half, count-1)
	}

	return (*validationDatasets)[start:end]
}

func fixRevenue(revenue float64) float64 {
	if revenue == 0.0 {
		return DEFAULT_REVENUE
	}
	return revenue
}
