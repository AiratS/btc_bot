package main

import (
	"log"
	"os"
)

func main() {
	// Logger
	f, err := os.OpenFile("bot_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Main
	config := initConfig()
	bot := NewBot(&config)
	for _, candle := range *ImportDatasets() {
		bot.DoStuff(candle)
	}
}

func initConfig() Config {
	return Config{
		HighSellPercentage:    5,
		TrailingTopPercentage: 5,
	}
}
