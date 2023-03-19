package main

import (
	"context"
	"encoding/csv"
	"github.com/rocketlaunchr/dataframe-go"
	"math/rand"
	"os"
)

const BOTS_COUNT = 25
const BEST_BOTS_COUNT = 7
const BEST_BOTS_FROM_PREV_GEN = 3
const GENERATION_COUNT = 20
const DEFAULT_REVENUE = -1000000

func InitBotsDataFrame() *dataframe.DataFrame {
	return dataframe.NewDataFrame(
		dataframe.NewSeriesFloat64("HighSellPercentage", nil),

		dataframe.NewSeriesFloat64("TrailingTopPercentage", nil),
		dataframe.NewSeriesInt64("TrailingUpdateTimesBeforeFinish", nil),

		dataframe.NewSeriesInt64("WaitAfterLastBuyPeriod", nil),

		dataframe.NewSeriesInt64("BigFallCandlesCount", nil),
		dataframe.NewSeriesFloat64("BigFallPercentage", nil),

		dataframe.NewSeriesInt64("DesiredPriceCandles", nil),

		dataframe.NewSeriesInt64("GradientDescentCandles", nil),
		dataframe.NewSeriesInt64("GradientDescentPeriod", nil),
		dataframe.NewSeriesFloat64("GradientDescentGradient", nil),

		dataframe.NewSeriesFloat64("TrailingSellActivationAdditionPercentage", nil),
		dataframe.NewSeriesFloat64("TrailingSellStopPercentage", nil),

		dataframe.NewSeriesFloat64("TotalMoneyAmount", nil),

		dataframe.NewSeriesFloat64("TotalRevenue", nil),
		dataframe.NewSeriesFloat64("TotalBuysCount", nil),
		dataframe.NewSeriesFloat64("Selection", nil),
	)
}

func GetInitialBots() *dataframe.DataFrame {
	initialBotsDataFrame := InitBotsDataFrame()
	for botNumber := 0; botNumber < BOTS_COUNT; botNumber++ {
		botConfig := InitBotConfig()
		initialBotsDataFrame.Append(nil, GetBotConfigMapInterface(botConfig))
	}
	return initialBotsDataFrame
}

func GetInitialBotsFromFile(fileName string) *dataframe.DataFrame {
	initialBotsDataFrame := InitBotsDataFrame()
	csvBotConfigs := ImportFromCsv(fileName)
	for _, botConfig := range csvBotConfigs {
		initialBotsDataFrame.Append(nil, GetBotConfigMapInterface(botConfig))
	}
	return initialBotsDataFrame
}

func ImportFromCsv(fileName string) []Config {
	file, err := os.Open(fileName)
	if err != nil {
		panic("Can not load initial bots from file.")
	}

	csvReader := csv.NewReader(file)
	rows, err := csvReader.ReadAll()

	var bots []Config
	for rowNumber, row := range rows {
		if rowNumber == 0 {
			continue
		}

		bot := Config{
			HighSellPercentage: convertStringToFloat64(row[0]),

			TrailingTopPercentage:           convertStringToFloat64(row[1]),
			TrailingUpdateTimesBeforeFinish: convertStringToInt(row[2]),

			WaitAfterLastBuyPeriod: convertStringToInt(row[3]),

			BigFallCandlesCount: convertStringToInt(row[4]),
			BigFallPercentage:   convertStringToFloat64(row[5]),

			DesiredPriceCandles: convertStringToInt(row[6]),

			GradientDescentCandles:  convertStringToInt(row[7]),
			GradientDescentPeriod:   convertStringToInt(row[8]),
			GradientDescentGradient: convertStringToFloat64(row[9]),

			TrailingSellActivationAdditionPercentage: convertStringToFloat64(row[10]),
			TrailingSellStopPercentage:               convertStringToFloat64(row[11]),

			TotalMoneyAmount: convertStringToFloat64(row[12]),

			TotalRevenue:   convertStringToFloat64(row[13]),
			TotalBuysCount: convertStringToInt(row[14]),
			Selection:      convertStringToFloat64(row[15]),
		}

		bots = append(bots, bot)
	}

	return bots
}

func InitBotConfig() Config {
	restrict := GetBotConfigRestrictions()

	return Config{
		HighSellPercentage: GetRandFloat64Config(restrict.HighSellPercentage),

		TrailingTopPercentage:           GetRandFloat64Config(restrict.TrailingTopPercentage),
		TrailingUpdateTimesBeforeFinish: GetRandIntConfig(restrict.TrailingUpdateTimesBeforeFinish),

		WaitAfterLastBuyPeriod: GetRandIntConfig(restrict.WaitAfterLastBuyPeriod),

		BigFallCandlesCount: GetRandIntConfig(restrict.BigFallCandlesCount),
		BigFallPercentage:   GetRandFloat64Config(restrict.BigFallPercentage),

		DesiredPriceCandles: GetRandIntConfig(restrict.DesiredPriceCandles),

		GradientDescentCandles:  GetRandIntConfig(restrict.GradientDescentCandles),
		GradientDescentPeriod:   GetRandIntConfig(restrict.GradientDescentPeriod),
		GradientDescentGradient: GetRandFloat64Config(restrict.GradientDescentGradient),

		TrailingSellActivationAdditionPercentage: GetRandFloat64Config(restrict.TrailingSellActivationAdditionPercentage),
		TrailingSellStopPercentage:               GetRandFloat64Config(restrict.TrailingSellStopPercentage),

		TotalMoneyAmount: GetRandFloat64Config(restrict.TotalMoneyAmount),
	}
}

func GetBotConfigMapInterface(botConfig Config) map[string]interface{} {
	return map[string]interface{}{
		"HighSellPercentage": botConfig.HighSellPercentage,

		"TrailingTopPercentage":           botConfig.TrailingTopPercentage,
		"TrailingUpdateTimesBeforeFinish": botConfig.TrailingUpdateTimesBeforeFinish,

		"WaitAfterLastBuyPeriod": botConfig.WaitAfterLastBuyPeriod,

		"BigFallCandlesCount": botConfig.BigFallCandlesCount,
		"BigFallPercentage":   botConfig.BigFallPercentage,

		"DesiredPriceCandles": botConfig.DesiredPriceCandles,

		"GradientDescentCandles":  botConfig.GradientDescentCandles,
		"GradientDescentPeriod":   botConfig.GradientDescentPeriod,
		"GradientDescentGradient": botConfig.GradientDescentGradient,

		"TrailingSellActivationAdditionPercentage": botConfig.TrailingSellActivationAdditionPercentage,
		"TrailingSellStopPercentage":               botConfig.TrailingSellStopPercentage,

		"TotalMoneyAmount": botConfig.TotalMoneyAmount,

		"TotalRevenue":   botConfig.TotalRevenue,
		"TotalBuysCount": botConfig.TotalBuysCount,
		"Selection":      botConfig.Selection,
	}
}

func SetBotTotalRevenue(
	bots *dataframe.DataFrame,
	botNumber int,
	revenue float64,
	totalBuysCount int,
) {
	bots.UpdateRow(botNumber, nil, map[string]interface{}{
		"TotalRevenue":   revenue,
		"TotalBuysCount": totalBuysCount,
		"Selection":      revenue,
	})
}

func SortBestBots(bots *dataframe.DataFrame) *dataframe.DataFrame {
	sks := []dataframe.SortKey{
		//{
		//	Key:  "TotalRevenue",
		//	Desc: true,
		//},
		//{
		//	Key:  "SuccessPercentage",
		//	Desc: true,
		//},
		{
			Key:  "Selection",
			Desc: true,
		},
	}
	ctx := context.Background()
	bots.Sort(ctx, sks)
	return bots
}

func SelectNBots(numberOfBots int, bots *dataframe.DataFrame) *dataframe.DataFrame {
	botsDataFrame := InitBotsDataFrame()
	iterator := bots.ValuesIterator(dataframe.ValuesOptions{0, 1, true})
	alreadyHasRevenue := []float64{}

	for {
		botNumber, bot, _ := iterator()
		if botNumber == nil || numberOfBots < botsDataFrame.NRows()+1 {
			break
		}

		botRevenue := convertToFloat64(bot["TotalRevenue"])
		if 0 < CountInArray(botRevenue, &alreadyHasRevenue) {
			continue
		}

		if botRevenue != 0.0 && botRevenue != DEFAULT_REVENUE {
			alreadyHasRevenue = append(alreadyHasRevenue, botRevenue)
		}

		botsDataFrame.Append(nil, createBotDataFrameRow(bot))
	}
	return botsDataFrame
}

func createBotDataFrameRow(bot map[interface{}]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"HighSellPercentage": bot["HighSellPercentage"],

		"TrailingTopPercentage":           bot["TrailingTopPercentage"],
		"TrailingUpdateTimesBeforeFinish": bot["TrailingUpdateTimesBeforeFinish"],

		"WaitAfterLastBuyPeriod": bot["WaitAfterLastBuyPeriod"],

		"BigFallCandlesCount": bot["BigFallCandlesCount"],
		"BigFallPercentage":   bot["BigFallPercentage"],

		"DesiredPriceCandles": bot["DesiredPriceCandles"],

		"GradientDescentCandles":  bot["GradientDescentCandles"],
		"GradientDescentPeriod":   bot["GradientDescentPeriod"],
		"GradientDescentGradient": bot["GradientDescentGradient"],

		"TrailingSellActivationAdditionPercentage": bot["TrailingSellActivationAdditionPercentage"],
		"TrailingSellStopPercentage":               bot["TrailingSellStopPercentage"],

		"TotalMoneyAmount": bot["TotalMoneyAmount"],

		"TotalRevenue":   bot["TotalRevenue"],
		"TotalBuysCount": bot["TotalBuysCount"],
		"Selection":      bot["Selection"],
	}
}

func CombineParentAndChildBots(
	parentBots *dataframe.DataFrame,
	childBots *dataframe.DataFrame,
) *dataframe.DataFrame {
	botsDataFrame := InitBotsDataFrame()
	queueBots := []*dataframe.DataFrame{
		parentBots,
		childBots,
	}

	for _, bots := range queueBots {
		iterator := bots.ValuesIterator(dataframe.ValuesOptions{0, 1, true})

		for {
			botNumber, bot, _ := iterator()
			if botNumber == nil {
				break
			}
			botsDataFrame.Append(nil, createBotDataFrameRow(bot))
		}
	}

	return botsDataFrame
}

func MakeChildren(parentBots *dataframe.DataFrame) *dataframe.DataFrame {
	childrenBots := InitBotsDataFrame()
	maleIterator := parentBots.ValuesIterator(dataframe.ValuesOptions{0, 1, true})
	for {
		maleBotNumber, maleBot, _ := maleIterator()
		if maleBotNumber == nil {
			break
		}

		femaleIterator := parentBots.ValuesIterator(dataframe.ValuesOptions{0, 1, true})
		for {
			femaleBotNumber, femaleBot, _ := femaleIterator()
			if femaleBotNumber == nil {
				break
			}

			if *maleBotNumber == *femaleBotNumber {
				continue
			}

			child := makeChild(
				ConvertDataFrameToBotConfig(maleBot),
				ConvertDataFrameToBotConfig(femaleBot),
			)

			childrenBots.Append(nil, child)
		}
	}
	childrenBots = shuffleBots(childrenBots)
	return SelectNBots(BOTS_COUNT, childrenBots)
}

func ConvertDataFrameToBotConfig(dataFrame map[interface{}]interface{}) Config {
	return Config{
		HighSellPercentage: convertToFloat64(dataFrame["HighSellPercentage"]),

		TrailingTopPercentage:           convertToFloat64(dataFrame["TrailingTopPercentage"]),
		TrailingUpdateTimesBeforeFinish: convertToInt(dataFrame["TrailingUpdateTimesBeforeFinish"]),

		WaitAfterLastBuyPeriod: convertToInt(dataFrame["WaitAfterLastBuyPeriod"]),

		BigFallCandlesCount: convertToInt(dataFrame["BigFallCandlesCount"]),
		BigFallPercentage:   convertToFloat64(dataFrame["BigFallPercentage"]),

		DesiredPriceCandles: convertToInt(dataFrame["DesiredPriceCandles"]),

		GradientDescentCandles:  convertToInt(dataFrame["GradientDescentCandles"]),
		GradientDescentPeriod:   convertToInt(dataFrame["GradientDescentPeriod"]),
		GradientDescentGradient: convertToFloat64(dataFrame["GradientDescentGradient"]),

		TrailingSellActivationAdditionPercentage: convertToFloat64(dataFrame["TrailingSellActivationAdditionPercentage"]),
		TrailingSellStopPercentage:               convertToFloat64(dataFrame["TrailingSellStopPercentage"]),

		TotalMoneyAmount: convertToFloat64(dataFrame["TotalMoneyAmount"]),
	}
}

func makeChild(
	maleBotConfig Config,
	femaleBotConfig Config,
) map[string]interface{} {
	childBotConfig := Config{
		HighSellPercentage: GetFloatFatherOrMomGen(maleBotConfig.HighSellPercentage, femaleBotConfig.HighSellPercentage),

		TrailingTopPercentage:           GetFloatFatherOrMomGen(maleBotConfig.TrailingTopPercentage, femaleBotConfig.TrailingTopPercentage),
		TrailingUpdateTimesBeforeFinish: GetIntFatherOrMomGen(maleBotConfig.TrailingUpdateTimesBeforeFinish, femaleBotConfig.TrailingUpdateTimesBeforeFinish),

		WaitAfterLastBuyPeriod: GetIntFatherOrMomGen(maleBotConfig.WaitAfterLastBuyPeriod, femaleBotConfig.WaitAfterLastBuyPeriod),

		BigFallCandlesCount: GetIntFatherOrMomGen(maleBotConfig.BigFallCandlesCount, femaleBotConfig.BigFallCandlesCount),
		BigFallPercentage:   GetFloatFatherOrMomGen(maleBotConfig.BigFallPercentage, femaleBotConfig.BigFallPercentage),

		DesiredPriceCandles: GetIntFatherOrMomGen(maleBotConfig.DesiredPriceCandles, femaleBotConfig.DesiredPriceCandles),

		GradientDescentCandles:  GetIntFatherOrMomGen(maleBotConfig.GradientDescentCandles, femaleBotConfig.GradientDescentCandles),
		GradientDescentPeriod:   GetIntFatherOrMomGen(maleBotConfig.GradientDescentPeriod, femaleBotConfig.GradientDescentPeriod),
		GradientDescentGradient: GetFloatFatherOrMomGen(maleBotConfig.GradientDescentGradient, femaleBotConfig.GradientDescentGradient),

		TrailingSellActivationAdditionPercentage: GetFloatFatherOrMomGen(maleBotConfig.TrailingSellActivationAdditionPercentage, femaleBotConfig.TrailingSellActivationAdditionPercentage),
		TrailingSellStopPercentage:               GetFloatFatherOrMomGen(maleBotConfig.TrailingSellStopPercentage, femaleBotConfig.TrailingSellStopPercentage),

		TotalMoneyAmount: GetFloatFatherOrMomGen(maleBotConfig.TotalMoneyAmount, femaleBotConfig.TotalMoneyAmount),
	}

	for i := 0; i < 8; i++ {
		mutateGens(&childBotConfig, GetRandInt(0, 12))
	}

	return GetBotConfigMapInterface(childBotConfig)
}

func mutateGens(botConfig *Config, randGenNumber int) {
	restrict := GetBotConfigRestrictions()

	mutateGenFloat64(randGenNumber, 0, &(botConfig.HighSellPercentage), restrict.HighSellPercentage)

	mutateGenFloat64(randGenNumber, 1, &(botConfig.TrailingTopPercentage), restrict.TrailingTopPercentage)
	mutateGenInt(randGenNumber, 2, &(botConfig.TrailingUpdateTimesBeforeFinish), restrict.TrailingUpdateTimesBeforeFinish)

	mutateGenInt(randGenNumber, 3, &(botConfig.WaitAfterLastBuyPeriod), restrict.WaitAfterLastBuyPeriod)

	mutateGenInt(randGenNumber, 4, &(botConfig.BigFallCandlesCount), restrict.BigFallCandlesCount)
	mutateGenFloat64(randGenNumber, 5, &(botConfig.BigFallPercentage), restrict.BigFallPercentage)

	mutateGenInt(randGenNumber, 6, &(botConfig.DesiredPriceCandles), restrict.DesiredPriceCandles)

	mutateGenInt(randGenNumber, 7, &(botConfig.GradientDescentCandles), restrict.GradientDescentCandles)
	mutateGenInt(randGenNumber, 8, &(botConfig.GradientDescentPeriod), restrict.GradientDescentPeriod)
	mutateGenFloat64(randGenNumber, 9, &(botConfig.GradientDescentGradient), restrict.GradientDescentGradient)

	mutateGenFloat64(randGenNumber, 10, &(botConfig.TrailingSellActivationAdditionPercentage), restrict.TrailingSellActivationAdditionPercentage)
	mutateGenFloat64(randGenNumber, 11, &(botConfig.TrailingSellStopPercentage), restrict.TrailingSellStopPercentage)

	mutateGenFloat64(randGenNumber, 12, &(botConfig.TotalMoneyAmount), restrict.TotalMoneyAmount)
}

func mutateGenFloat64(randGenNumber, genNumber int, genValue *float64, restrictMinMax MinMaxFloat64) {
	if randGenNumber == genNumber {
		*genValue = MutateLittleFloat64(*genValue, restrictMinMax)
	}
}

func mutateGenInt(randGenNumber, genNumber int, genValue *int, restrictMinMax MinMaxInt) {
	if randGenNumber == genNumber {
		*genValue = MutateLittleInt(*genValue, restrictMinMax)
	}
}

func shuffleBots(bots *dataframe.DataFrame) *dataframe.DataFrame {
	shuffledBots := InitBotsDataFrame()
	shuffleNumbers := rand.Perm(bots.NRows())
	for i := 0; i < bots.NRows()-1; i++ {
		shuffledBots.Append(nil, bots.Row(shuffleNumbers[i], false))
	}
	return shuffledBots
}

func GetFloatFatherOrMomGen(maleGen, femaleGen float64) float64 {
	if GetRandInt(0, 1) == 1 {
		return maleGen
	}

	return femaleGen
}

func GetIntFatherOrMomGen(maleGen, femaleGen int) int {
	if GetRandInt(0, 1) == 1 {
		return maleGen
	}

	return femaleGen
}
