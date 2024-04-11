package binance

import "sync"

var spotOrderbookOnce sync.Once
var spotMarketOnce sync.Once
var usdtSwapMarketOnce sync.Once
var usdtSwapOrderbookOnce sync.Once

var coinSwapMarketOnce sync.Once
var coinSwapOrderbookOnce sync.Once

var fundingRateOnce sync.Once
