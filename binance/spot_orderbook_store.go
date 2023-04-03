package binance

import (
	"lp_market/logger"
	stdmarket "lp_market/std_market"
	"lp_market/types"
	"strings"
	"sync"
)

type SpotOrderbookStore struct {
	// OrderBookList map[string]types.OrderBookItem
	OrderBookList sync.Map // 放入已经解析好的orderbook
}

var SpotOrderbookStoreInstance stdmarket.StdSpotOrderbook

func (so *SpotOrderbookStore) Init() {

}

func (so *SpotOrderbookStore) GetOrderbook() sync.Map {
	return so.OrderBookList
}

// 设置现货Orderbook
func (so *SpotOrderbookStore) SetSpotOrderbook(streamName string, data *types.OrderBookItem) {
	symbolInfo := strings.Split(streamName, "@")
	symbolStdInfo, ok := SpotSymbolList_Global.Load(symbolInfo[0])
	if !ok {
		logger.Orderbook.Errorf("没有找到标准的symbol%s", symbolInfo[0])
		return
	}

	symbolStdInfoStruct := symbolStdInfo.(types.ExchangeInfoSymbolApiResult)
	data.StdSymbol = symbolStdInfoStruct.StdSymbol
	data.Symbol = symbolStdInfoStruct.Symbol

	so.OrderBookList.Store(symbolStdInfoStruct.StdSymbol, data)
	if !ShowSpotOrderbookInfo {
		return
	}
	logger.Orderbook.Debug("【🟩🟩🟩🟩🟩🟩】", symbolStdInfoStruct.StdSymbol, data)
}

func GetSpotOrderbookStoreInstance() stdmarket.StdSpotOrderbook {
	spotOrderbookOnce.Do(func() {
		logger.SpotMarket.Debug("初始化现货Orderbook🌶")
		SpotOrderbookStoreInstance = &SpotOrderbookStore{
			OrderBookList: sync.Map{},
		}
		SpotOrderbookStoreInstance.Init()
	})
	return SpotOrderbookStoreInstance
}
