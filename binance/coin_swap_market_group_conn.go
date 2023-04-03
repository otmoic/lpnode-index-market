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
	orderbookStore  stdmarket.StdCoinSwapOrderbook // u本位合约行情的存储
}

func (csmgc *CoinSwapMarketGroupConn) Init() {
	logger.CSwapMarket.Debug("初始化u本位合约的orderbook存储")
	csmgc.orderbookStore = GetCoinSwapOrderbookStoreInstance()
	logger.CSwapMarket.Debug("存储的对象是", csmgc.orderbookStore)
}

func (csmgc *CoinSwapMarketGroupConn) SetSymbolList(symbolList []types.ExchangeInfoSymbolApiResult) {
	csmgc.symbolList = symbolList // 设置币对数据
}

// Run 开始运行
func (csmgc *CoinSwapMarketGroupConn) Run() {
	go func() {
		csmgc.Do()
		logger.CSwapMarket.Debug("[coinSwap]币对Group启动程序已关闭...")
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
	logger.SpotMarket.Debug("需要订阅的Json信息", jsonVal)
	return jsonVal
}
func (csmgc *CoinSwapMarketGroupConn) Do() {
	logger.Httpd.Debug("开始Do............")
	// 准备控制句柄
	webSocketCtx, cancelWebSocket := context.WithCancel(context.Background())
	csmgc.webSocketCtx = webSocketCtx
	connectUrl := CoinSwapMarketWssBaseUrl + csmgc.GetStreams()
	logger.SpotMarket.Debug("需要链接的Url是:", connectUrl)
	subParamsStr := csmgc.GetSubParams()
	csmgc.webSocketCancel = cancelWebSocket
	config := utils.WebSocketClientConnOption{
		Url:              connectUrl,
		NoDataDisconnect: true,
		NoDataTimeout:    70,
		SendPingInterval: 0, // 不要主动发送Ping消息出去
	}

	webSocket := utils.NewWebSocketClientConn(webSocketCtx, config)

	webSocket.OnMessage = func(message string) {
		// logger.USwapMarket.Debug("收到了Message", message)
		orderItem := types.CoinSwapOrderBookInMessage{}
		err := sonic.Unmarshal([]byte(message), &orderItem)
		if err != nil {
			logger.USwapMarket.Debug("发生了错误", err)
			return
		}
		if orderItem.Data.StdStream == "" {
			logger.USwapMarket.Debug("可能是Ping", message)
			return
		}
		if orderItem.Data.EventType != "depthUpdate" {
			logger.USwapMarket.Debug("不是有效的Update")
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
		logger.USwapMarket.Debugf("链接重新建立%s", csmgc.GetSymbolsString())
	}
	webSocket.OnConnect = func(conn *utils.WebSocketClientConn) {
		log.Println("到币安的链接已经完成", conn.Url, "发送订阅信息", subParamsStr)
		webSocket.SendTextMessage(subParamsStr)
	}
	webSocket.Initialize()

}

// Drop 丢弃这个类,不在处理 ，系统重新分配资源时执行
func (csmgc *CoinSwapMarketGroupConn) Drop() {
	csmgc.webSocketCancel()
}
