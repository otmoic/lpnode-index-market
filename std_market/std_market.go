package stdmarket

import (
	"context"
	"lp_market/types"
	"sync"
)

type StdSpotMarket interface {
	GetGlobalUsedSymbolList() sync.Map
	RefreshMarket() error
	SetUsedSymbol(data []string)
	InitSymbolList()
	Init(ctx context.Context) error
}
type StdUsdtSwapMarket interface {
	GetGlobalUsedSymbolList() sync.Map
	GetStdSymbol(symbol string) string
	RefreshMarket() error
	SetUsedSymbol(data []string)
	InitSymbolList()
	Init(ctx context.Context) error
}
type StdCoinSwapMarket interface {
	GetGlobalUsedSymbolList() sync.Map
	GetStdSymbol(symbol string) string
	RefreshMarket() error
	SetUsedSymbol(data []string)
	InitSymbolList()
	Init(ctx context.Context) error
}

type StdSpotOrderbook interface {
	Init()
	GetOrderbook() sync.Map
	SetSpotOrderbook(streamName string, data *types.OrderBookItem)
}
type StdUsdtSwapOrderbook interface {
	Init()
	GetOrderbook() sync.Map
	SetOrderbook(streamName string, data *types.OrderBookItem)
}
type StdCoinSwapOrderbook interface {
	Init()
	GetOrderbook() sync.Map
	SetOrderbook(streamName string, data *types.OrderBookItem)
}

type StdFundingRate interface {
	Init(ctx context.Context) error
	GetUsdtFundingRate(stdSymbol string) *types.StdFundingRate
	GetCoinFundingRate(stdSymbol string) *types.StdFundingRate
	GetUsdtFundingRateList() sync.Map
}
