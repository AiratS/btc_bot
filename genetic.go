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
		dataframe.NewSeriesFloat64("FirstBuyHighSellPercentage", nil),

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

		dataframe.NewSeriesInt64("LinearRegressionCandles", nil),
		dataframe.NewSeriesInt64("LinearRegressionPeriod", nil),
		dataframe.NewSeriesFloat64("LinearRegressionMse", nil),
		dataframe.NewSeriesFloat64("LinearRegressionK", nil),
		dataframe.NewSeriesFloat64("LinearRegressionDeviation", nil),

		dataframe.NewSeriesInt64("GradientSwingIndicatorCandles", nil),
		dataframe.NewSeriesInt64("GradientSwingIndicatorPeriod", nil),
		dataframe.NewSeriesInt64("GradientSwingIndicatorSwingType", nil),

		dataframe.NewSeriesInt64("CatchingFallingKnifeCandles", nil),
		dataframe.NewSeriesFloat64("CatchingFallingKnifeSellPercentage", nil),
		dataframe.NewSeriesFloat64("CatchingFallingKnifeAdditionalBuyPercentage", nil),

		dataframe.NewSeriesFloat64("TotalMoneyAmount", nil),
		dataframe.NewSeriesFloat64("TotalMoneyIncreasePercentage", nil),
		dataframe.NewSeriesFloat64("FirstBuyMoneyIncreasePercentage", nil),
		dataframe.NewSeriesInt64("StopIncreaseMoneyAfterBuysCount", nil),
		dataframe.NewSeriesInt64("Leverage", nil),
		dataframe.NewSeriesInt64("FuturesAvgSellTimeMinutes", nil),
		dataframe.NewSeriesFloat64("FuturesLeverageActivationPercentage", nil),

		dataframe.NewSeriesFloat64("LessThanPreviousBuyPercentage", nil),
		dataframe.NewSeriesFloat64("MoreThanPreviousBuyPercentage", nil),
		dataframe.NewSeriesFloat64("ParabolaDivider", nil),

		dataframe.NewSeriesFloat64("BoostBuyFallPercentage", nil),
		dataframe.NewSeriesInt64("BoostBuyPeriodMinutes", nil),
		dataframe.NewSeriesFloat64("BoostBuyMoneyIncreasePercentage", nil),

		dataframe.NewSeriesInt64("StopAfterUnsuccessfullySellMinutes", nil),

		// ------------------------
		dataframe.NewSeriesInt64("WindowWindowsCount", nil),
		dataframe.NewSeriesFloat64("WindowBasePercentage", nil),
		dataframe.NewSeriesFloat64("WindowOffsetPercentage", nil),
		dataframe.NewSeriesInt64("WindowBasePeriodMinutes", nil),
		dataframe.NewSeriesInt64("WindowOffsetPeriodMinutes", nil),
		// ------------------------

		dataframe.NewSeriesFloat64("TotalRevenue", nil),
		dataframe.NewSeriesFloat64("FinalBalance", nil),
		dataframe.NewSeriesFloat64("FinalRevenue", nil),
		dataframe.NewSeriesInt64("TotalBuysCount", nil),
		dataframe.NewSeriesInt64("UnsoldBuysCount", nil),
		dataframe.NewSeriesInt64("LiquidationCount", nil),
		dataframe.NewSeriesFloat64("AvgSellTime", nil),

		dataframe.NewSeriesFloat64("ValidationTotalRevenue", nil),
		dataframe.NewSeriesInt64("ValidationTotalBuysCount", nil),
		dataframe.NewSeriesInt64("ValidationUnsoldBuysCount", nil),
		dataframe.NewSeriesInt64("ValidationLiquidationCount", nil),
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
			HighSellPercentage:         convertStringToFloat64(row[0]),
			FirstBuyHighSellPercentage: convertStringToFloat64(row[1]),

			TrailingTopPercentage:           convertStringToFloat64(row[2]),
			TrailingUpdateTimesBeforeFinish: convertStringToInt(row[3]),

			WaitAfterLastBuyPeriod: convertStringToInt(row[4]),

			BigFallCandlesCount: convertStringToInt(row[5]),
			BigFallSmoothPeriod: convertStringToInt(row[6]),
			BigFallPercentage:   convertStringToFloat64(row[7]),

			DesiredPriceCandles: convertStringToInt(row[8]),

			GradientDescentCandles:  convertStringToInt(row[9]),
			GradientDescentPeriod:   convertStringToInt(row[10]),
			GradientDescentGradient: convertStringToFloat64(row[11]),

			TrailingSellActivationAdditionPercentage: convertStringToFloat64(row[12]),
			TrailingSellStopPercentage:               convertStringToFloat64(row[13]),

			LinearRegressionCandles:   convertStringToInt(row[14]),
			LinearRegressionPeriod:    convertStringToInt(row[15]),
			LinearRegressionMse:       convertStringToFloat64(row[16]),
			LinearRegressionK:         convertStringToFloat64(row[17]),
			LinearRegressionDeviation: convertStringToFloat64(row[18]),

			GradientSwingIndicatorCandles:   convertStringToInt(row[19]),
			GradientSwingIndicatorPeriod:    convertStringToInt(row[20]),
			GradientSwingIndicatorSwingType: convertStringToInt(row[21]),

			CatchingFallingKnifeCandles:   convertStringToInt(row[22]),
			CatchingFallingKnifeSellPercentage:    convertStringToInt(row[23]),
			CatchingFallingKnifeAdditionalBuyPercentage: convertStringToInt(row[24]),

			TotalMoneyAmount:                convertStringToFloat64(row[25]),
			TotalMoneyIncreasePercentage:    convertStringToFloat64(row[26]),
			FirstBuyMoneyIncreasePercentage: convertStringToFloat64(row[27]),
			StopIncreaseMoneyAfterBuysCount: convertStringToInt(row[28]),
			Leverage:                        convertStringToInt(row[29]),

			FuturesAvgSellTimeMinutes:           convertStringToInt(row[30]),
			FuturesLeverageActivationPercentage: convertStringToFloat64(row[31]),

			LessThanPreviousBuyPercentage: convertStringToFloat64(row[32]),
			MoreThanPreviousBuyPercentage: convertStringToFloat64(row[33]),
			ParabolaDivider:               convertStringToFloat64(row[34]),

			BoostBuyFallPercentage:          convertStringToFloat64(row[35]),
			BoostBuyPeriodMinutes:           convertStringToInt(row[36]),
			BoostBuyMoneyIncreasePercentage: convertStringToFloat64(row[37]),

			StopAfterUnsuccessfullySellMinutes: convertStringToInt(row[38]),

			// ------------------------
			WindowWindowsCount:        convertStringToInt(row[39]),
			WindowBasePercentage:      convertStringToFloat64(row[40]),
			WindowOffsetPercentage:    convertStringToFloat64(row[41]),
			WindowBasePeriodMinutes:   convertStringToInt(row[42]),
			WindowOffsetPeriodMinutes: convertStringToInt(row[43]),
			// ------------------------

			TotalRevenue:     convertStringToFloat64(row[44]),
			FinalBalance:     convertStringToFloat64(row[45]),
			FinalRevenue:     convertStringToFloat64(row[46]),
			TotalBuysCount:   convertStringToInt(row[47]),
			UnsoldBuysCount:  convertStringToInt(row[48]),
			LiquidationCount: convertStringToInt(row[49]),
			AvgSellTime:      convertStringToFloat64(row[50]),

			ValidationTotalRevenue:     convertStringToFloat64(row[51]),
			ValidationTotalBuysCount:   convertStringToInt(row[52]),
			ValidationUnsoldBuysCount:  convertStringToInt(row[53]),
			ValidationLiquidationCount: convertStringToInt(row[54]),
			ValidationAvgSellTime:      convertStringToFloat64(row[55]),

			Selection: convertStringToFloat64(row[56]),
		}

		bots = append(bots, bot)
	}

	return bots
}

func InitBotConfig() Config {
	restrict := GetBotConfigRestrictions()

	return Config{
		HighSellPercentage:         GetRandFloat64Config(restrict.HighSellPercentage),
		FirstBuyHighSellPercentage: GetRandFloat64Config(restrict.FirstBuyHighSellPercentage),

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

		LinearRegressionCandles:   GetRandIntConfig(restrict.LinearRegressionCandles),
		LinearRegressionPeriod:    GetRandIntConfig(restrict.LinearRegressionPeriod),
		LinearRegressionMse:       GetRandFloat64Config(restrict.LinearRegressionMse),
		LinearRegressionK:         GetRandFloat64Config(restrict.LinearRegressionK),
		LinearRegressionDeviation: GetRandFloat64Config(restrict.LinearRegressionDeviation),

		GradientSwingIndicatorCandles:   GetRandIntConfig(restrict.GradientSwingIndicatorCandles),
		GradientSwingIndicatorPeriod:    GetRandIntConfig(restrict.GradientSwingIndicatorPeriod),
		GradientSwingIndicatorSwingType: GetRandIntConfig(restrict.GradientSwingIndicatorSwingType),

		CatchingFallingKnifeCandles:    GetRandIntConfig(restrict.CatchingFallingKnifeCandles),
		CatchingFallingKnifeSellPercentage:       GetRandFloat64Config(restrict.CatchingFallingKnifeSellPercentage),
		CatchingFallingKnifeAdditionalBuyPercentage:         GetRandFloat64Config(restrict.CatchingFallingKnifeAdditionalBuyPercentage),

		TotalMoneyAmount:                    GetRandFloat64Config(restrict.TotalMoneyAmount),
		TotalMoneyIncreasePercentage:        GetRandFloat64Config(restrict.TotalMoneyIncreasePercentage),
		FirstBuyMoneyIncreasePercentage:     GetRandFloat64Config(restrict.FirstBuyMoneyIncreasePercentage),
		StopIncreaseMoneyAfterBuysCount:     GetRandIntConfig(restrict.StopIncreaseMoneyAfterBuysCount),
		Leverage:                            GetRandIntConfig(restrict.Leverage),
		FuturesAvgSellTimeMinutes:           GetRandIntConfig(restrict.FuturesAvgSellTimeMinutes),
		FuturesLeverageActivationPercentage: GetRandFloat64Config(restrict.FuturesLeverageActivationPercentage),

		LessThanPreviousBuyPercentage: GetRandFloat64Config(restrict.LessThanPreviousBuyPercentage),
		MoreThanPreviousBuyPercentage: GetRandFloat64Config(restrict.MoreThanPreviousBuyPercentage),
		ParabolaDivider:               GetRandFloat64Config(restrict.ParabolaDivider),

		BoostBuyFallPercentage:          GetRandFloat64Config(restrict.BoostBuyFallPercentage),
		BoostBuyPeriodMinutes:           GetRandIntConfig(restrict.BoostBuyPeriodMinutes),
		BoostBuyMoneyIncreasePercentage: GetRandFloat64Config(restrict.BoostBuyMoneyIncreasePercentage),

		StopAfterUnsuccessfullySellMinutes: GetRandIntConfig(restrict.StopAfterUnsuccessfullySellMinutes),

		// ----------------------------
		WindowWindowsCount:        GetRandIntConfig(restrict.WindowWindowsCount),
		WindowBasePercentage:      GetRandFloat64Config(restrict.WindowBasePercentage),
		WindowOffsetPercentage:    GetRandFloat64Config(restrict.WindowOffsetPercentage),
		WindowBasePeriodMinutes:   GetRandIntConfig(restrict.WindowBasePeriodMinutes),
		WindowOffsetPeriodMinutes: GetRandIntConfig(restrict.WindowOffsetPeriodMinutes),
		// ----------------------------
	}
}

func GetBotConfigMapInterface(botConfig Config) map[string]interface{} {
	return map[string]interface{}{
		"HighSellPercentage":         botConfig.HighSellPercentage,
		"FirstBuyHighSellPercentage": botConfig.FirstBuyHighSellPercentage,

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

		"LinearRegressionCandles":   botConfig.LinearRegressionCandles,
		"LinearRegressionPeriod":    botConfig.LinearRegressionPeriod,
		"LinearRegressionMse":       botConfig.LinearRegressionMse,
		"LinearRegressionK":         botConfig.LinearRegressionK,
		"LinearRegressionDeviation": botConfig.LinearRegressionDeviation,

		"GradientSwingIndicatorCandles":   botConfig.GradientSwingIndicatorCandles,
		"GradientSwingIndicatorPeriod":    botConfig.GradientSwingIndicatorPeriod,
		"GradientSwingIndicatorSwingType": botConfig.GradientSwingIndicatorSwingType,

		"CatchingFallingKnifeCandles":   botConfig.CatchingFallingKnifeCandles,
		"CatchingFallingKnifeSellPercentage":    botConfig.CatchingFallingKnifeSellPercentage,
		"CatchingFallingKnifeAdditionalBuyPercentage": botConfig.CatchingFallingKnifeAdditionalBuyPercentage,

		"TotalMoneyAmount":                    botConfig.TotalMoneyAmount,
		"TotalMoneyIncreasePercentage":        botConfig.TotalMoneyIncreasePercentage,
		"FirstBuyMoneyIncreasePercentage":     botConfig.FirstBuyMoneyIncreasePercentage,
		"StopIncreaseMoneyAfterBuysCount":     botConfig.StopIncreaseMoneyAfterBuysCount,
		"Leverage":                            botConfig.Leverage,
		"FuturesAvgSellTimeMinutes":           botConfig.FuturesAvgSellTimeMinutes,
		"FuturesLeverageActivationPercentage": botConfig.FuturesLeverageActivationPercentage,

		"LessThanPreviousBuyPercentage": botConfig.LessThanPreviousBuyPercentage,
		"MoreThanPreviousBuyPercentage": botConfig.MoreThanPreviousBuyPercentage,
		"ParabolaDivider":               botConfig.ParabolaDivider,

		"BoostBuyFallPercentage":          botConfig.BoostBuyFallPercentage,
		"BoostBuyPeriodMinutes":           botConfig.BoostBuyPeriodMinutes,
		"BoostBuyMoneyIncreasePercentage": botConfig.BoostBuyMoneyIncreasePercentage,

		"StopAfterUnsuccessfullySellMinutes": botConfig.StopAfterUnsuccessfullySellMinutes,

		// --------------------------
		"WindowWindowsCount":        botConfig.WindowWindowsCount,
		"WindowBasePercentage":      botConfig.WindowBasePercentage,
		"WindowOffsetPercentage":    botConfig.WindowOffsetPercentage,
		"WindowBasePeriodMinutes":   botConfig.WindowBasePeriodMinutes,
		"WindowOffsetPeriodMinutes": botConfig.WindowOffsetPeriodMinutes,
		// --------------------------

		"TotalRevenue":     botConfig.TotalRevenue,
		"FinalBalance":     botConfig.FinalBalance,
		"FinalRevenue":     botConfig.FinalRevenue,
		"TotalBuysCount":   botConfig.TotalBuysCount,
		"UnsoldBuysCount":  botConfig.UnsoldBuysCount,
		"LiquidationCount": botConfig.LiquidationCount,
		"AvgSellTime":      botConfig.AvgSellTime,

		"ValidationTotalRevenue":     botConfig.ValidationTotalRevenue,
		"ValidationTotalBuysCount":   botConfig.ValidationTotalBuysCount,
		"ValidationUnsoldBuysCount":  botConfig.ValidationUnsoldBuysCount,
		"ValidationLiquidationCount": botConfig.ValidationLiquidationCount,
		"ValidationAvgSellTime":      botConfig.ValidationAvgSellTime,

		"Selection": botConfig.Selection,
	}
}

func SetBotTotalRevenue(
	bots *dataframe.DataFrame,
	botNumber int,

	revenue float64,
	FinalBalance float64,
	totalBuysCount int,
	UnsoldBuysCount int,
	LiquidationCount int,
	avgSellTime float64,

	ValidationTotalRevenue float64,
	ValidationTotalBuysCount int,
	ValidationUnsoldBuysCount int,
	ValidationLiquidationCount int,
	ValidationAvgSellTime float64,
) {
	selectionDivider := 1.0

	if ENABLE_AVG_TIME {
		selectionDivider = avgSellTime * SELL_TIME_PUNISHMENT
		if avgSellTime < 1.5 {
			selectionDivider = 1
		}
	}

	//plusValidation := 0.0
	//if NO_VALIDATION {
	//	ValidationTotalRevenue = 0
	//} else {
	//	plusValidation = ValidationTotalRevenue
	//	//if ValidationTotalRevenue < 0 {
	//	//	plusValidation = -10 * ValidationTotalRevenue
	//	//} else {
	//	//	plusValidation = ValidationTotalRevenue
	//	//}
	//}

	minusFromBalance := BALANCE_MONEY - FinalBalance
	FinalRevenue := revenue - minusFromBalance

	bots.UpdateRow(botNumber, nil, map[string]interface{}{
		"TotalRevenue":     revenue,
		"FinalBalance":     FinalBalance,
		"FinalRevenue":     FinalRevenue,
		"TotalBuysCount":   totalBuysCount,
		"UnsoldBuysCount":  UnsoldBuysCount,
		"LiquidationCount": LiquidationCount,
		"AvgSellTime":      avgSellTime,

		"ValidationTotalRevenue":     ValidationTotalRevenue,
		"ValidationTotalBuysCount":   ValidationTotalBuysCount,
		"ValidationUnsoldBuysCount":  ValidationUnsoldBuysCount,
		"ValidationLiquidationCount": ValidationLiquidationCount,
		"ValidationAvgSellTime":      ValidationAvgSellTime,

		"Selection": FinalRevenue / selectionDivider,
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
		"HighSellPercentage":         bot["HighSellPercentage"],
		"FirstBuyHighSellPercentage": bot["FirstBuyHighSellPercentage"],

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

		"LinearRegressionCandles":   bot["LinearRegressionCandles"],
		"LinearRegressionPeriod":    bot["LinearRegressionPeriod"],
		"LinearRegressionMse":       bot["LinearRegressionMse"],
		"LinearRegressionK":         bot["LinearRegressionK"],
		"LinearRegressionDeviation": bot["LinearRegressionDeviation"],

		"GradientSwingIndicatorCandles":   bot["GradientSwingIndicatorCandles"],
		"GradientSwingIndicatorPeriod":    bot["GradientSwingIndicatorPeriod"],
		"GradientSwingIndicatorSwingType": bot["GradientSwingIndicatorSwingType"],

		"CatchingFallingKnifeCandles":   bot["CatchingFallingKnifeCandles"],
		"CatchingFallingKnifeSellPercentage":    bot["CatchingFallingKnifeSellPercentage"],
		"CatchingFallingKnifeAdditionalBuyPercentage": bot["CatchingFallingKnifeAdditionalBuyPercentage"],

		"TotalMoneyAmount":                    bot["TotalMoneyAmount"],
		"TotalMoneyIncreasePercentage":        bot["TotalMoneyIncreasePercentage"],
		"FirstBuyMoneyIncreasePercentage":     bot["FirstBuyMoneyIncreasePercentage"],
		"StopIncreaseMoneyAfterBuysCount":     bot["StopIncreaseMoneyAfterBuysCount"],
		"Leverage":                            bot["Leverage"],
		"FuturesAvgSellTimeMinutes":           bot["FuturesAvgSellTimeMinutes"],
		"FuturesLeverageActivationPercentage": bot["FuturesLeverageActivationPercentage"],

		"LessThanPreviousBuyPercentage": bot["LessThanPreviousBuyPercentage"],
		"MoreThanPreviousBuyPercentage": bot["MoreThanPreviousBuyPercentage"],
		"ParabolaDivider":               bot["ParabolaDivider"],

		"BoostBuyFallPercentage":          bot["BoostBuyFallPercentage"],
		"BoostBuyPeriodMinutes":           bot["BoostBuyPeriodMinutes"],
		"BoostBuyMoneyIncreasePercentage": bot["BoostBuyMoneyIncreasePercentage"],

		"StopAfterUnsuccessfullySellMinutes": bot["StopAfterUnsuccessfullySellMinutes"],

		// -------------------------------
		"WindowWindowsCount":        bot["WindowWindowsCount"],
		"WindowBasePercentage":      bot["WindowBasePercentage"],
		"WindowOffsetPercentage":    bot["WindowOffsetPercentage"],
		"WindowBasePeriodMinutes":   bot["WindowBasePeriodMinutes"],
		"WindowOffsetPeriodMinutes": bot["WindowOffsetPeriodMinutes"],
		// -------------------------------

		"TotalRevenue":     bot["TotalRevenue"],
		"FinalBalance":     bot["FinalBalance"],
		"FinalRevenue":     bot["FinalRevenue"],
		"TotalBuysCount":   bot["TotalBuysCount"],
		"UnsoldBuysCount":  bot["UnsoldBuysCount"],
		"LiquidationCount": bot["LiquidationCount"],
		"AvgSellTime":      bot["AvgSellTime"],

		"ValidationTotalRevenue":     bot["ValidationTotalRevenue"],
		"ValidationTotalBuysCount":   bot["ValidationTotalBuysCount"],
		"ValidationUnsoldBuysCount":  bot["ValidationUnsoldBuysCount"],
		"ValidationLiquidationCount": bot["ValidationLiquidationCount"],
		"ValidationAvgSellTime":      bot["ValidationAvgSellTime"],

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
		HighSellPercentage:         convertToFloat64(dataFrame["HighSellPercentage"]),
		FirstBuyHighSellPercentage: convertToFloat64(dataFrame["FirstBuyHighSellPercentage"]),

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

		LinearRegressionCandles:   convertToInt(dataFrame["LinearRegressionCandles"]),
		LinearRegressionPeriod:    convertToInt(dataFrame["LinearRegressionPeriod"]),
		LinearRegressionMse:       convertToFloat64(dataFrame["LinearRegressionMse"]),
		LinearRegressionK:         convertToFloat64(dataFrame["LinearRegressionK"]),
		LinearRegressionDeviation: convertToFloat64(dataFrame["LinearRegressionDeviation"]),

		GradientSwingIndicatorCandles:   convertToInt(dataFrame["GradientSwingIndicatorCandles"]),
		GradientSwingIndicatorPeriod:    convertToInt(dataFrame["GradientSwingIndicatorPeriod"]),
		GradientSwingIndicatorSwingType: convertToInt(dataFrame["GradientSwingIndicatorSwingType"]),

		CatchingFallingKnifeCandles:    convertToInt(dataFrame["CatchingFallingKnifeCandles"]),
		CatchingFallingKnifeSellPercentage:       convertToFloat64(dataFrame["CatchingFallingKnifeSellPercentage"]),
		CatchingFallingKnifeAdditionalBuyPercentage:         convertToFloat64(dataFrame["CatchingFallingKnifeAdditionalBuyPercentage"]),

		TotalMoneyAmount:                    convertToFloat64(dataFrame["TotalMoneyAmount"]),
		TotalMoneyIncreasePercentage:        convertToFloat64(dataFrame["TotalMoneyIncreasePercentage"]),
		FirstBuyMoneyIncreasePercentage:     convertToFloat64(dataFrame["FirstBuyMoneyIncreasePercentage"]),
		StopIncreaseMoneyAfterBuysCount:     convertToInt(dataFrame["StopIncreaseMoneyAfterBuysCount"]),
		Leverage:                            convertToInt(dataFrame["Leverage"]),
		FuturesAvgSellTimeMinutes:           convertToInt(dataFrame["FuturesAvgSellTimeMinutes"]),
		FuturesLeverageActivationPercentage: convertToFloat64(dataFrame["FuturesLeverageActivationPercentage"]),

		LessThanPreviousBuyPercentage: convertToFloat64(dataFrame["LessThanPreviousBuyPercentage"]),
		MoreThanPreviousBuyPercentage: convertToFloat64(dataFrame["MoreThanPreviousBuyPercentage"]),
		ParabolaDivider:               convertToFloat64(dataFrame["ParabolaDivider"]),

		BoostBuyFallPercentage:          convertToFloat64(dataFrame["BoostBuyFallPercentage"]),
		BoostBuyPeriodMinutes:           convertToInt(dataFrame["BoostBuyPeriodMinutes"]),
		BoostBuyMoneyIncreasePercentage: convertToFloat64(dataFrame["BoostBuyMoneyIncreasePercentage"]),

		StopAfterUnsuccessfullySellMinutes: convertToInt(dataFrame["StopAfterUnsuccessfullySellMinutes"]),

		// -----------------------------
		WindowWindowsCount:        convertToInt(dataFrame["WindowWindowsCount"]),
		WindowBasePercentage:      convertToFloat64(dataFrame["WindowBasePercentage"]),
		WindowOffsetPercentage:    convertToFloat64(dataFrame["WindowOffsetPercentage"]),
		WindowBasePeriodMinutes:   convertToInt(dataFrame["WindowBasePeriodMinutes"]),
		WindowOffsetPeriodMinutes: convertToInt(dataFrame["WindowOffsetPeriodMinutes"]),
		// -----------------------------
	}
}

func makeChild(
	maleBotConfig Config,
	femaleBotConfig Config,
) map[string]interface{} {
	childBotConfig := Config{
		HighSellPercentage:         GetFloatFatherOrMomGen(maleBotConfig.HighSellPercentage, femaleBotConfig.HighSellPercentage),
		FirstBuyHighSellPercentage: GetFloatFatherOrMomGen(maleBotConfig.FirstBuyHighSellPercentage, femaleBotConfig.FirstBuyHighSellPercentage),

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

		LinearRegressionCandles:   GetIntFatherOrMomGen(maleBotConfig.LinearRegressionCandles, femaleBotConfig.LinearRegressionCandles),
		LinearRegressionPeriod:    GetIntFatherOrMomGen(maleBotConfig.LinearRegressionPeriod, femaleBotConfig.LinearRegressionPeriod),
		LinearRegressionMse:       GetFloatFatherOrMomGen(maleBotConfig.LinearRegressionMse, femaleBotConfig.LinearRegressionMse),
		LinearRegressionK:         GetFloatFatherOrMomGen(maleBotConfig.LinearRegressionK, femaleBotConfig.LinearRegressionK),
		LinearRegressionDeviation: GetFloatFatherOrMomGen(maleBotConfig.LinearRegressionDeviation, femaleBotConfig.LinearRegressionDeviation),

		GradientSwingIndicatorCandles:   GetIntFatherOrMomGen(maleBotConfig.GradientSwingIndicatorCandles, femaleBotConfig.GradientSwingIndicatorCandles),
		GradientSwingIndicatorPeriod:    GetIntFatherOrMomGen(maleBotConfig.GradientSwingIndicatorPeriod, femaleBotConfig.GradientSwingIndicatorPeriod),
		GradientSwingIndicatorSwingType: GetIntFatherOrMomGen(maleBotConfig.GradientSwingIndicatorSwingType, femaleBotConfig.GradientSwingIndicatorSwingType),

		CatchingFallingKnifeCandles:    GetIntFatherOrMomGen(maleBotConfig.CatchingFallingKnifeCandles, femaleBotConfig.CatchingFallingKnifeCandles),
		CatchingFallingKnifeSellPercentage:       GetFloatFatherOrMomGen(maleBotConfig.CatchingFallingKnifeSellPercentage, femaleBotConfig.CatchingFallingKnifeSellPercentage),
		CatchingFallingKnifeAdditionalBuyPercentage:         GetFloatFatherOrMomGen(maleBotConfig.CatchingFallingKnifeAdditionalBuyPercentage, femaleBotConfig.CatchingFallingKnifeAdditionalBuyPercentage),

		TotalMoneyAmount:                    GetFloatFatherOrMomGen(maleBotConfig.TotalMoneyAmount, femaleBotConfig.TotalMoneyAmount),
		TotalMoneyIncreasePercentage:        GetFloatFatherOrMomGen(maleBotConfig.TotalMoneyIncreasePercentage, femaleBotConfig.TotalMoneyIncreasePercentage),
		FirstBuyMoneyIncreasePercentage:     GetFloatFatherOrMomGen(maleBotConfig.FirstBuyMoneyIncreasePercentage, femaleBotConfig.FirstBuyMoneyIncreasePercentage),
		StopIncreaseMoneyAfterBuysCount:     GetIntFatherOrMomGen(maleBotConfig.StopIncreaseMoneyAfterBuysCount, femaleBotConfig.StopIncreaseMoneyAfterBuysCount),
		Leverage:                            GetIntFatherOrMomGen(maleBotConfig.Leverage, femaleBotConfig.Leverage),
		FuturesAvgSellTimeMinutes:           GetIntFatherOrMomGen(maleBotConfig.FuturesAvgSellTimeMinutes, femaleBotConfig.FuturesAvgSellTimeMinutes),
		FuturesLeverageActivationPercentage: GetFloatFatherOrMomGen(maleBotConfig.FuturesLeverageActivationPercentage, femaleBotConfig.FuturesLeverageActivationPercentage),

		LessThanPreviousBuyPercentage: GetFloatFatherOrMomGen(maleBotConfig.LessThanPreviousBuyPercentage, femaleBotConfig.LessThanPreviousBuyPercentage),
		MoreThanPreviousBuyPercentage: GetFloatFatherOrMomGen(maleBotConfig.MoreThanPreviousBuyPercentage, femaleBotConfig.MoreThanPreviousBuyPercentage),
		ParabolaDivider:               GetFloatFatherOrMomGen(maleBotConfig.ParabolaDivider, femaleBotConfig.ParabolaDivider),

		BoostBuyFallPercentage:          GetFloatFatherOrMomGen(maleBotConfig.BoostBuyFallPercentage, femaleBotConfig.BoostBuyFallPercentage),
		BoostBuyPeriodMinutes:           GetIntFatherOrMomGen(maleBotConfig.BoostBuyPeriodMinutes, femaleBotConfig.BoostBuyPeriodMinutes),
		BoostBuyMoneyIncreasePercentage: GetFloatFatherOrMomGen(maleBotConfig.BoostBuyMoneyIncreasePercentage, femaleBotConfig.BoostBuyMoneyIncreasePercentage),

		StopAfterUnsuccessfullySellMinutes: GetIntFatherOrMomGen(maleBotConfig.StopAfterUnsuccessfullySellMinutes, femaleBotConfig.StopAfterUnsuccessfullySellMinutes),

		// -------------------------------
		WindowWindowsCount:        GetIntFatherOrMomGen(maleBotConfig.WindowWindowsCount, femaleBotConfig.WindowWindowsCount),
		WindowBasePercentage:      GetFloatFatherOrMomGen(maleBotConfig.WindowBasePercentage, femaleBotConfig.WindowBasePercentage),
		WindowOffsetPercentage:    GetFloatFatherOrMomGen(maleBotConfig.WindowOffsetPercentage, femaleBotConfig.WindowOffsetPercentage),
		WindowBasePeriodMinutes:   GetIntFatherOrMomGen(maleBotConfig.WindowBasePeriodMinutes, femaleBotConfig.WindowBasePeriodMinutes),
		WindowOffsetPeriodMinutes: GetIntFatherOrMomGen(maleBotConfig.WindowOffsetPeriodMinutes, femaleBotConfig.WindowOffsetPeriodMinutes),
		// -------------------------------
	}

	for i := 0; i < 35; i++ {
		mutateGens(&childBotConfig, GetRandInt(0, 40))
	}

	return GetBotConfigMapInterface(childBotConfig)
}

func mutateGens(botConfig *Config, randGenNumber int) {
	restrict := GetBotConfigRestrictions()

	mutateGenFloat64(randGenNumber, 0, &(botConfig.HighSellPercentage), restrict.HighSellPercentage)
	mutateGenFloat64(randGenNumber, 1, &(botConfig.FirstBuyHighSellPercentage), restrict.FirstBuyHighSellPercentage)

	mutateGenFloat64(randGenNumber, 2, &(botConfig.TrailingTopPercentage), restrict.TrailingTopPercentage)
	mutateGenInt(randGenNumber, 3, &(botConfig.TrailingUpdateTimesBeforeFinish), restrict.TrailingUpdateTimesBeforeFinish)

	mutateGenInt(randGenNumber, 4, &(botConfig.WaitAfterLastBuyPeriod), restrict.WaitAfterLastBuyPeriod)

	mutateGenInt(randGenNumber, 5, &(botConfig.BigFallCandlesCount), restrict.BigFallCandlesCount)
	mutateGenInt(randGenNumber, 6, &(botConfig.BigFallSmoothPeriod), restrict.BigFallSmoothPeriod)
	mutateGenFloat64(randGenNumber, 7, &(botConfig.BigFallPercentage), restrict.BigFallPercentage)

	mutateGenInt(randGenNumber, 8, &(botConfig.DesiredPriceCandles), restrict.DesiredPriceCandles)

	mutateGenInt(randGenNumber, 9, &(botConfig.GradientDescentCandles), restrict.GradientDescentCandles)
	mutateGenInt(randGenNumber, 10, &(botConfig.GradientDescentPeriod), restrict.GradientDescentPeriod)
	mutateGenFloat64(randGenNumber, 11, &(botConfig.GradientDescentGradient), restrict.GradientDescentGradient)

	mutateGenFloat64(randGenNumber, 12, &(botConfig.TrailingSellActivationAdditionPercentage), restrict.TrailingSellActivationAdditionPercentage)
	mutateGenFloat64(randGenNumber, 13, &(botConfig.TrailingSellStopPercentage), restrict.TrailingSellStopPercentage)

	mutateGenInt(randGenNumber, 14, &(botConfig.LinearRegressionCandles), restrict.LinearRegressionCandles)
	mutateGenInt(randGenNumber, 15, &(botConfig.LinearRegressionPeriod), restrict.LinearRegressionPeriod)
	mutateGenFloat64(randGenNumber, 16, &(botConfig.LinearRegressionMse), restrict.LinearRegressionMse)
	mutateGenFloat64(randGenNumber, 17, &(botConfig.LinearRegressionK), restrict.LinearRegressionK)
	mutateGenFloat64(randGenNumber, 18, &(botConfig.LinearRegressionDeviation), restrict.LinearRegressionDeviation)

	mutateGenInt(randGenNumber, 19, &(botConfig.GradientSwingIndicatorCandles), restrict.GradientSwingIndicatorCandles)
	mutateGenInt(randGenNumber, 20, &(botConfig.GradientSwingIndicatorPeriod), restrict.GradientSwingIndicatorPeriod)
	mutateGenInt(randGenNumber, 21, &(botConfig.GradientSwingIndicatorSwingType), restrict.GradientSwingIndicatorSwingType)


	mutateGenInt(randGenNumber, 22, &(botConfig.CatchingFallingKnifeCandles), restrict.CatchingFallingKnifeCandles)
	mutateGenFloat64(randGenNumber, 23, &(botConfig.CatchingFallingKnifeSellPercentage), restrict.CatchingFallingKnifeSellPercentage)
	mutateGenFloat64(randGenNumber, 24, &(botConfig.CatchingFallingKnifeAdditionalBuyPercentage), restrict.CatchingFallingKnifeAdditionalBuyPercentage)


	mutateGenFloat64(randGenNumber, 25, &(botConfig.TotalMoneyAmount), restrict.TotalMoneyAmount)
	mutateGenFloat64(randGenNumber, 26, &(botConfig.TotalMoneyIncreasePercentage), restrict.TotalMoneyIncreasePercentage)
	mutateGenFloat64(randGenNumber, 27, &(botConfig.FirstBuyMoneyIncreasePercentage), restrict.FirstBuyMoneyIncreasePercentage)
	mutateGenInt(randGenNumber, 28, &(botConfig.StopIncreaseMoneyAfterBuysCount), restrict.StopIncreaseMoneyAfterBuysCount)
	mutateGenInt(randGenNumber, 29, &(botConfig.Leverage), restrict.Leverage)
	mutateGenInt(randGenNumber, 30, &(botConfig.FuturesAvgSellTimeMinutes), restrict.FuturesAvgSellTimeMinutes)
	mutateGenFloat64(randGenNumber, 31, &(botConfig.FuturesLeverageActivationPercentage), restrict.FuturesLeverageActivationPercentage)

	mutateGenFloat64(randGenNumber, 32, &(botConfig.LessThanPreviousBuyPercentage), restrict.LessThanPreviousBuyPercentage)
	mutateGenFloat64(randGenNumber, 33, &(botConfig.MoreThanPreviousBuyPercentage), restrict.MoreThanPreviousBuyPercentage)
	mutateGenFloat64(randGenNumber, 34, &(botConfig.ParabolaDivider), restrict.ParabolaDivider)

	mutateGenFloat64(randGenNumber, 35, &(botConfig.BoostBuyFallPercentage), restrict.BoostBuyFallPercentage)
	mutateGenInt(randGenNumber, 36, &(botConfig.BoostBuyPeriodMinutes), restrict.BoostBuyPeriodMinutes)
	mutateGenFloat64(randGenNumber, 37, &(botConfig.BoostBuyMoneyIncreasePercentage), restrict.BoostBuyMoneyIncreasePercentage)

	mutateGenInt(randGenNumber, 38, &(botConfig.StopAfterUnsuccessfullySellMinutes), restrict.StopAfterUnsuccessfullySellMinutes)

	// --------------------------------------
	mutateGenInt(randGenNumber, 39, &(botConfig.WindowWindowsCount), restrict.WindowWindowsCount)
	mutateGenFloat64(randGenNumber, 40, &(botConfig.WindowBasePercentage), restrict.WindowBasePercentage)
	mutateGenFloat64(randGenNumber, 41, &(botConfig.WindowOffsetPercentage), restrict.WindowOffsetPercentage)
	mutateGenInt(randGenNumber, 42, &(botConfig.WindowBasePeriodMinutes), restrict.WindowBasePeriodMinutes)
	mutateGenInt(randGenNumber, 43, &(botConfig.WindowOffsetPeriodMinutes), restrict.WindowOffsetPeriodMinutes)
	// --------------------------------------
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
