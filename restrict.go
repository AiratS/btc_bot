package main

type ConfigRestriction struct {
	HighSellPercentage MinMaxFloat64

	TrailingTopPercentage           MinMaxFloat64
	TrailingUpdateTimesBeforeFinish MinMaxInt
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
			min: .5,
			max: 7,
		},

		TrailingTopPercentage: MinMaxFloat64{
			min: 6,
			max: 1,
		},
		TrailingUpdateTimesBeforeFinish: MinMaxInt{
			min: 1,
			max: 2,
		},
	}
}
