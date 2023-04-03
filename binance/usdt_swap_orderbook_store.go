package binance

import (
	"lp_market/logger"
	stdmarket "lp_market/std_market"
	"lp_market/types"
	"strings"
	"sync"
)

type UsdtSwapOrderbookStore struct {
	// OrderBookList map[string]types.OrderBookItem
	OrderBookList sync.Map
}

var UsdtSwapOrderbookStoreInstance stdmarket.StdUsdtSwapOrderbook

func (usos *UsdtSwapOrderbookStore) Init() {

}

func (usos *UsdtSwapOrderbookStore) GetOrderbook() sync.Map {
	return usos.OrderBookList
}

func (usos *UsdtSwapOrderbookStore) SetOrderbook(streamName string, data *types.OrderBookItem) {
	streamName = strings.ToLower(streamName)
	symbolInfo := strings.Split(streamName, "@") // u in message streamName 是大写，要转小写
	symbolStdInfo, ok := UsdtSwapSymbolList_Global.Load(symbolInfo[0])
	if !ok {
		logger.Orderbook.Errorf("[UsdtSwap]没有找到标准的symbol%s", symbolInfo[0])
		return
	}

	symbolStdInfoStruct := symbolStdInfo.(types.ExchangeInfoSymbolApiResult)
	data.StdSymbol = symbolStdInfoStruct.StdSymbol
	data.Symbol = symbolStdInfoStruct.Symbol

	usos.OrderBookList.Store(symbolStdInfoStruct.StdSymbol, data)
	if !ShowUSwapOrderbookInfo { // 如果不打印消息则退出
		return
	}
	logger.Orderbook.Debug("【🟦🟦🟦🟦🟦🟦】", symbolStdInfoStruct.StdSymbol, data)
}

func GetUsdtSwapOrderbookStoreInstance() stdmarket.StdUsdtSwapOrderbook {
	usdtSwapOrderbookOnce.Do(func() {
		logger.SpotMarket.Debug("初始化U本位合约Orderbook存储🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶🌶")
		UsdtSwapOrderbookStoreInstance = &UsdtSwapOrderbookStore{
			OrderBookList: sync.Map{},
		}
		UsdtSwapOrderbookStoreInstance.Init()
	})
	return UsdtSwapOrderbookStoreInstance
}
