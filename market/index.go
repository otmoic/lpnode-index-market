package market

import (
	"context"
	"lp_market/binance"
	"lp_market/logger"
	stdmarket "lp_market/std_market"
)

type MarketCenter struct {
	exchangeName string
}

func (mc *MarketCenter) GetSpotIns() stdmarket.StdSpotMarket {
	return binance.GetSpotMarketInstance()
}
func (mc *MarketCenter) GetUsdtSwapIns() stdmarket.StdUsdtSwapMarket {
	return binance.GetUSwapMarketInstance()
}
func (mc *MarketCenter) GetCoinSwapIns() stdmarket.StdCoinSwapMarket {
	return binance.GetCoinSwapMarketInstance()
}
func (mc *MarketCenter) GetSpotOrderbookIns() stdmarket.StdSpotOrderbook {
	return binance.GetSpotOrderbookStoreInstance()
}
func (mc *MarketCenter) GetUsdtSwapOrderbookIns() stdmarket.StdUsdtSwapOrderbook {
	return binance.GetUsdtSwapOrderbookStoreInstance()
}
func (mc *MarketCenter) GetCoinSwapOrderbookIns() stdmarket.StdCoinSwapOrderbook {
	return binance.GetCoinSwapOrderbookStoreInstance()
}

func (mc *MarketCenter) InitSpot(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		err := binance.GetSpotMarketInstance().Init(ctx)
		if err != nil {
			logger.MainMessage.Errorf("initialization of Spot failed:%s", err)
		}
	}()
}
func (mc *MarketCenter) InitUsdtSwap(ctx context.Context, cancel context.CancelFunc) {

	go func() {
		err := binance.GetUSwapMarketInstance().Init(ctx)
		if err != nil {
			logger.MainMessage.Errorf("initialization of USwapMarket failed:%s", err)
		}
	}()
}
func (mc *MarketCenter) InitCoinSwap(ctx context.Context, cancel context.CancelFunc) {

	go func() {
		err := binance.GetCoinSwapMarketInstance().Init(ctx)
		if err != nil {
			logger.MainMessage.Errorf("initialization of CoinSwapMarket failed:%s", err)
		}
	}()
}
func (mc *MarketCenter) InitFundingRate(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		err := binance.GetFundingRateIns().Init(ctx)
		if err != nil {
			logger.MainMessage.Errorf("InitFundingRate failed:%s", err)
		}
	}()
}
func (mc *MarketCenter) GetFundingRateIns() stdmarket.StdFundingRate {
	return binance.GetFundingRateIns()
}

func init() {
	println("init market")
}
