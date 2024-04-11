package binance

import (
	"context"
	"fmt"
	"log"
	"lp_market/logger"
	stdmarket "lp_market/std_market"
	"lp_market/types"
	"strings"
	"time"

	"lp_market/utils"

	"github.com/bytedance/sonic"
	"github.com/tidwall/sjson"
)

type SpotMarketGroupConn struct {
	RequestId             int64
	symbolList            []types.ExchangeInfoSymbolApiResult
	webSocketCtx          context.Context
	webSocketCancel       context.CancelFunc
	SpotOrderbookInstance stdmarket.StdSpotOrderbook
}

func (smgc *SpotMarketGroupConn) Init() {
	logger.SpotMarket.Debug("initializing OrderBook storage..")
	smgc.SpotOrderbookInstance = GetSpotOrderbookStoreInstance()
	logger.SpotMarket.Debug("storage object is", smgc.SpotOrderbookInstance)
}
func (smgc *SpotMarketGroupConn) SetSymbolList(symbolList []types.ExchangeInfoSymbolApiResult) {
	smgc.symbolList = symbolList // setting currency pair data
}

// commencing execution
func (smgc *SpotMarketGroupConn) Run() {
	go func() {
		smgc.Do()
		logger.SpotMarket.Debug("beginning processing", smgc.GetSymbolsString())
	}()
}
func (smgc *SpotMarketGroupConn) GetLastMessageId() int64 {
	if smgc.RequestId > 99999999999 {
		smgc.RequestId = 0
	}
	smgc.RequestId++
	return smgc.RequestId
}
func (smgc *SpotMarketGroupConn) GetStream() string {
	var binanceSymbols []string = []string{}
	for _, v := range smgc.symbolList {
		binanceSymbols = append(binanceSymbols, strings.ToLower(v.Symbol))
	}
	return strings.Join(binanceSymbols, "/")
}
func (smgc *SpotMarketGroupConn) GetSymbolsString() string {
	var binanceSymbols []string = []string{}
	for _, v := range smgc.symbolList {
		binanceSymbols = append(binanceSymbols, strings.ToLower(v.Symbol))
	}
	return strings.Join(binanceSymbols, "|")
}
func (smgc *SpotMarketGroupConn) GetSubParams() string {
	jsonStr := "{}"
	jsonVal, _ := sjson.Set(jsonStr, "method", "SUBSCRIBE")
	for index, v := range smgc.symbolList {
		key := fmt.Sprintf("params.%d", index)
		symbolValue := fmt.Sprintf("%s@depth%d@%dms", strings.ToLower(v.Symbol), SpotLevels, SpotOrderbookUpdateInterval)
		jsonVal, _ = sjson.Set(jsonVal, key, symbolValue)
	}
	jsonVal, _ = sjson.Set(jsonVal, "id", smgc.GetLastMessageId())
	logger.SpotMarket.Debug("JSON information required for subscription", jsonVal)
	return jsonVal
}
func (smgc *SpotMarketGroupConn) Do() {
	logger.Httpd.Debug("stat do ............")
	// preparing control handle
	webSocketCtx, cancelWebSocket := context.WithCancel(context.Background())
	smgc.webSocketCtx = webSocketCtx
	connectUrl := SpotMarketWssBaseUrl + smgc.GetStream()
	logger.SpotMarket.Debug("URL to connect is:", connectUrl)
	logger.SpotMarket.Debug("current list of currency pairs being processed is", smgc.GetSymbolsString())
	subParamsStr := smgc.GetSubParams()
	smgc.webSocketCancel = cancelWebSocket
	config := utils.WebSocketClientConnOption{
		Url:              connectUrl,
		NoDataDisconnect: true,
		NoDataTimeout:    10,
		SendPingInterval: 0, // do not actively send Ping messages
	}

	webSocket := utils.NewWebSocketClientConn(webSocketCtx, config)

	webSocket.OnMessage = func(message string) {
		log.Println(message)
		orderItem := types.SpotOrderBookInMessage{}
		err := sonic.Unmarshal([]byte(message), &orderItem)
		if err != nil {
			logger.SpotMarket.Debug("an error occurred", err)
			return
		}
		if orderItem.Stream == "" {
			logger.SpotMarket.Debug("possibly due to Ping", message)
			return
		}
		orderBookItem := &types.OrderBookItem{}
		orderBookItem.IncomingTimestamp = time.Now().UnixNano() / 1e6
		orderBookItem.LastUpdateId = orderItem.Data.LastUpdateId
		orderBookItem.Asks = orderItem.Data.Asks
		orderBookItem.Bids = orderItem.Data.Bids
		orderBookItem.StreamName = orderItem.Stream
		orderBookItem.Timestamp = time.Now().UnixNano() / 1e6

		smgc.SpotOrderbookInstance.SetSpotOrderbook(orderItem.Stream, orderBookItem)
	}
	webSocket.OnReconnect = func(connectCount int64, lastError string) {
		logger.SpotMarket.Debugf("re-establishing link %s", smgc.GetSymbolsString())
	}
	webSocket.OnConnect = func(conn *utils.WebSocketClientConn) {
		log.Println("link has been completed", conn.Url, "send sub", subParamsStr)
		webSocket.SendTextMessage(subParamsStr)
	}
	webSocket.Initialize()

}

// discard this class, cease handling; execute when system reallocates resources
func (smgc *SpotMarketGroupConn) Drop() {
	smgc.webSocketCancel()
}
