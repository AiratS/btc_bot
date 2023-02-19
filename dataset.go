package main

import (
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"os"
)

var importedCandles *[]Candle

func GetDatasetDates() []string {
	return []string{
		//"2021-10",
		//"2022-04",
		"2023-01",
	}
}

func ImportDatasets() *[]Candle {
	candles := []Candle{}

	for _, date := range GetDatasetDates() {
		fileName := fmt.Sprintf("%s/%s-%s-%s.csv", DATASETS_DIRECTORY, CANDLE_SYMBOL, CANDLE_INTERVAL, date)
		if !FileExists(fileName) {
			panic(fmt.Sprintf("No dataset for date: %s", fileName))
		}

		candles = append(candles, CsvFileToCandles(fileName)...)
	}

	importedCandles = &candles

	return importedCandles
}

func CsvFileToCandles(fileName string) []Candle {
	var candles []Candle

	file, _ := os.Open(fileName)
	csvDataFrame := dataframe.ReadCSV(file)

	for i := 0; i < csvDataFrame.Nrow(); i++ {
		row := csvDataFrame.Subset(i)
		candles = append(candles, CsvRowToCandle(row, CANDLE_SYMBOL, i))
	}

	return candles
}

func CsvRowToCandle(candleDataFrame dataframe.DataFrame, symbol string, index int) Candle {
	firstRow := 0
	openTime, _ := candleDataFrame.Elem(firstRow, OPEN_TIME).Int()
	closeTime, _ := candleDataFrame.Elem(firstRow, CLOSE_TIME).Int()

	return Candle{
		Symbol:                   symbol,
		OpenTime:                 FormatTimestamp(int64(openTime)),
		OpenPrice:                candleDataFrame.Elem(firstRow, OPEN_PRICE).Float(),
		HighPrice:                candleDataFrame.Elem(firstRow, HIGH_PRICE).Float(),
		LowPrice:                 candleDataFrame.Elem(firstRow, LOW_PRICE).Float(),
		ClosePrice:               candleDataFrame.Elem(firstRow, CLOSE_PRICE).Float(),
		Volume:                   candleDataFrame.Elem(firstRow, VOLUME).Float(),
		CloseTime:                FormatTimestamp(int64(closeTime)),
		QuoteAssetVolume:         candleDataFrame.Elem(firstRow, QUOTE_ASSET_VOLUME).Float(),
		NumberOfTrades:           0,
		TakerBuyBaseAssetVolume:  candleDataFrame.Elem(firstRow, TAKER_BUY_BASE_ASSET_VOLUME).Float(),
		TakerBuyQuoteAssetVolume: candleDataFrame.Elem(firstRow, TAKER_BUY_QUOTE_ASSET_VOLUME).Float(),
		Ignore:                   0,
	}
}
