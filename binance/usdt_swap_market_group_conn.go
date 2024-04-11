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

type UsdtSwapMarketGroupConn struct {
	RequestId       int64
	symbolList      []types.ExchangeInfoSymbolApiResult
	webSocketCtx    context.Context
	webSocketCancel context.CancelFunc
	orderbookStore  stdmarket.StdUsdtSwapOrderbook
}

func (usmgc *UsdtSwapMarketGroupConn) Init() {
	logger.USwapMarket.Debug("initialize orderbook storage for u-based contracts")
	usmgc.orderbookStore = GetUsdtSwapOrderbookStoreInstance()
	logger.USwapMarket.Debug("storage object is", usmgc.orderbookStore)
}
func (usmgc *UsdtSwapMarketGroupConn) SetSymbolList(symbolList []types.ExchangeInfoSymbolApiResult) {
	usmgc.symbolList = symbolList // set pairs data
}

func (usmgc *UsdtSwapMarketGroupConn) Run() {
	go func() {
		usmgc.Do()
		logger.USwapMarket.Debug("[uswap] currency pair group startup program has closed...")
	}()
}
func (usmgc *UsdtSwapMarketGroupConn) GetLastMessageId() int64 {
	if usmgc.RequestId > 99999999999 {
		usmgc.RequestId = 0
	}
	usmgc.RequestId++
	return usmgc.RequestId
}
func (usmgc *UsdtSwapMarketGroupConn) GetStream() string {
	var binanceSymbols []string = []string{}
	for _, v := range usmgc.symbolList {
		binanceSymbols = append(binanceSymbols, strings.ToLower(v.Symbol))
	}
	return strings.Join(binanceSymbols, "/")
}
func (usmgc *UsdtSwapMarketGroupConn) GetSymbolsString() string {
	var binanceSymbols []string = []string{}
	for _, v := range usmgc.symbolList {
		binanceSymbols = append(binanceSymbols, strings.ToLower(v.Symbol))
	}
	return strings.Join(binanceSymbols, "|")
}
func (usmgc *UsdtSwapMarketGroupConn) GetSubParams() string {
	jsonStr := "{}"
	jsonVal, _ := sjson.Set(jsonStr, "method", "SUBSCRIBE")
	for index, v := range usmgc.symbolList {
		key := fmt.Sprintf("params.%d", index)
		symbolValue := fmt.Sprintf("%s@depth%d@%dms", strings.ToLower(v.Symbol), UsdtSwapLevels, UsdtSwapOrderbookUpdateInterval)
		jsonVal, _ = sjson.Set(jsonVal, key, symbolValue)
	}
	jsonVal, _ = sjson.Set(jsonVal, "id", usmgc.GetLastMessageId())
	logger.SpotMarket.Debug("JSON information required for subscription", jsonVal)
	return jsonVal
}
func (usmgc *UsdtSwapMarketGroupConn) Do() {
	webSocketCtx, cancelWebSocket := context.WithCancel(context.Background())
	usmgc.webSocketCtx = webSocketCtx
	connectUrl := UsdtSwapMarketWssBaseUrl
	logger.SpotMarket.Debug("URL to connect is:", connectUrl)
	subParamsStr := usmgc.GetSubParams()
	usmgc.webSocketCancel = cancelWebSocket
	config := utils.WebSocketClientConnOption{
		Url:              connectUrl,
		NoDataDisconnect: true,
		NoDataTimeout:    70,
		SendPingInterval: 0,
	}

	webSocket := utils.NewWebSocketClientConn(webSocketCtx, config)

	webSocket.OnMessage = func(message string) {
		orderItem := types.UsdtSwapOrderBookInMessage{}
		err := sonic.Unmarshal([]byte(message), &orderItem)
		if err != nil {
			logger.USwapMarket.Debug("an error occurred", err)
			return
		}
		if orderItem.Stream == "" {
			logger.USwapMarket.Debug("possibly a Ping", message)
			return
		}
		if orderItem.EventType != "depthUpdate" {
			logger.USwapMarket.Debug("not a valid Update")
			return
		}
		orderBookItem := &types.OrderBookItem{}
		orderBookItem.IncomingTimestamp = time.Now().UnixNano() / 1e6
		orderBookItem.LastUpdateId = orderItem.EventTimestamp
		orderBookItem.Asks = orderItem.Asks
		orderBookItem.Bids = orderItem.Bids
		orderBookItem.StreamName = orderItem.Stream
		orderBookItem.Timestamp = time.Now().UnixNano() / 1e6

		usmgc.orderbookStore.SetOrderbook(orderItem.Stream, orderBookItem)
	}
	webSocket.OnReconnect = func(connectCount int64, lastError string) {
		logger.USwapMarket.Debugf("link reestablished %s", usmgc.GetSymbolsString())
	}
	webSocket.OnConnect = func(conn *utils.WebSocketClientConn) {
		log.Println("connection to Binance established", conn.Url, "sending subscription information", subParamsStr)
		webSocket.SendTextMessage(subParamsStr)
	}
	webSocket.Initialize()
}

// discard this class, cease handling; execute when system reallocates resources
func (usmgc *UsdtSwapMarketGroupConn) Drop() {
	usmgc.webSocketCancel()
}
