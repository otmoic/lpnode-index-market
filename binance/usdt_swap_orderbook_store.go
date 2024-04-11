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
	symbolInfo := strings.Split(streamName, "@")
	symbolStdInfo, ok := UsdtSwapSymbolList_Global.Load(symbolInfo[0])
	if !ok {
		logger.Orderbook.Errorf("[UsdtSwap]unable to find standard symbol%s", symbolInfo[0])
		return
	}

	symbolStdInfoStruct := symbolStdInfo.(types.ExchangeInfoSymbolApiResult)
	data.StdSymbol = symbolStdInfoStruct.StdSymbol
	data.Symbol = symbolStdInfoStruct.Symbol

	usos.OrderBookList.Store(symbolStdInfoStruct.StdSymbol, data)
	if !ShowUSwapOrderbookInfo {
		return
	}
	logger.Orderbook.Debug("【🟦】", symbolStdInfoStruct.StdSymbol, data)
}

func GetUsdtSwapOrderbookStoreInstance() stdmarket.StdUsdtSwapOrderbook {
	usdtSwapOrderbookOnce.Do(func() {
		logger.SpotMarket.Debug("initialize U-based contract orderbook storage 🌶️")
		UsdtSwapOrderbookStoreInstance = &UsdtSwapOrderbookStore{
			OrderBookList: sync.Map{},
		}
		UsdtSwapOrderbookStoreInstance.Init()
	})
	return UsdtSwapOrderbookStoreInstance
}
