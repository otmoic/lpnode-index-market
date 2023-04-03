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
	logger.SpotMarket.Debug("初始化OrderBook的存储..")
	smgc.SpotOrderbookInstance = GetSpotOrderbookStoreInstance()
	logger.SpotMarket.Debug("存储的对象是", smgc.SpotOrderbookInstance)
}
func (smgc *SpotMarketGroupConn) SetSymbolList(symbolList []types.ExchangeInfoSymbolApiResult) {
	smgc.symbolList = symbolList // 设置币对数据
}

// 开始运行
func (smgc *SpotMarketGroupConn) Run() {
	go func() {
		smgc.Do()
		logger.SpotMarket.Debug("开始处理", smgc.GetSymbolsString())
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
	logger.SpotMarket.Debug("需要订阅的Json信息", jsonVal)
	return jsonVal
}
func (smgc *SpotMarketGroupConn) Do() {
	logger.Httpd.Debug("开始Do............")
	// 准备控制句柄
	webSocketCtx, cancelWebSocket := context.WithCancel(context.Background())
	smgc.webSocketCtx = webSocketCtx
	connectUrl := SpotMarketWssBaseUrl + smgc.GetStream()
	logger.SpotMarket.Debug("需要链接的Url是:", connectUrl)
	logger.SpotMarket.Debug("当前处理的币对列表是", smgc.GetSymbolsString())
	subParamsStr := smgc.GetSubParams()
	smgc.webSocketCancel = cancelWebSocket
	config := utils.WebSocketClientConnOption{
		Url:              connectUrl,
		NoDataDisconnect: true,
		NoDataTimeout:    10,
		SendPingInterval: 0, // 不要主动发送Ping消息出去
	}

	webSocket := utils.NewWebSocketClientConn(webSocketCtx, config)

	webSocket.OnMessage = func(message string) {
		log.Println(message)
		orderItem := types.SpotOrderBookInMessage{}
		err := sonic.Unmarshal([]byte(message), &orderItem)
		if err != nil {
			logger.SpotMarket.Debug("发生了错误", err)
			return
		}
		if orderItem.Stream == "" {
			logger.SpotMarket.Debug("可能是Ping", message)
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
		logger.SpotMarket.Debugf("链接重新建立%s", smgc.GetSymbolsString())
	}
	webSocket.OnConnect = func(conn *utils.WebSocketClientConn) {
		log.Println("到币安的链接已经完成", conn.Url, "发送订阅信息", subParamsStr)
		webSocket.SendTextMessage(subParamsStr)
	}
	webSocket.Initialize()

}

// 丢弃这个类,不在处理 ，系统重新分配资源时执行
func (smgc *SpotMarketGroupConn) Drop() {
	smgc.webSocketCancel()
}
