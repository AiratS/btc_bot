package main

type ConfigRestriction struct {
	HighSellPercentage MinMaxFloat64

	TrailingTopPercentage           MinMaxFloat64
	TrailingUpdateTimesBeforeFinish MinMaxInt

	WaitAfterLastBuyPeriod MinMaxInt

	BigFallCandlesCount MinMaxInt
	BigFallPercentage   MinMaxFloat64

	DesiredPriceCandles MinMaxInt

	GradientDescentCandles  MinMaxInt
	GradientDescentPeriod   MinMaxInt
	GradientDescentGradient MinMaxFloat64
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
			min: 0.3,
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
			min: 4,
			max: 15,
		},

		BigFallCandlesCount: MinMaxInt{
			min: 4,
			max: 15,
		},
		BigFallPercentage: MinMaxFloat64{
			min: 0.3,
			max: 1,
		},

		DesiredPriceCandles: MinMaxInt{
			min: 24 * 20 * 1,
			max: 24 * 20 * 7,
		},

		GradientDescentCandles: MinMaxInt{
			min: 40,
			max: 60,
		},
		GradientDescentPeriod: MinMaxInt{
			min: 20,
			max: 40,
		},
		GradientDescentGradient: MinMaxFloat64{
			min: 0,
			max: 5,
		},
	}
}
