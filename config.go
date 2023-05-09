package main

// Main
const IS_REAL_ENABLED = false
const ENABLE_FUTURES = true
const USE_REAL_MONEY = false

const (
	BUY_ORDER_REDUCTION_ENABLED      = false
	BUY_ORDER_REDUCTION_PERCENTAGE   = 0.1
	BUY_ORDER_REJECTION_TIME_MINUTES = 15
)

const REAL_MONEY_DB_NAME = "amazing_real"

//const ADD_REVENUE_TO_BALANCE = false

// Candle
const CANDLE_SYMBOL = "BTCUSDT"
const CANDLE_INTERVAL = "30m"
const BALANCE_MONEY = 3000.0
const COMMISSION = 0.06
const DATASETS_DIRECTORY = "datasets"
const UNSOLD_BUYS_COUNT = 20
const ENABLE_STOP_INCREASE_AFTER_BUYS_COUNT = false
const ENABLE_BOOST_BUY_INDICATOR = false
const ENABLE_FIRST_BUY_HIGER_SELL_PERCENTAGE = true

// Genetic
const NO_VALIDATION = true
const BOTS_COUNT = 25
const BEST_BOTS_COUNT = 7
const BEST_BOTS_FROM_PREV_GEN = 3
const GENERATION_COUNT = 20
const DEFAULT_REVENUE = -1000000
const ENABLE_AVG_TIME = false
const SELL_TIME_PUNISHMENT = 1.0

const ENABLE_TIME_CANCEL = false

type Config struct {
	HighSellPercentage         float64
	FirstBuyHighSellPercentage float64

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

	LinearRegressionCandles int
	LinearRegressionPeriod  int
	LinearRegressionMse     float64
	LinearRegressionK       float64

	TotalMoneyAmount                    float64
	TotalMoneyIncreasePercentage        float64
	FirstBuyMoneyIncreasePercentage     float64
	StopIncreaseMoneyAfterBuysCount     int
	Leverage                            int
	FuturesAvgSellTimeMinutes           int
	FuturesLeverageActivationPercentage float64

	LessThanPreviousBuyPercentage float64

	BoostBuyFallPercentage          float64
	BoostBuyPeriodMinutes           int
	BoostBuyMoneyIncreasePercentage float64

	StopAfterUnsuccessfullySellMinutes int

	TotalRevenue     float64
	FinalBalance     float64
	FinalRevenue     float64
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
