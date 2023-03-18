package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type PlotData struct {
	BuyId        int64                  `json:"buyId"`
	BuyTime      string                 `json:"buyTime"`
	SellTime     string                 `json:"sellTime"`
	TrailingSell []TrailingSellPlotData `json:"trailingSell"`
}

type TrailingSellPlotData struct {
	Time      string  `json:"time"`
	StopPrice float64 `json:"stopPrice"`
}

var plots map[int64]*PlotData

func PlotAddBuy(buyId int64, buyTime string) {
	if !canPlot() {
		return
	}

	initPlots()

	plots[buyId] = &PlotData{
		BuyId:   buyId,
		BuyTime: buyTime,
	}
}

func PlotAddSell(buyId int64, sellTime string) {
	if !canPlot() {
		return
	}

	if item, ok := plots[buyId]; ok {
		item.SellTime = sellTime
	}
}

func PlotAddTrailingSellPoint(buyId int64, time string, stopPrice float64) {
	if !canPlot() {
		return
	}

	if item, ok := plots[buyId]; ok {
		item.TrailingSell = append(item.TrailingSell, TrailingSellPlotData{
			Time:      time,
			StopPrice: stopPrice,
		})
	}
}

func PlotToJson(fileName string) {
	if !canPlot() {
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

func canPlot() bool {
	return 1 == BOTS_COUNT && 1 == GENERATION_COUNT
}
