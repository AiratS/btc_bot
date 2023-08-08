package main

import (
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"os"
)

func GetDatasetDates() []string {
	return []string{
		// Learn
		//"2022-07",
		//"2022-08",

		// Test
		//"2019-01",
		//"2019-02",
		//"2019-03",
		//"2019-04",
		//"2019-05",
		//"2019-06",
		//"2019-07",
		//"2019-08",
		//"2019-09",
		//"2019-10",
		//"2019-11",
		//"2019-12",

		// Own
		"2022-09",
		"2022-10",
		"2022-11",
		"2022-12",
	}
}

func GetValidationDatasetDates() []string {
	return []string{
		"2022-11",
		"2022-12",
	}
}

func ImportDatasets(dates []string) *[]Candle {
	var importedCandles *[]Candle
	candles := []Candle{}

	for _, date := range dates {
		fileName := fmt.Sprintf("%s/%s-%s-%s.csv", DATASETS_DIRECTORY, CANDLE_SYMBOL, CANDLE_INTERVAL, date)
		if !FileExists(fileName) {
			panic(fmt.Sprintf("No dataset for date: %s", fileName))
		}

		fileCandles := CsvFileToCandles(fileName)
		firstC := fileCandles[0]
		lastC := fileCandles[len(fileCandles)-1]

		fmt.Println(firstC, lastC)

		for _, aCandles := range fileCandles {
			candles = append(candles, aCandles)
		}
	}

	importedCandles = &candles

	return importedCandles
}

func CsvFileToCandles(fileName string) []Candle {
	var candles []Candle

	file, err := os.Open(fileName)
	csvDataFrame := dataframe.ReadCSV(file)

	fmt.Println(err)

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
