package main

type ConfigRestriction struct {
	HighSellPercentage MinMaxFloat64

	TrailingTopPercentage           MinMaxFloat64
	TrailingUpdateTimesBeforeFinish MinMaxInt

	WaitAfterLastBuyPeriod MinMaxInt
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
			min: 0.5,
			max: 1,
		},

		TrailingTopPercentage: MinMaxFloat64{
			min: 0.7,
			max: 6,
		},
		TrailingUpdateTimesBeforeFinish: MinMaxInt{
			min: 1,
			max: 2,
		},

		WaitAfterLastBuyPeriod: MinMaxInt{
			min: 120,
			max: 1000,
		},
	}
}
