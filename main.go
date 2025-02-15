package main

//
//import (
//	"log"
//	"os"
//)
//
//func main() {
//	// Logger
//	logFileName := resolveLogFileName()
//	_, e := os.OpenFile(logFileName, os.O_RDONLY, 0666)
//	if !os.IsNotExist(e) {
//		e := os.Remove(logFileName)
//		if e != nil {
//			log.Fatal(e)
//		}
//	}
//
//	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
//	if err != nil {
//		log.Fatalf("error opening file: %v", err)
//	}
//	defer f.Close()
//	log.SetOutput(f)
//
//	if IS_REAL_ENABLED {
//		if ENABLE_FUTURES {
//			if USE_MEXC_STOCK {
//				RunMexcFuturesRealTime()
//			} else {
//				RunFuturesRealTime()
//			}
//		} else {
//			RunRealTime()
//		}
//		return
//	}
//
//	RunTest()
//}
//
//func resolveLogFileName() string {
//	if IS_REAL_ENABLED {
//		return "real_bot_log.txt"
//	}
//
//	return "bot_log.txt"
//}

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Структура для хранения информации о торговой паре
type SymbolInfo struct {
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
	Filters    []struct {
		FilterType  string `json:"filterType"`
		StepSize    string `json:"stepSize,omitempty"`    // lotSize
		MinQty      string `json:"minQty,omitempty"`      // Минимальное количество
		MaxQty      string `json:"maxQty,omitempty"`      // Максимальное количество
		TickSize    string `json:"tickSize,omitempty"`    // Шаг цены
		MinNotional string `json:"minNotional,omitempty"` // Минимальная сумма ордера
	} `json:"filters"`
}

// Структура для хранения ответа от API
type MexcExchangeInfo struct {
	Symbols []SymbolInfo `json:"symbols"`
}

// Функция для получения информации о торговых парах
func getMexcExchangeInfo() (*MexcExchangeInfo, error) {
	url := "https://api.mexc.com/api/v3/exchangeInfo"

	// Отправляем GET-запрос
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	fmt.Println(string(body))

	// Парсим JSON
	var exchangeInfo MexcExchangeInfo
	if err := json.Unmarshal(body, &exchangeInfo); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге JSON: %v", err)
	}

	return &exchangeInfo, nil
}

func main() {

	//getMexcExchangeInfo()

	createMarketOrder("BTCUSDT", "5.0")

}
