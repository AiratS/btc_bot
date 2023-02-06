package main

import (
	"os"
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
