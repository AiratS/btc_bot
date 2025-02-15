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
	Timezone   string       `json:"timezone"`
	ServerTime int64        `json:"serverTime"`
	Symbols    []MexcSymbol `json:"symbols"`
}

// Структура для хранения информации о конкретной торговой паре
type MexcSymbol struct {
	Symbol     string       `json:"symbol"`
	Status     string       `json:"status"`
	BaseAsset  string       `json:"baseAsset"`
	QuoteAsset string       `json:"quoteAsset"`
	Filters    []MexcFilter `json:"filters"`
}

// Структура для хранения фильтров (minQty, maxQty, stepSize и т.д.)
type MexcFilter struct {
	FilterType string `json:"filterType"`
	MinQty     string `json:"minQty,omitempty"`
	MaxQty     string `json:"maxQty,omitempty"`
	StepSize   string `json:"stepSize,omitempty"`
	MinPrice   string `json:"minPrice,omitempty"`
	MaxPrice   string `json:"maxPrice,omitempty"`
	TickSize   string `json:"tickSize,omitempty"`
}

func NewMexcFuturesOrderManager() MexcFuturesOrderManager {
	service := MexcFuturesOrderManager{
		isEnabled: USE_REAL_MONEY,
	}

	//if USE_REAL_MONEY {
	//	res, err := futuresClient.NewExchangeInfoService().
	//		Do(context.Background())
	//
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	balance := service.getBalance()
	//	if 10.0 >= balance {
	//		panic(fmt.Sprintf("Not enough money in balance: %f", balance))
	//	}
	//
	//	info := NewFuturesExchangeInfo(res)
	//	service.exchangeInfo = &info
	//
	//	service.adjustConfiguration()
	//}
	//
	return service
}

func (manager *MexcFuturesOrderManager) HasEnoughMoneyForBuy(usedMoney float64) bool {
	balance := manager.getBalance()

	return balance >= usedMoney
}

//
//func (manager *MexcFuturesOrderManager) CreateMarketBuyOrder(symbol string, price, usedMoney float64) (int64, float64, float64) {
//	if !manager.isEnabled {
//		return 0, 0.0, price
//	}
//
//	if info, hasLotSize := manager.exchangeInfo.GetInfoForSymbol(symbol); hasLotSize {
//		quantityLotSize := valueToLotSize(calcQuantity(price, manager.getOrderMoney(usedMoney)), info.LotSize.stepSize)
//		priceConverted := valueToPriceSize(price, info.PriceFilter.tickSize)
//
//		fmt.Println(fmt.Sprintf("CreateBuyOrder: %f, %f", priceConverted, quantityLotSize))
//
//		var errorMessage string
//		for i := 0; i < RETRIES_COUNT; i++ {
//			order, err := manager.futuresClient.
//				NewCreateOrderService().
//				Symbol(symbol).
//				Side(futures.SideTypeBuy).
//				Type(futures.OrderTypeMarket).
//				//PositionSide(futures.PositionSideTypeLong).
//				Quantity(floatToBinancePrice(quantityLotSize)).
//				Do(context.Background())
//
//			if err != nil {
//				errorMessage = err.Error()
//				Log(errorMessage)
//				time.Sleep(RETRY_DELAY * time.Millisecond)
//				continue
//			}
//
//			realBuyPrice, realQuantity := manager.getRealBuyPriceAndQuantity(priceConverted, quantityLotSize, order)
//
//			return order.OrderID, realQuantity, realBuyPrice
//		}
//
//		panic(errorMessage)
//	}
//
//	return 0, 0.0, 0.0
//}

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

//func getExchangeInfo() (*MexcExchangeInfoResponse, error) {
//	endpoint := "/api/v3/exchangeInfo"
//
//	// Создаем HTTP запрос
//	req, err := http.NewRequest("GET", MEXC_API_URL+endpoint, nil)
//	if err != nil {
//		return nil, err
//	}
//
//	// Выполняем запрос
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	// Читаем ответ
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return nil, err
//	}
//
//	st := string(body)
//	fmt.Println(st)
//
//	// Парсим JSON в структуру
//	var exchangeInfo MexcExchangeInfoResponse
//	err = json.Unmarshal(body, &exchangeInfo)
//	if err != nil {
//		return nil, err
//	}
//
//	return &exchangeInfo, nil
//}

func signRequest(queryString string) string {
	mac := hmac.New(sha256.New, []byte(MEXC_API_SECRET))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}
