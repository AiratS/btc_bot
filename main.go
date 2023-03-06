package main

import (
	"log"
	"os"
)

const IS_REAL_ENABLED = false
const USE_REAL_MONEY = false
const REAL_MONEY_DB_NAME = "amazing_real"

func main() {
	// Logger
	logFileName := resolveLogFileName()
	_, e := os.OpenFile(logFileName, os.O_RDONLY, 0666)
	if !os.IsNotExist(e) {
		e := os.Remove(logFileName)
		if e != nil {
			log.Fatal(e)
		}
	}

	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	if IS_REAL_ENABLED {
		RunRealTime()
		return
	}

	RunTest()
}

func resolveLogFileName() string {
	if IS_REAL_ENABLED {
		return "real_bot_log.txt"
	}

	return "bot_log.txt"
}
