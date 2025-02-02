package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type SymbolIntervalPair struct {
	Symbol   string
	Interval string
}

// WsHandler handle raw websocket message
type WsHandler func(message []byte)

// ErrHandler handles errors
type ErrHandler func(err error)

type WsKlineHandler func(klineEvent *MexcKlineEvent)

type MexcKlineEvent struct {
	Symbol  string     `json:"symbol"`
	Data    MexcCandle `json:"data"`
	Channel string     `json:"channel"`
	Ts      int        `json:"ts"`
}

type MexcCandle struct {
	Symbol     string `json:"symbol"`
	Interval   string `json:"interval"`
	StartTime  int64  `json:"t"`
	EndTime    int64
	OpenPrice  float64 `json:"o"`
	ClosePrice float64 `json:"c"`
	HighPrice  float64 `json:"h"`
	LowPrice   float64 `json:"l"`
	Volume     float64 `json:"a"`
	Q          float32 `json:"q"`
	Ro         float32 `json:"r"`
	Rc         float32 `json:"rc"`
	Rh         float32 `json:"rh"`
	Rl         float32 `json:"rl"`
}

var (
	WebsocketEndpoint = "wss://contract.mexc.com/edge"
	// WebsocketTimeout is an interval for sending ping/pong messages if WebsocketKeepalive is enabled
	WebsocketTimeout = time.Second * 60
)

var wsServe = func(handler WsHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error, c *websocket.Conn) {
	Dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: false,
	}

	c, _, err = Dialer.Dial(WebsocketEndpoint, nil)
	if err != nil {
		return nil, nil, err, nil
	}
	c.SetReadLimit(655350)
	doneC = make(chan struct{})
	stopC = make(chan struct{})
	go func() {
		// This function will exit either on error from
		// websocket.Conn.ReadMessage or when the stopC channel is
		// closed by the client.
		defer close(doneC)

		keepAlive(c, WebsocketTimeout)
		// Wait for the stopC channel to be closed.  We do that in a
		// separate goroutine because ReadMessage is a blocking
		// operation.
		silent := false
		go func() {
			select {
			case <-stopC:
				silent = true
			case <-doneC:
			}
			c.Close()
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if !silent {
					errHandler(err)
				}
				return
			}
			handler(message)
		}
	}()
	return
}

func RunMexcFuturesRealTime() {
	pair := SymbolIntervalPair{
		Symbol:   ConvertBinanceToMexcSymbol(CANDLE_SYMBOL),
		Interval: ConvertBinanceToMexcInterval(CANDLE_INTERVAL),
	}

	for {
		fmt.Println("Connect to MEXC...")
		doneC, _, err := WsCombinedKlineServe(pair, wsKlineHandler, wsErrHandler)
		if err != nil {
			fmt.Println(err)
			continue
		}
		<-doneC

		fmt.Println("Disconnected, Reconnect in 3 seconds")
		time.Sleep(time.Second * 3)
	}
}

var prevMexcSecCandle Candle

func wsKlineHandler(klineEvent *MexcKlineEvent) {
	secCandle := MexcCandleToKlineCandleFutures(klineEvent.Data, &prevMexcSecCandle)
	fmt.Println(fmt.Sprintf("MEXC FUTURES: %s - Coin: %s, Price: %f", secCandle.CloseTime, secCandle.Symbol, secCandle.ClosePrice))

	if IS_REAL_ENABLED {
		realBot.runBuyIndicators(secCandle)
	}

	if convertedCandle, ok := candleConverter.Convert(secCandle); ok {
		Log("I am OK")
		realBot.DoStuff(convertedCandle)
	}
}

func wsErrHandler(err error) {
	fmt.Println("Error:", err)
}

func WsCombinedKlineServe(symbolIntervalPair SymbolIntervalPair, handler WsKlineHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(message []byte) {
		klineEvent := MexcKlineEvent{}

		err := json.Unmarshal(message, &klineEvent)
		if err != nil {
			fmt.Println(err)
			return
		}

		handler(&klineEvent)
	}

	doneC, stopC, err, wsConnect := wsServe(wsHandler, errHandler)
	if err != nil {
		return nil, nil, err
	}

	wsConnect.WriteJSON(map[string]interface{}{
		"method": "sub.kline",
		"param": map[string]interface{}{
			"symbol":   symbolIntervalPair.Symbol,
			"interval": symbolIntervalPair.Interval,
		},
	})

	return doneC, stopC, err
}

func keepAlive(c *websocket.Conn, timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	go func() {
		defer ticker.Stop()
		for {
			err := c.WriteJSON(map[string]interface{}{
				"method": "ping",
			})
			if err != nil {
				log.Println("write:", err)
				return
			}

			<-ticker.C
			//if time.Since(lastResponse) > timeout {
			//	c.Close()
			//	return
			//}
		}
	}()
}
