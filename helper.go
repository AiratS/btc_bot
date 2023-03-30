package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func FileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func FormatTimestamp(timestamp int64) string {
	date := ParseMilliTimestamp(timestamp)
	//date = date.Add(time.Hour * 3)
	return date.Format("2006-01-02 15:04:05")
}

func ParseMilliTimestamp(tm int64) time.Time {
	sec := tm / 1000
	msec := tm % 1000
	return time.Unix(sec, msec*int64(time.Millisecond))
}

func Min(values []float64) float64 {
	if 0 == len(values) {
		panic("No values for MIN function")
	}

	min := values[0]

	for _, value := range values {
		if value < min {
			min = value
		}
	}

	return min
}

func Max(values []float64) float64 {
	if 0 == len(values) {
		panic("No values for MAX function")
	}

	max := values[0]

	for _, value := range values {
		if value > max {
			max = value
		}
	}

	return max
}

func MaxInt(values []int) int {
	if 0 == len(values) {
		panic("No values for MAX function")
	}

	max := values[0]

	for _, value := range values {
		if value > max {
			max = value
		}
	}

	return max
}

func BuySliceIntersect(buys1, buys2 []Buy) []Buy {
	var result []Buy

	for _, buy1 := range buys1 {
		for _, buy2 := range buys2 {
			if buy1.Id == buy2.Id {
				result = append(result, buy1)
				break
			}
		}
	}

	return result
}

func CalcGrowth(startPrice, endPrice float64) float64 {
	if startPrice == 0 || endPrice == 0 {
		return 0.0
	}

	return ((endPrice * 100) / startPrice) - 100
}

func GetClosePrice(candles []Candle, count int) []float64 {
	var prices []float64

	firstIdx := getKlineCandleListFirstIdx(&candles, count)
	lastIdx := getKlineCandleListLastIdx(&candles)

	for _, candle := range candles[firstIdx:lastIdx] {
		prices = append(prices, candle.ClosePrice)
	}

	return prices
}

func getKlineCandleListLastIdx(candles *[]Candle) int {
	return len(*candles) - 1
}

func getKlineCandleListFirstIdx(candles *[]Candle, candlesCount int) int {
	firstIdx := len(*candles) - candlesCount - 1
	if firstIdx < 0 {
		return 0
	}

	return firstIdx
}

func Log(msg string) {
	if IS_REAL_ENABLED {
		LogAndPrintAndSendTg(msg)
	} else {
		LogAndPrint(msg + "\n------------------------------\n")
	}
}

func LogAndPrint(msg string) {
	fmt.Println(msg)
	log.Println(msg)
}

func LogAndPrintAndSendTg(msg string) {
	fmt.Println(msg)
	log.Println(msg)
	SendTgBotMessage(msg)
}

func GetRandIntConfig(minMax MinMaxInt) int {
	return GetRandInt(minMax.min, minMax.max)
}

func GetRandFloat64Config(minMax MinMaxFloat64) float64 {
	return GetRandFloat64(minMax.min, minMax.max)
}

func GetRandInt(lower int, upper int) int {
	rand.Seed(time.Now().UnixNano())
	return lower + rand.Intn(upper-lower+1)
}

func GetRandFloat64(lower float64, upper float64) float64 {
	rand.Seed(time.Now().UnixNano())
	return lower + rand.Float64()*(upper-lower)
}

func convertStringToFloat64(typeValue string) float64 {
	value, _ := strconv.ParseFloat(typeValue, 64)
	return value
}

func convertStringToInt(typeValue string) int {
	value, _ := strconv.ParseInt(typeValue, 10, 64)
	return int(value)
}

func convertToInt(value interface{}) int {
	switch typeValue := value.(type) {
	case int64:
		return int(typeValue)
	}
	return 0
}

func convertToFloat64(value interface{}) float64 {
	switch typeValue := value.(type) {
	case float64:
		return float64(typeValue)
	}
	return math.NaN()
}

func convertBinanceToFloat64(value interface{}) float64 {
	val, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	if err != nil {
		panic(err)
	}
	return val
}

func convertBinanceToInt(value interface{}) int {
	val, err := strconv.Atoi(value.(string))
	if err != nil {
		panic(err)
	}
	return val
}

func CountInArray(needle float64, array *[]float64) int {
	count := 0
	searchArray := *array
	for _, element := range searchArray {
		if needle == element {
			count++
		}
	}
	return count
}

func ConvertDateStringToTime(dateString string) time.Time {
	layout := "2006-01-02 15:04:05"
	parsedTime, _ := time.Parse(layout, dateString)
	return parsedTime
}

func GetCurrentMinusTime(candleTime time.Time, durationRaw int) time.Time {
	duration := time.Minute
	if CANDLE_INTERVAL == "1s" {
		duration = time.Second
	}

	candleTime = candleTime.Add(-duration * time.Duration(durationRaw))

	return candleTime
}

func CalcUpperPrice(price, percentage float64) float64 {
	return price + ((price * percentage) / 100)
}

func CalcBottomPrice(price, percentage float64) float64 {
	return price - ((price * percentage) / 100)
}

func GetOpenPrices(candles []Candle) []float64 {
	var values []float64

	for _, candle := range candles {
		values = append(values, candle.GetPrice())
	}

	return values
}

func GetClosePrices(candles []Candle) []float64 {
	var values []float64

	for _, candle := range candles {
		values = append(values, candle.GetPrice())
	}

	return values
}

func GetAvg(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}

	return total / float64(len(values))
}

func Median(data []float64) float64 {
	dataCopy := make([]float64, len(data))
	copy(dataCopy, data)

	sort.Float64s(dataCopy)

	var median float64
	l := len(dataCopy)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (dataCopy[l/2-1] + dataCopy[l/2]) / 2
	} else {
		median = dataCopy[l/2]
	}

	return median
}

func FilterZeroPrices(prices []float64) []float64 {
	var values []float64

	for _, price := range prices {
		if 0 < price {
			values = append(values, price)
		}
	}

	return values
}

func Sum(values []float64) float64 {
	result := 0.0

	for _, value := range values {
		result += value
	}

	return result
}

func Int64SliceToStringSlice(items []int64) []string {
	var stringItems []string

	for _, item := range items {
		stringItems = append(stringItems, fmt.Sprintf("%d", item))
	}

	return stringItems
}

func JoinInt64(items []int64) string {
	return strings.Join(Int64SliceToStringSlice(items), ", ")
}

func BuildPlots() {
	files, _ := filepath.Glob("plots/*.png")
	for _, file := range files {
		os.Remove(file)
	}

	// Create plots
	app := "python3.9"
	arg0 := "main.py"

	cmd := exec.Command(app, arg0)
	_, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func GetLeverageLiquidationPercentage() float64 {
	if LEVERAGE == 1 {
		return 100
	}

	if LEVERAGE == 2 {
		return 50
	}

	if LEVERAGE == 3 {
		return 33
	}

	if LEVERAGE == 4 {
		return 25
	}

	if LEVERAGE == 5 {
		return 20
	}

	if LEVERAGE == 6 {
		return 16.5
	}

	if LEVERAGE == 7 {
		return 14.3
	}

	if LEVERAGE == 8 {
		return 12.3
	}

	if LEVERAGE == 9 {
		return 11.2
	}

	if LEVERAGE == 10 {
		return 10
	}

	return 100
}
