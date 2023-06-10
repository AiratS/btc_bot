package main

type ConfigRestriction struct {
	HighSellPercentage         MinMaxFloat64
	FirstBuyHighSellPercentage MinMaxFloat64

	TrailingTopPercentage           MinMaxFloat64
	TrailingUpdateTimesBeforeFinish MinMaxInt

	WaitAfterLastBuyPeriod MinMaxInt

	BigFallCandlesCount MinMaxInt
	BigFallSmoothPeriod MinMaxInt
	BigFallPercentage   MinMaxFloat64

	DesiredPriceCandles MinMaxInt

	GradientDescentCandles  MinMaxInt
	GradientDescentPeriod   MinMaxInt
	GradientDescentGradient MinMaxFloat64

	TrailingSellActivationAdditionPercentage MinMaxFloat64
	TrailingSellStopPercentage               MinMaxFloat64

	LinearRegressionCandles   MinMaxInt
	LinearRegressionPeriod    MinMaxInt
	LinearRegressionMse       MinMaxFloat64
	LinearRegressionK         MinMaxFloat64
	LinearRegressionDeviation MinMaxFloat64

	GradientSwingIndicatorCandles   MinMaxInt
	GradientSwingIndicatorPeriod    MinMaxInt
	GradientSwingIndicatorSwingType MinMaxInt

	TotalMoneyAmount                    MinMaxFloat64
	TotalMoneyIncreasePercentage        MinMaxFloat64
	FirstBuyMoneyIncreasePercentage     MinMaxFloat64
	StopIncreaseMoneyAfterBuysCount     MinMaxInt
	Leverage                            MinMaxInt
	FuturesAvgSellTimeMinutes           MinMaxInt
	FuturesLeverageActivationPercentage MinMaxFloat64

	LessThanPreviousBuyPercentage MinMaxFloat64
	MoreThanPreviousBuyPercentage MinMaxFloat64
	ParabolaDivider               MinMaxFloat64

	BoostBuyFallPercentage          MinMaxFloat64
	BoostBuyPeriodMinutes           MinMaxInt
	BoostBuyMoneyIncreasePercentage MinMaxFloat64

	StopAfterUnsuccessfullySellMinutes MinMaxInt
}

type MinMaxInt struct {
	min int
	max int
}

type MinMaxFloat64 struct {
	min float64
	max float64
}

func GetBotConfigRestrictions() ConfigRestriction {
	return ConfigRestriction{
		HighSellPercentage: MinMaxFloat64{
			min: 0.12,
			max: 0.3,
		},
		FirstBuyHighSellPercentage: MinMaxFloat64{
			min: 0.2,
			max: 1,
		},

		TrailingTopPercentage: MinMaxFloat64{
			min: 0.2,
			max: 0.5,
		},
		TrailingUpdateTimesBeforeFinish: MinMaxInt{
			min: 1,
			max: 3,
		},

		WaitAfterLastBuyPeriod: MinMaxInt{
			min: 1,
			max: 30,
		},

		BigFallCandlesCount: MinMaxInt{
			min: 3,
			max: 10,
		},
		BigFallSmoothPeriod: MinMaxInt{
			min: 4,
			max: 10,
		},
		BigFallPercentage: MinMaxFloat64{
			min: 0.25,
			max: 0.65,
		},

		DesiredPriceCandles: MinMaxInt{
			min: 24 * 20 * 1,
			max: 24 * 20 * 7,
		},

		GradientDescentCandles: MinMaxInt{
			min: 10,
			max: 30,
		},
		GradientDescentPeriod: MinMaxInt{
			min: 4,
			max: 25,
		},
		GradientDescentGradient: MinMaxFloat64{
			min: 0,
			max: 5,
		},

		TrailingSellActivationAdditionPercentage: MinMaxFloat64{
			min: 0.4,
			max: 1,
		},
		TrailingSellStopPercentage: MinMaxFloat64{
			min: 0.4,
			max: 2,
		},

		// ------------------------------------------------
		LinearRegressionCandles: MinMaxInt{
			min: 5,
			max: 25,
		},
		LinearRegressionPeriod: MinMaxInt{
			min: 1,
			max: 25,
		},
		LinearRegressionMse: MinMaxFloat64{
			min: 100,
			max: 2000,
		},
		LinearRegressionK: MinMaxFloat64{
			min: 0,
			max: 2,
		},
		LinearRegressionDeviation: MinMaxFloat64{
			min: 5,
			max: 30,
		},

		// ------------------------------------------------
		GradientSwingIndicatorCandles: MinMaxInt{
			min: 6,
			max: 30,
		},
		GradientSwingIndicatorPeriod: MinMaxInt{
			min: 3,
			max: 15,
		},
		GradientSwingIndicatorSwingType: MinMaxInt{
			min: 0,
			max: 2,
		},

		// ------------------------------------------------
		TotalMoneyAmount: MinMaxFloat64{
			min: 3,
			max: 200,
		},
		TotalMoneyIncreasePercentage: MinMaxFloat64{
			min: 30,
			max: 100,
		},
		FirstBuyMoneyIncreasePercentage: MinMaxFloat64{
			min: 0,
			max: 0,
		},
		StopIncreaseMoneyAfterBuysCount: MinMaxInt{
			min: 3,
			max: 10,
		},
		Leverage: MinMaxInt{
			min: 2,
			max: 10,
		},

		// ------------------------------------------------
		FuturesAvgSellTimeMinutes: MinMaxInt{
			min: 60 * 1,       // 1 hour
			max: 60 * 24 * 14, // 5 days
		},
		FuturesLeverageActivationPercentage: MinMaxFloat64{
			min: 10,
			max: 300,
		},

		LessThanPreviousBuyPercentage: MinMaxFloat64{
			min: -1,
			max: -0.15,
		},
		MoreThanPreviousBuyPercentage: MinMaxFloat64{
			min: 0,
			max: 1,
		},
		ParabolaDivider: MinMaxFloat64{
			min: 0.2,
			max: 1.5,
		},

		BoostBuyFallPercentage: MinMaxFloat64{
			min: 0.5,
			max: 2,
		},
		BoostBuyPeriodMinutes: MinMaxInt{
			min: 20,
			max: 100,
		},
		BoostBuyMoneyIncreasePercentage: MinMaxFloat64{
			min: 100,
			max: 2000,
		},

		// ------------------------------------------------
		StopAfterUnsuccessfullySellMinutes: MinMaxInt{
			min: 60 * 1,       // 1 hour
			max: 60 * 24 * 14, // 5 days
		},
	}
}
