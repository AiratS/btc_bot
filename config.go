package main

// Main
const IS_REAL_ENABLED = false
const ENABLE_FUTURES = true
const USE_REAL_MONEY = false
const REAL_MONEY_DB_NAME = "amazing_real"

// Candle
const CANDLE_SYMBOL = "BTCUSDT"
const CANDLE_INTERVAL = "30m"
const BALANCE_MONEY = 1000.0
const COMMISSION = 0.06
const DATASETS_DIRECTORY = "datasets"
const UNSOLD_BUYS_COUNT = 20

// Genetic
const NO_VALIDATION = true
const BOTS_COUNT = 25
const BEST_BOTS_COUNT = 7
const BEST_BOTS_FROM_PREV_GEN = 3
const GENERATION_COUNT = 20
const DEFAULT_REVENUE = -1000000
const ENABLE_AVG_TIME = true
const SELL_TIME_PUNISHMENT = 1.0

const ENABLE_TIME_CANCEL = true

type Config struct {
	HighSellPercentage float64

	TrailingTopPercentage           float64
	TrailingUpdateTimesBeforeFinish int

	WaitAfterLastBuyPeriod int

	BigFallCandlesCount int
	BigFallSmoothPeriod int
	BigFallPercentage   float64

	DesiredPriceCandles int

	GradientDescentCandles  int
	GradientDescentPeriod   int
	GradientDescentGradient float64

	TrailingSellActivationAdditionPercentage float64
	TrailingSellStopPercentage               float64

	TotalMoneyAmount                    float64
	Leverage                            int
	FuturesAvgSellTimeMinutes           int
	FuturesLeverageActivationPercentage float64

	TotalRevenue     float64
	TotalBuysCount   int
	UnsoldBuysCount  int
	LiquidationCount int
	AvgSellTime      float64

	ValidationTotalRevenue     float64
	ValidationTotalBuysCount   int
	ValidationUnsoldBuysCount  int
	ValidationLiquidationCount int
	ValidationAvgSellTime      float64

	Selection float64
}
