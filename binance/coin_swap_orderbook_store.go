package binance

import (
	"lp_market/logger"
	stdmarket "lp_market/std_market"
	"lp_market/types"
	"strings"
	"sync"
)

type CoinSwapOrderbookStore struct {
	OrderBookList sync.Map
}

var CoinSwapOrderbookStoreInstance stdmarket.StdCoinSwapOrderbook

func (csos *CoinSwapOrderbookStore) Init() {

}

func (csos *CoinSwapOrderbookStore) GetOrderbook() sync.Map {
	return csos.OrderBookList
}

func (csos *CoinSwapOrderbookStore) SetOrderbook(streamName string, data *types.OrderBookItem) {
	streamName = strings.ToLower(streamName)
	symbolInfo := strings.Split(streamName, "@") // u in message streamName 是大写，要转小写
	symbolStdInfo, ok := CoinSwapSymbolList_Global.Load(symbolInfo[0])
	if !ok {
		logger.Orderbook.Errorf("[CoinSwap]没有找到标准的symbol:【%s】", symbolInfo[0])
		return
	}

	symbolStdInfoStruct := symbolStdInfo.(types.ExchangeInfoSymbolApiResult)
	data.StdSymbol = symbolStdInfoStruct.StdSymbol
	data.Symbol = symbolStdInfoStruct.Symbol
	csos.OrderBookList.Store(symbolStdInfoStruct.StdSymbol, data)
	if !ShowCSwapOrderbookInfo {
		return
	}
	logger.Orderbook.Debug("【🟨🟨🟨🟨🟨🟨】", symbolStdInfoStruct.StdSymbol, data)

}

func GetCoinSwapOrderbookStoreInstance() stdmarket.StdCoinSwapOrderbook {
	coinSwapOrderbookOnce.Do(func() {
		logger.SpotMarket.Debug("初始化币本位合约Orderbook存储🌶")
		CoinSwapOrderbookStoreInstance = &CoinSwapOrderbookStore{
			OrderBookList: sync.Map{},
		}
		CoinSwapOrderbookStoreInstance.Init()
	})
	return CoinSwapOrderbookStoreInstance
}
