package main

type ConfigRestriction struct {
	HighSellPercentage MinMaxFloat64

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

	TotalMoneyAmount                    MinMaxFloat64
	Leverage                            MinMaxInt
	FuturesAvgSellTimeMinutes           MinMaxInt
	FuturesLeverageActivationPercentage MinMaxFloat64
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
			min: 10,
			max: 15,
		},
		BigFallSmoothPeriod: MinMaxInt{
			min: 4,
			max: 10,
		},
		BigFallPercentage: MinMaxFloat64{
			min: 0.2,
			max: 2,
		},

		DesiredPriceCandles: MinMaxInt{
			min: 24 * 20 * 1,
			max: 24 * 20 * 7,
		},

		GradientDescentCandles: MinMaxInt{
			min: 6,
			max: 60,
		},
		GradientDescentPeriod: MinMaxInt{
			min: 1,
			max: 6,
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

		TotalMoneyAmount: MinMaxFloat64{
			min: 1000,
			max: 1000,
		},
		Leverage: MinMaxInt{
			min: 10,
			max: 10,
		},
		FuturesAvgSellTimeMinutes: MinMaxInt{
			min: 60 * 1,       // 1 hour
			max: 60 * 24 * 14, // 5 days
		},
		FuturesLeverageActivationPercentage: MinMaxFloat64{
			min: 10,
			max: 300,
		},
	}
}
