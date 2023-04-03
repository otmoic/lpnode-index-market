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
	orderbookStore  stdmarket.StdUsdtSwapOrderbook // u本位合约行情的存储
}

func (usmgc *UsdtSwapMarketGroupConn) Init() {
	logger.USwapMarket.Debug("初始化u本位合约的orderbook存储")
	usmgc.orderbookStore = GetUsdtSwapOrderbookStoreInstance()
	logger.USwapMarket.Debug("存储的对象是", usmgc.orderbookStore)
}
func (usmgc *UsdtSwapMarketGroupConn) SetSymbolList(symbolList []types.ExchangeInfoSymbolApiResult) {
	usmgc.symbolList = symbolList // 设置币对数据
}

// 开始运行
func (usmgc *UsdtSwapMarketGroupConn) Run() {
	go func() {
		usmgc.Do()
		logger.USwapMarket.Debug("[uswap]币对Group启动程序已关闭...")
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
	logger.SpotMarket.Debug("需要订阅的Json信息", jsonVal)
	return jsonVal
}
func (usmgc *UsdtSwapMarketGroupConn) Do() {
	logger.Httpd.Debug("开始Do............")
	// 准备控制句柄
	webSocketCtx, cancelWebSocket := context.WithCancel(context.Background())
	usmgc.webSocketCtx = webSocketCtx
	connectUrl := UsdtSwapMarketWssBaseUrl
	logger.SpotMarket.Debug("需要链接的Url是:", connectUrl)
	subParamsStr := usmgc.GetSubParams()
	usmgc.webSocketCancel = cancelWebSocket
	config := utils.WebSocketClientConnOption{
		Url:              connectUrl,
		NoDataDisconnect: true,
		NoDataTimeout:    70,
		SendPingInterval: 0, // 不要主动发送Ping消息出去
	}

	webSocket := utils.NewWebSocketClientConn(webSocketCtx, config)

	webSocket.OnMessage = func(message string) {
		orderItem := types.UsdtSwapOrderBookInMessage{}
		err := sonic.Unmarshal([]byte(message), &orderItem)
		if err != nil {
			logger.USwapMarket.Debug("发生了错误", err)
			return
		}
		if orderItem.Stream == "" {
			logger.USwapMarket.Debug("可能是Ping", message)
			return
		}
		if orderItem.EventType != "depthUpdate" {
			logger.USwapMarket.Debug("不是有效的Update")
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
		logger.USwapMarket.Debugf("链接重新建立%s", usmgc.GetSymbolsString())
	}
	webSocket.OnConnect = func(conn *utils.WebSocketClientConn) {
		log.Println("到币安的链接已经完成", conn.Url, "发送订阅信息", subParamsStr)
		webSocket.SendTextMessage(subParamsStr)
	}
	webSocket.Initialize()

}

// 丢弃这个类,不在处理 ，系统重新分配资源时执行
func (usmgc *UsdtSwapMarketGroupConn) Drop() {
	usmgc.webSocketCancel()
}
