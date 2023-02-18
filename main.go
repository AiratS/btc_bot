package main

import (
	"context"
	"fmt"
	"github.com/rocketlaunchr/dataframe-go"
	"github.com/rocketlaunchr/dataframe-go/exports"
	"log"
	"os"
)

func main() {
	// Logger
	_, e := os.OpenFile("bot_log.txt", os.O_RDONLY, 0666)
	if !os.IsNotExist(e) {
		e := os.Remove("bot_log.txt")
		if e != nil {
			log.Fatal(e)
		}
	}

	f, err := os.OpenFile("bot_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Main
	LogAndPrint("Gen has started!")

	bots := GetInitialBots()
	fitnessDatasets := ImportDatasets()

	for generation := 0; generation < GENERATION_COUNT; generation++ {
		var botRevenueChan = make(chan BotRevenue, 5)

		iterator := bots.ValuesIterator(dataframe.ValuesOptions{0, 1, true})
		for {
			botNumber, bot, _ := iterator()
			if botNumber == nil {
				break
			}

			if *botNumber < BEST_BOTS_FROM_PREV_GEN && generation > 0 {
				rev := convertToFloat64(bot["TotalRevenue"])
				successPercentage := convertToFloat64(bot["SuccessPercentage"])

				fmt.Println(fmt.Sprintf("Gen: %d, Bot: %d", generation, *botNumber))
				fmt.Println(fmt.Sprintf("Gen: %d, Bot: %d, Revenue: %f, SuccessPercentage: %f\n", generation, *botNumber, rev, successPercentage))
				SetBotTotalRevenue(bots, *botNumber, rev)
				continue
			}

			fmt.Println(fmt.Sprintf("Gen: %d, Bot: %d", generation, *botNumber))
			botConfig := ConvertDataFrameToBotConfig(bot)
			go Fitness(botConfig, *botNumber, botRevenueChan, fitnessDatasets)
		}

		channelsCount := bots.NRows()
		if generation > 0 {
			channelsCount = channelsCount - BEST_BOTS_FROM_PREV_GEN
		}

		for i := 0; i < channelsCount; i++ {
			botRevenue := <-botRevenueChan
			rev := fixRevenue(botRevenue.Revenue)
			SetBotTotalRevenue(bots, botRevenue.BotNumber, rev)
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
}

func fixRevenue(revenue float64) float64 {
	if revenue == 0.0 {
		return DEFAULT_REVENUE
	}
	return revenue
}
