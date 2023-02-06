package main

func main() {
	config := initConfig()
	bot := NewBot(&config)

	for _, candle := range *ImportDatasets() {
		bot.DoStuff(candle)
	}
}

func initConfig() Config {
	return Config{}
}
