package main

import (
	"context"
	"encoding/csv"
	"github.com/rocketlaunchr/dataframe-go"
	"math/rand"
	"os"
)

func InitBotsDataFrame() *dataframe.DataFrame {
	return dataframe.NewDataFrame(
		dataframe.NewSeriesFloat64("HighSellPercentage", nil),

		dataframe.NewSeriesFloat64("TrailingTopPercentage", nil),
		dataframe.NewSeriesInt64("TrailingUpdateTimesBeforeFinish", nil),

		dataframe.NewSeriesInt64("WaitAfterLastBuyPeriod", nil),

		dataframe.NewSeriesInt64("BigFallCandlesCount", nil),
		dataframe.NewSeriesInt64("BigFallSmoothPeriod", nil),
		dataframe.NewSeriesFloat64("BigFallPercentage", nil),

		dataframe.NewSeriesInt64("DesiredPriceCandles", nil),

		dataframe.NewSeriesInt64("GradientDescentCandles", nil),
		dataframe.NewSeriesInt64("GradientDescentPeriod", nil),
		dataframe.NewSeriesFloat64("GradientDescentGradient", nil),

		dataframe.NewSeriesFloat64("TrailingSellActivationAdditionPercentage", nil),
		dataframe.NewSeriesFloat64("TrailingSellStopPercentage", nil),

		dataframe.NewSeriesFloat64("TotalMoneyAmount", nil),
		dataframe.NewSeriesInt64("Leverage", nil),
		dataframe.NewSeriesInt64("FuturesAvgSellTimeMinutes", nil),
		dataframe.NewSeriesFloat64("FuturesLeverageActivationPercentage", nil),

		dataframe.NewSeriesFloat64("TotalRevenue", nil),
		dataframe.NewSeriesInt64("TotalBuysCount", nil),
		dataframe.NewSeriesFloat64("AvgSellTime", nil),

		dataframe.NewSeriesFloat64("ValidationTotalRevenue", nil),
		dataframe.NewSeriesInt64("ValidationTotalBuysCount", nil),
		dataframe.NewSeriesFloat64("ValidationAvgSellTime", nil),

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
			BigFallSmoothPeriod: convertStringToInt(row[5]),
			BigFallPercentage:   convertStringToFloat64(row[6]),

			DesiredPriceCandles: convertStringToInt(row[7]),

			GradientDescentCandles:  convertStringToInt(row[8]),
			GradientDescentPeriod:   convertStringToInt(row[9]),
			GradientDescentGradient: convertStringToFloat64(row[10]),

			TrailingSellActivationAdditionPercentage: convertStringToFloat64(row[11]),
			TrailingSellStopPercentage:               convertStringToFloat64(row[12]),

			TotalMoneyAmount:                    convertStringToFloat64(row[13]),
			Leverage:                            convertStringToInt(row[14]),
			FuturesAvgSellTimeMinutes:           convertStringToInt(row[15]),
			FuturesLeverageActivationPercentage: convertStringToFloat64(row[16]),

			TotalRevenue:   convertStringToFloat64(row[17]),
			TotalBuysCount: convertStringToInt(row[18]),
			AvgSellTime:    convertStringToFloat64(row[19]),

			ValidationTotalRevenue:   convertStringToFloat64(row[20]),
			ValidationTotalBuysCount: convertStringToInt(row[21]),
			ValidationAvgSellTime:    convertStringToFloat64(row[22]),

			Selection: convertStringToFloat64(row[23]),
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
		BigFallSmoothPeriod: GetRandIntConfig(restrict.BigFallSmoothPeriod),
		BigFallPercentage:   GetRandFloat64Config(restrict.BigFallPercentage),

		DesiredPriceCandles: GetRandIntConfig(restrict.DesiredPriceCandles),

		GradientDescentCandles:  GetRandIntConfig(restrict.GradientDescentCandles),
		GradientDescentPeriod:   GetRandIntConfig(restrict.GradientDescentPeriod),
		GradientDescentGradient: GetRandFloat64Config(restrict.GradientDescentGradient),

		TrailingSellActivationAdditionPercentage: GetRandFloat64Config(restrict.TrailingSellActivationAdditionPercentage),
		TrailingSellStopPercentage:               GetRandFloat64Config(restrict.TrailingSellStopPercentage),

		TotalMoneyAmount:                    GetRandFloat64Config(restrict.TotalMoneyAmount),
		Leverage:                            GetRandIntConfig(restrict.Leverage),
		FuturesAvgSellTimeMinutes:           GetRandIntConfig(restrict.FuturesAvgSellTimeMinutes),
		FuturesLeverageActivationPercentage: GetRandFloat64Config(restrict.FuturesLeverageActivationPercentage),
	}
}

func GetBotConfigMapInterface(botConfig Config) map[string]interface{} {
	return map[string]interface{}{
		"HighSellPercentage": botConfig.HighSellPercentage,

		"TrailingTopPercentage":           botConfig.TrailingTopPercentage,
		"TrailingUpdateTimesBeforeFinish": botConfig.TrailingUpdateTimesBeforeFinish,

		"WaitAfterLastBuyPeriod": botConfig.WaitAfterLastBuyPeriod,

		"BigFallCandlesCount": botConfig.BigFallCandlesCount,
		"BigFallSmoothPeriod": botConfig.BigFallSmoothPeriod,
		"BigFallPercentage":   botConfig.BigFallPercentage,

		"DesiredPriceCandles": botConfig.DesiredPriceCandles,

		"GradientDescentCandles":  botConfig.GradientDescentCandles,
		"GradientDescentPeriod":   botConfig.GradientDescentPeriod,
		"GradientDescentGradient": botConfig.GradientDescentGradient,

		"TrailingSellActivationAdditionPercentage": botConfig.TrailingSellActivationAdditionPercentage,
		"TrailingSellStopPercentage":               botConfig.TrailingSellStopPercentage,

		"TotalMoneyAmount":                    botConfig.TotalMoneyAmount,
		"Leverage":                            botConfig.Leverage,
		"FuturesAvgSellTimeMinutes":           botConfig.FuturesAvgSellTimeMinutes,
		"FuturesLeverageActivationPercentage": botConfig.FuturesLeverageActivationPercentage,

		"TotalRevenue":   botConfig.TotalRevenue,
		"TotalBuysCount": botConfig.TotalBuysCount,
		"AvgSellTime":    botConfig.AvgSellTime,

		"ValidationTotalRevenue":   botConfig.ValidationTotalRevenue,
		"ValidationTotalBuysCount": botConfig.ValidationTotalBuysCount,
		"ValidationAvgSellTime":    botConfig.ValidationAvgSellTime,

		"Selection": botConfig.Selection,
	}
}

func SetBotTotalRevenue(
	bots *dataframe.DataFrame,
	botNumber int,

	revenue float64,
	totalBuysCount int,
	avgSellTime float64,

	ValidationTotalRevenue float64,
	ValidationTotalBuysCount int,
	ValidationAvgSellTime float64,
) {
	selectionDivider := 1.0

	if ENABLE_AVG_TIME {
		selectionDivider = avgSellTime * SELL_TIME_PUNISHMENT
		if avgSellTime == 0 {
			selectionDivider = 1
		}
	}

	if NO_VALIDATION {
		ValidationTotalRevenue = 0
	}

	bots.UpdateRow(botNumber, nil, map[string]interface{}{
		"TotalRevenue":   revenue,
		"TotalBuysCount": totalBuysCount,
		"AvgSellTime":    avgSellTime,

		"ValidationTotalRevenue":   ValidationTotalRevenue,
		"ValidationTotalBuysCount": ValidationTotalBuysCount,
		"ValidationAvgSellTime":    ValidationAvgSellTime,

		"Selection": (revenue + ValidationTotalRevenue) / selectionDivider,
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
		//{
		//	Key:  "ValidationTotalRevenue",
		//	Desc: true,
		//},
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
		"BigFallSmoothPeriod": bot["BigFallSmoothPeriod"],
		"BigFallPercentage":   bot["BigFallPercentage"],

		"DesiredPriceCandles": bot["DesiredPriceCandles"],

		"GradientDescentCandles":  bot["GradientDescentCandles"],
		"GradientDescentPeriod":   bot["GradientDescentPeriod"],
		"GradientDescentGradient": bot["GradientDescentGradient"],

		"TrailingSellActivationAdditionPercentage": bot["TrailingSellActivationAdditionPercentage"],
		"TrailingSellStopPercentage":               bot["TrailingSellStopPercentage"],

		"TotalMoneyAmount":                    bot["TotalMoneyAmount"],
		"Leverage":                            bot["Leverage"],
		"FuturesAvgSellTimeMinutes":           bot["FuturesAvgSellTimeMinutes"],
		"FuturesLeverageActivationPercentage": bot["FuturesLeverageActivationPercentage"],

		"TotalRevenue":   bot["TotalRevenue"],
		"TotalBuysCount": bot["TotalBuysCount"],
		"AvgSellTime":    bot["AvgSellTime"],

		"ValidationTotalRevenue":   bot["ValidationTotalRevenue"],
		"ValidationTotalBuysCount": bot["ValidationTotalBuysCount"],
		"ValidationAvgSellTime":    bot["ValidationAvgSellTime"],

		"Selection": bot["Selection"],
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
		BigFallSmoothPeriod: convertToInt(dataFrame["BigFallSmoothPeriod"]),
		BigFallPercentage:   convertToFloat64(dataFrame["BigFallPercentage"]),

		DesiredPriceCandles: convertToInt(dataFrame["DesiredPriceCandles"]),

		GradientDescentCandles:  convertToInt(dataFrame["GradientDescentCandles"]),
		GradientDescentPeriod:   convertToInt(dataFrame["GradientDescentPeriod"]),
		GradientDescentGradient: convertToFloat64(dataFrame["GradientDescentGradient"]),

		TrailingSellActivationAdditionPercentage: convertToFloat64(dataFrame["TrailingSellActivationAdditionPercentage"]),
		TrailingSellStopPercentage:               convertToFloat64(dataFrame["TrailingSellStopPercentage"]),

		TotalMoneyAmount:                    convertToFloat64(dataFrame["TotalMoneyAmount"]),
		Leverage:                            convertToInt(dataFrame["Leverage"]),
		FuturesAvgSellTimeMinutes:           convertToInt(dataFrame["FuturesAvgSellTimeMinutes"]),
		FuturesLeverageActivationPercentage: convertToFloat64(dataFrame["FuturesLeverageActivationPercentage"]),
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
		BigFallSmoothPeriod: GetIntFatherOrMomGen(maleBotConfig.BigFallSmoothPeriod, femaleBotConfig.BigFallSmoothPeriod),
		BigFallPercentage:   GetFloatFatherOrMomGen(maleBotConfig.BigFallPercentage, femaleBotConfig.BigFallPercentage),

		DesiredPriceCandles: GetIntFatherOrMomGen(maleBotConfig.DesiredPriceCandles, femaleBotConfig.DesiredPriceCandles),

		GradientDescentCandles:  GetIntFatherOrMomGen(maleBotConfig.GradientDescentCandles, femaleBotConfig.GradientDescentCandles),
		GradientDescentPeriod:   GetIntFatherOrMomGen(maleBotConfig.GradientDescentPeriod, femaleBotConfig.GradientDescentPeriod),
		GradientDescentGradient: GetFloatFatherOrMomGen(maleBotConfig.GradientDescentGradient, femaleBotConfig.GradientDescentGradient),

		TrailingSellActivationAdditionPercentage: GetFloatFatherOrMomGen(maleBotConfig.TrailingSellActivationAdditionPercentage, femaleBotConfig.TrailingSellActivationAdditionPercentage),
		TrailingSellStopPercentage:               GetFloatFatherOrMomGen(maleBotConfig.TrailingSellStopPercentage, femaleBotConfig.TrailingSellStopPercentage),

		TotalMoneyAmount:                    GetFloatFatherOrMomGen(maleBotConfig.TotalMoneyAmount, femaleBotConfig.TotalMoneyAmount),
		Leverage:                            GetIntFatherOrMomGen(maleBotConfig.Leverage, femaleBotConfig.Leverage),
		FuturesAvgSellTimeMinutes:           GetIntFatherOrMomGen(maleBotConfig.FuturesAvgSellTimeMinutes, femaleBotConfig.FuturesAvgSellTimeMinutes),
		FuturesLeverageActivationPercentage: GetFloatFatherOrMomGen(maleBotConfig.FuturesLeverageActivationPercentage, femaleBotConfig.FuturesLeverageActivationPercentage),
	}

	for i := 0; i < 10; i++ {
		mutateGens(&childBotConfig, GetRandInt(0, 16))
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
	mutateGenInt(randGenNumber, 5, &(botConfig.BigFallSmoothPeriod), restrict.BigFallSmoothPeriod)
	mutateGenFloat64(randGenNumber, 6, &(botConfig.BigFallPercentage), restrict.BigFallPercentage)

	mutateGenInt(randGenNumber, 7, &(botConfig.DesiredPriceCandles), restrict.DesiredPriceCandles)

	mutateGenInt(randGenNumber, 8, &(botConfig.GradientDescentCandles), restrict.GradientDescentCandles)
	mutateGenInt(randGenNumber, 9, &(botConfig.GradientDescentPeriod), restrict.GradientDescentPeriod)
	mutateGenFloat64(randGenNumber, 10, &(botConfig.GradientDescentGradient), restrict.GradientDescentGradient)

	mutateGenFloat64(randGenNumber, 11, &(botConfig.TrailingSellActivationAdditionPercentage), restrict.TrailingSellActivationAdditionPercentage)
	mutateGenFloat64(randGenNumber, 12, &(botConfig.TrailingSellStopPercentage), restrict.TrailingSellStopPercentage)

	mutateGenFloat64(randGenNumber, 13, &(botConfig.TotalMoneyAmount), restrict.TotalMoneyAmount)
	mutateGenInt(randGenNumber, 14, &(botConfig.Leverage), restrict.Leverage)
	mutateGenInt(randGenNumber, 15, &(botConfig.FuturesAvgSellTimeMinutes), restrict.FuturesAvgSellTimeMinutes)
	mutateGenFloat64(randGenNumber, 16, &(botConfig.FuturesLeverageActivationPercentage), restrict.FuturesLeverageActivationPercentage)
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
