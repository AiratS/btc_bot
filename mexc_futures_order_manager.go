package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	MEXC_API_URL    = "https://api.mexc.com"
	MEXC_API_KEY    = "mx0vglsToGxHAceI9U"
	MEXC_API_SECRET = "6277f92023854ad4b3e6b4db2634e03a"
)

type MexcAccountInfo struct {
	Balances []MexcBalanceInfo `json:"balances"`
}

type MexcBalanceInfo struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type MexcFuturesOrderManager struct {
	isEnabled    bool
	exchangeInfo *FuturesExchangeInfo
}

type MexcExchangeInfoResponse struct {
	Timezone   string           `json:"timezone"`
	ServerTime int64            `json:"serverTime"`
	Symbols    []MexcSymbolInfo `json:"symbols"`
}

// Структура для хранения информации о торговой паре
type MexcSymbolInfo struct {
	Symbol                     string   `json:"symbol"`
	Status                     string   `json:"status"`
	BaseAsset                  string   `json:"baseAsset"`
	BaseAssetPrecision         int      `json:"baseAssetPrecision"`
	QuoteAsset                 string   `json:"quoteAsset"`
	QuotePrecision             int      `json:"quotePrecision"`
	QuoteAssetPrecision        int      `json:"quoteAssetPrecision"`
	BaseCommissionPrecision    float64  `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision   int      `json:"quoteCommissionPrecision"`
	OrderTypes                 []string `json:"orderTypes"`
	IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`
	QuoteAmountPrecision       string   `json:"quoteAmountPrecision"`
	BaseSizePrecision          string   `json:"baseSizePrecision"`
	Permissions                []string `json:"permissions"`
	Filters                    []string `json:"filters"` // Если filters могут содержать объекты, замените на соответствующую структуру
	MaxQuoteAmount             string   `json:"maxQuoteAmount"`
	MakerCommission            string   `json:"makerCommission"`
	TakerCommission            string   `json:"takerCommission"`
	QuoteAmountPrecisionMarket string   `json:"quoteAmountPrecisionMarket"`
	MaxQuoteAmountMarket       string   `json:"maxQuoteAmountMarket"`
	FullName                   string   `json:"fullName"`
	TradeSideType              int      `json:"tradeSideType"`
}

func NewMexcFuturesOrderManager() MexcFuturesOrderManager {
	service := MexcFuturesOrderManager{
		isEnabled: USE_REAL_MONEY,
	}

	if USE_REAL_MONEY {
		// Check balance
		balance := service.getBalance()
		fmt.Println(balance)
		if 10.0 >= balance {
			panic(fmt.Sprintf("Not enough money in balance: %f", balance))
		}

		// Init exchange info
		res, err := getExchangeInfo()
		if err != nil {
			panic(err)
		}
		info := NewMexcExchangeInfo(res)
		service.exchangeInfo = &info
	}

	return service
}

func getExchangeInfo() (*MexcExchangeInfoResponse, error) {
	endpoint := "/api/v3/exchangeInfo"

	// Создаем HTTP запрос
	req, err := http.NewRequest("GET", MEXC_API_URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Парсим JSON в структуру
	var exchangeInfo MexcExchangeInfoResponse
	err = json.Unmarshal(body, &exchangeInfo)
	if err != nil {
		return nil, err
	}

	return &exchangeInfo, nil
}

func (manager *MexcFuturesOrderManager) CanBuyForPrice(symbol string, price, usedMoney float64) bool {
	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney(usedMoney)), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		return info.LotSize.minQty <= quantityLotSize && quantityLotSize <= info.LotSize.maxQty &&
			info.PriceFilter.minPrice <= priceConverted && priceConverted <= info.PriceFilter.maxPrice
	}

	return false
}

func (manager *MexcFuturesOrderManager) HasEnoughMoneyForBuy(usedMoney float64) bool {
	balance := manager.getBalance()

	return balance >= usedMoney
}

func (manager *MexcFuturesOrderManager) CreateMarketBuyOrder(symbol string, price, usedMoney float64) (int64, float64, float64) {
	if !manager.isEnabled {
		return 0, 0.0, price
	}

	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney(usedMoney)), info.LotSize.stepSize)
		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)

		fmt.Println(fmt.Sprintf("CreateBuyOrder: %f, %f", priceConverted, quantityLotSize))

		var errorMessage string
		for i := 0; i < RETRIES_COUNT; i++ {
			order, err := createMarketOrder(symbol, Float64ToString(quantityLotSize))
			if err != nil {
				errorMessage = err.Error()
				Log(errorMessage)
				time.Sleep(RETRY_DELAY * time.Millisecond)
				continue
			}

			return order.OrderID, quantityLotSize, realBuyPrice

			// order, err := manager.futuresClient.
			// 	NewCreateOrderService().
			// 	Symbol(symbol).
			// 	Side(futures.SideTypeBuy).
			// 	Type(futures.OrderTypeMarket).
			// 	//PositionSide(futures.PositionSideTypeLong).
			// 	Quantity(floatToBinancePrice(quantityLotSize)).
			// 	Do(context.Background())

			// if err != nil {
			// 	errorMessage = err.Error()
			// 	Log(errorMessage)
			// 	time.Sleep(RETRY_DELAY * time.Millisecond)
			// 	continue
			// }

			// realBuyPrice, realQuantity := manager.getRealBuyPriceAndQuantity(priceConverted, quantityLotSize, order)

			// return order.OrderID, realQuantity, realBuyPrice
		}

		panic(errorMessage)
	}

	return 0, 0.0, 0.0
}

type OrderResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       int64  `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	TransactTime  int64  `json:"transactTime"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	Side          string `json:"side"`
}

func createMarketOrder(symbol, quoteOrderQty string) (*OrderResponse, error) {
	endpoint := "/api/v3/order"
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Параметры запроса
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", "BUY")
	params.Set("type", "MARKET")
	params.Set("quoteOrderQty", quoteOrderQty)
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))

	// Подписываем запрос
	signature := signRequest(params.Encode())

	// Добавляем подпись к параметрам
	params.Set("signature", signature)

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", MEXC_API_URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	req.Header.Set("X-MEXC-APIKEY", MEXC_API_KEY)
	req.URL.RawQuery = params.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	e := string(body)
	fmt.Println(e)

	// Парсим JSON в структуру
	var orderResponse OrderResponse
	err = json.Unmarshal(body, &orderResponse)
	if err != nil {
		return nil, err
	}

	return &orderResponse, nil
}

// Limit sell
func (manager *MexcFuturesOrderManager) CreateSellOrder(symbol string, stopPrice, quantity float64) int64 {
	if !manager.isEnabled {
		return 0
	}

	return 0
}

func createLimitSellOrder(symbol, quantity, price string) (*OrderResponse, error) {
	endpoint := "/api/v3/order"
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	timeInForce := "GTC"

	// Параметры запроса
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", "SELL") // Указываем "SELL" для продажи
	params.Set("type", "LIMIT")
	params.Set("quantity", quantity)       // Количество базового актива для продажи
	params.Set("price", price)             // Цена, по которой вы хотите продать
	params.Set("timeInForce", timeInForce) // Время действия ордера
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))

	// Подписываем запрос
	signature := signRequest(params.Encode())

	// Добавляем подпись к параметрам
	params.Set("signature", signature)

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", MEXC_API_URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	req.Header.Set("X-MEXC-APIKEY", MEXC_API_KEY)
	req.URL.RawQuery = params.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Парсим JSON в структуру
	var orderResponse OrderResponse

	fmt.Println(string(body))

	err = json.Unmarshal(body, &orderResponse)
	if err != nil {
		return nil, err
	}

	return &orderResponse, nil
}

func (manager *MexcFuturesOrderManager) GetOrderInfo(orderId int64) {
	// var errorMessage string

	// for i := 0; i < RETRIES_COUNT; i++ {
	// 	res, err := manager.futuresClient.NewGetOrderService().
	// 		OrderID(orderId).
	// 		Symbol(CANDLE_SYMBOL).
	// 		Do(context.Background())

	// 	if err != nil {
	// 		errorMessage = err.Error()
	// 		Log(errorMessage)
	// 		time.Sleep(RETRY_DELAY * time.Millisecond)
	// 		continue
	// 	}

	// 	fmt.Println(res.Status)

	// 	return res
	// }

	// panic(errorMessage)
}

// Структура для хранения информации о ордере
type OrderInfoResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       int64  `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	Status        string `json:"status"`
	TimeInForce   string `json:"timeInForce"`
	Type          string `json:"type"`
	Side          string `json:"side"`
	StopPrice     string `json:"stopPrice"`
	IcebergQty    string `json:"icebergQty"`
	Time          int64  `json:"time"`
	UpdateTime    int64  `json:"updateTime"`
	IsWorking     bool   `json:"isWorking"`
}

// Функция для получения информации о ордере
func getOrderInfo(symbol, orderID string) (*OrderInfoResponse, error) {
	endpoint := "/api/v3/order"
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Параметры запроса
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))

	// Подписываем запрос
	signature := signRequest(params.Encode())

	// Добавляем подпись к параметрам
	params.Set("signature", signature)

	// Создаем HTTP запрос
	req, err := http.NewRequest("GET", MEXC_API_URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	req.Header.Set("X-MEXC-APIKEY", MEXC_API_KEY)
	req.URL.RawQuery = params.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Парсим JSON в структуру
	var orderInfo OrderInfoResponse

	fmt.Println(string(body))

	err = json.Unmarshal(body, &orderInfo)
	if err != nil {
		return nil, err
	}

	return &orderInfo, nil
}

func (manager *MexcFuturesOrderManager) CancelOrder(symbol string, orderId int64) int64 {
	if !manager.isEnabled {
		return 0
	}

	return 0
}

// Структура для хранения информации о отмененном ордере
type CancelOrderResponse struct {
	Symbol            string `json:"symbol"`
	OrderID           int64  `json:"orderId"`
	OrigClientOrderID string `json:"origClientOrderId"`
	ClientOrderID     string `json:"clientOrderId"`
	Price             string `json:"price"`
	OrigQty           string `json:"origQty"`
	ExecutedQty       string `json:"executedQty"`
	Status            string `json:"status"`
	TimeInForce       string `json:"timeInForce"`
	Type              string `json:"type"`
	Side              string `json:"side"`
}

// Функция для отмены ордера
func cancelOrder(symbol, orderID string) (*CancelOrderResponse, error) {
	endpoint := "/api/v3/order"
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Параметры запроса
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))

	// Подписываем запрос
	signature := signRequest(params.Encode())

	// Добавляем подпись к параметрам
	params.Set("signature", signature)

	// Создаем HTTP запрос
	req, err := http.NewRequest("DELETE", MEXC_API_URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	req.Header.Set("X-MEXC-APIKEY", MEXC_API_KEY)
	req.URL.RawQuery = params.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Парсим JSON в структуру
	var cancelResponse CancelOrderResponse

	fmt.Println(string(body))

	err = json.Unmarshal(body, &cancelResponse)
	if err != nil {
		return nil, err
	}

	return &cancelResponse, nil
}

func (manager *MexcFuturesOrderManager) getBalance() float64 {
	var errorMessage string

	for i := 0; i < RETRIES_COUNT; i++ {
		res, err := getSpotBalance()
		if err != nil {
			errorMessage = err.Error()
			Log(errorMessage)
			time.Sleep(RETRY_DELAY * time.Millisecond)
			continue
		}

		var accountInfo MexcAccountInfo
		err = json.Unmarshal([]byte(res), &accountInfo)
		if err != nil {
			Log(err.Error())
			time.Sleep(RETRY_DELAY * time.Millisecond)
			continue
		}

		for _, balance := range accountInfo.Balances {
			if balance.Asset == "USDT" {
				return convertBinanceToFloat64(balance.Free)
			}
		}

		return 0
	}

	panic(errorMessage)
}

func getSpotBalance() (string, error) {
	endpoint := "/api/v3/account"
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Создаем строку для подписи
	queryString := fmt.Sprintf("timestamp=%d", timestamp)
	signature := signRequest(queryString)

	// Создаем HTTP запрос
	req, err := http.NewRequest("GET", MEXC_API_URL+endpoint+"?"+queryString+"&signature="+signature, nil)
	if err != nil {
		return "", err
	}

	// Устанавливаем заголовки
	req.Header.Set("X-MEXC-APIKEY", MEXC_API_KEY)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func signRequest(queryString string) string {
	mac := hmac.New(sha256.New, []byte(MEXC_API_SECRET))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

func (manager *MexcFuturesOrderManager) getOrderMoney(usedMoney float64) float64 {
	return usedMoney
}
