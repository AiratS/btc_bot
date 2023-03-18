package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type PlotData struct {
	BuyId    int64  `json:"buyId"`
	BuyTime  string `json:"buyTime"`
	SellTime string `json:"sellTime"`
}

var plots map[int64]*PlotData

func PlotAddBuy(buyId int64, buyTime string) {
	if 1 < BOTS_COUNT {
		return
	}

	initPlots()

	plots[buyId] = &PlotData{
		BuyId:   buyId,
		BuyTime: buyTime,
	}
}

func PlotAddSell(buyId int64, sellTime string) {
	if 1 < BOTS_COUNT {
		return
	}

	initPlots()

	if item, ok := plots[buyId]; ok {
		item.SellTime = sellTime
	}
}

func PlotToJson(fileName string) {
	if 1 < BOTS_COUNT {
		return
	}

	var buys []PlotData
	for _, plotData := range plots {
		buys = append(buys, *plotData)
	}

	jsonStr, err := json.Marshal(buys)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		_ = ioutil.WriteFile(fileName, jsonStr, 0644)
	}
}

func initPlots() {
	if 0 == len(plots) {
		plots = map[int64]*PlotData{}
	}
}
