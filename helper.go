package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
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

func GetCurrentMinusTime(candleTime time.Time, minutes int) time.Time {
	candleTime = candleTime.Add(-time.Minute * time.Duration(minutes))

	return candleTime
}
