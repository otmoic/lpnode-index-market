package binance

import (
	"context"
	"fmt"
	"log"
	"lp_market/logger"
	stdmarket "lp_market/std_market"
	"lp_market/types"
	"lp_market/utils"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/tidwall/sjson"
)

type CoinSwapMarketGroupConn struct {
	RequestId       int64
	symbolList      []types.ExchangeInfoSymbolApiResult
	webSocketCtx    context.Context
	webSocketCancel context.CancelFunc
	orderbookStore  stdmarket.StdCoinSwapOrderbook
}

func (csmgc *CoinSwapMarketGroupConn) Init() {
	logger.CSwapMarket.Debug("initialize orderbook storage for u-based contract")
	csmgc.orderbookStore = GetCoinSwapOrderbookStoreInstance()
	logger.CSwapMarket.Debug("stored object is", csmgc.orderbookStore)
}

func (csmgc *CoinSwapMarketGroupConn) SetSymbolList(symbolList []types.ExchangeInfoSymbolApiResult) {
	csmgc.symbolList = symbolList // set currency pair data
}

// Run start running
func (csmgc *CoinSwapMarketGroupConn) Run() {
	go func() {
		csmgc.Do()
		logger.CSwapMarket.Debug("[coinSwap]currency pair group startup program has been closed...")
	}()
}

func (csmgc *CoinSwapMarketGroupConn) GetLastMessageId() int64 {
	if csmgc.RequestId > 99999999999 {
		csmgc.RequestId = 0
	}
	csmgc.RequestId++
	return csmgc.RequestId
}

func (csmgc *CoinSwapMarketGroupConn) GetStreams() string {
	var binanceSymbols []string = []string{}
	for _, v := range csmgc.symbolList {
		binanceSymbols = append(binanceSymbols, strings.ToLower(v.Symbol))
	}
	return strings.Join(binanceSymbols, "/")
}
func (csmgc *CoinSwapMarketGroupConn) GetSymbolsString() string {
	var binanceSymbols []string = []string{}
	for _, v := range csmgc.symbolList {
		binanceSymbols = append(binanceSymbols, strings.ToLower(v.Symbol))
	}
	return strings.Join(binanceSymbols, "|")
}
func (csmgc *CoinSwapMarketGroupConn) GetSubParams() string {
	jsonStr := "{}"
	jsonVal, _ := sjson.Set(jsonStr, "method", "SUBSCRIBE")
	for index, v := range csmgc.symbolList {
		key := fmt.Sprintf("params.%d", index)
		symbolValue := fmt.Sprintf("%s@depth%d@%dms", strings.ToLower(v.Symbol), UsdtSwapLevels, UsdtSwapOrderbookUpdateInterval)
		jsonVal, _ = sjson.Set(jsonVal, key, symbolValue)
	}
	jsonVal, _ = sjson.Set(jsonVal, "id", csmgc.GetLastMessageId())
	logger.SpotMarket.Debug("json information needed for subscription", jsonVal)
	return jsonVal
}
func (csmgc *CoinSwapMarketGroupConn) Do() {
	logger.Httpd.Debug("start do............")
	// prepare control handle
	webSocketCtx, cancelWebSocket := context.WithCancel(context.Background())
	csmgc.webSocketCtx = webSocketCtx
	connectUrl := CoinSwapMarketWssBaseUrl + csmgc.GetStreams()
	logger.SpotMarket.Debug("url needed for linking is:", connectUrl)
	subParamsStr := csmgc.GetSubParams()
	csmgc.webSocketCancel = cancelWebSocket
	config := utils.WebSocketClientConnOption{
		Url:              connectUrl,
		NoDataDisconnect: true,
		NoDataTimeout:    70,
		SendPingInterval: 0, // do not actively send ping messages out
	}

	webSocket := utils.NewWebSocketClientConn(webSocketCtx, config)

	webSocket.OnMessage = func(message string) {
		// logger.USwapMarket.Debug("received message", message)
		orderItem := types.CoinSwapOrderBookInMessage{}
		err := sonic.Unmarshal([]byte(message), &orderItem)
		if err != nil {
			logger.USwapMarket.Debug("an error occurred", err)
			return
		}
		if orderItem.Data.StdStream == "" {
			logger.USwapMarket.Debug("might be a ping", message)
			return
		}
		if orderItem.Data.EventType != "depthUpdate" {
			logger.USwapMarket.Debug("not a valid update")
			return
		}
		orderBookItem := &types.OrderBookItem{}
		orderBookItem.IncomingTimestamp = time.Now().UnixNano() / 1e6
		orderBookItem.LastUpdateId = orderItem.Data.EventTimestamp
		orderBookItem.Asks = orderItem.Data.Asks
		orderBookItem.Bids = orderItem.Data.Bids
		orderBookItem.StreamName = orderItem.Stream
		orderBookItem.Timestamp = time.Now().UnixNano() / 1e6

		csmgc.orderbookStore.SetOrderbook(orderItem.Stream, orderBookItem)
	}
	webSocket.OnReconnect = func(connectCount int64, lastError string) {
		logger.USwapMarket.Debugf("link reestablishment underway%s", csmgc.GetSymbolsString())
	}
	webSocket.OnConnect = func(conn *utils.WebSocketClientConn) {
		log.Println("link establishment complete", conn.Url, "subscription information sent", subParamsStr)
		webSocket.SendTextMessage(subParamsStr)
	}
	webSocket.Initialize()

}

// Drop discard this class; do not process; execute when system reallocates resources
func (csmgc *CoinSwapMarketGroupConn) Drop() {
	csmgc.webSocketCancel()
}
