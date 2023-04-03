package binance

import "sync"

// 现货部分
var spotOrderbookOnce sync.Once // 用于处理SpotOrderbook的单例 （Store存储）
var spotMarketOnce sync.Once    // 用于处理spotMarket

// u本位合约部分
var usdtSwapMarketOnce sync.Once // 用于u本位合约market 的初始化
var usdtSwapOrderbookOnce sync.Once

// 币本位合约部分

var coinSwapMarketOnce sync.Once
var coinSwapOrderbookOnce sync.Once

// 资金费率
var fundingRateOnce sync.Once
