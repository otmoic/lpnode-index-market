package binance

import (
	"context"
	"errors"
	"fmt"
	"lp_market/logger"
	stdmarket "lp_market/std_market"
	"lp_market/types"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/parnurzeal/gorequest"
)

type FundingRate struct {
	usdtSwapFundingRate sync.Map
	coinSwapFundingRate sync.Map
}

func (fr *FundingRate) Init(ctx context.Context) error {
	fr.SyncUsdt()
	fr.SyncCoin()

	<-ctx.Done()
	logger.SpotMarket.Debug(".......Spot Market Manager.goroutine exit")
	return nil
}
func (fr *FundingRate) GetUsdtFundingRateList() sync.Map {
	return fr.usdtSwapFundingRate
}
func (fr *FundingRate) GetUsdtFundingRate(stdSymbol string) *types.StdFundingRate {
	sourceItem, ok := fr.usdtSwapFundingRate.Load(stdSymbol)
	if !ok {
		return nil
	}
	item := sourceItem.(*types.StdFundingRate)
	return item
}
func (fr *FundingRate) GetCoinFundingRate(stdSymbol string) *types.StdFundingRate {
	sourceItem, ok := fr.coinSwapFundingRate.Load(stdSymbol)
	if !ok {
		return nil
	}
	item := sourceItem.(*types.StdFundingRate)
	return item
}

func (fr *FundingRate) SyncUsdt() error {
	sync := func() error {
		url := fmt.Sprintf("%s%s", UsdtSwapMarketHttpsBaseUrl, UsdtSwapFundingRate)
		logger.FundingRate.Debugf("initiating request%s", url)
		_, body, errs := gorequest.New().Get(url).End()
		if len(errs) > 0 {
			return errors.New(fmt.Errorf("request encountered an error :%s", errs[0]).Error())
		}
		var ret types.BinanceUsdtSwapFundingRateApiResult
		err := sonic.Unmarshal([]byte(body), &ret)
		if err != nil {
			return errors.New(fmt.Errorf("decoding encountered an error:%s", err).Error())
		}
		fr.SaveUsdtFundingRate(ret)
		return nil
	}
	go func() error {
		for {
			err := sync()
			if err != nil {
				logger.FundingRate.Error("error processing funding rates:%s", err.Error())
			}
			time.Sleep(time.Second * time.Duration(UsdtSwapFundingRateUpdateInterval))
		}
	}()
	return nil
}
func (fr *FundingRate) SyncCoin() error {
	sync := func() error {
		url := fmt.Sprintf("%s%s", CoinSwapMarketHttpsBaseUrl, CoinSwapFundingRatePath)
		logger.FundingRate.Debugf("starting request to %s", url)
		_, body, errs := gorequest.New().Get(url).End()
		if len(errs) > 0 {
			return errors.New(fmt.Errorf("an error occurred during request: %s", errs[0]).Error())
		}
		var ret []types.BinanceCoinSwapFundingRate
		err := sonic.Unmarshal([]byte(body), &ret)
		if err != nil {
			return errors.New(fmt.Errorf("an error occurred during decoding: %s", err).Error())
		}
		fr.SaveCoinFundingRate(ret)
		return nil
	}

	go func() error {
		for {
			if err := sync(); err != nil {
				logger.FundingRate.Errorf("error processing funding rates: %s", err.Error())
			}
			time.Sleep(time.Second * time.Duration(CoinSwapFundingRateUpdateInterval))
		}
	}()
	return nil
}
func (fr *FundingRate) SaveUsdtFundingRate(ret []types.BinanceUsdtSwapFundingRate) {
	for _, item := range ret {
		stdSymbol := GetUSwapMarketInstance().GetStdSymbol(item.Symbol)
		if stdSymbol == "" {
			// logger.FundingRate.Warnf("unable to find corresponding standard settings :%s", item.Symbol)
			continue
		}
		fr.usdtSwapFundingRate.Store(stdSymbol, &types.StdFundingRate{
			FundingRate: item.LastFundingRate,
			StdSymbol:   stdSymbol,
			Symbol:      item.Symbol,
		})
	}
}
func (fr *FundingRate) SaveCoinFundingRate(ret []types.BinanceCoinSwapFundingRate) {
	for _, item := range ret {
		stdSymbol := GetCoinSwapMarketInstance().GetStdSymbol(item.Pair)
		if stdSymbol == "" {
			//logger.FundingRate.Warnf("unable to find corresponding standard settings :%s", item.Symbol)
			continue
		}
		fr.coinSwapFundingRate.Store(stdSymbol, &types.StdFundingRate{
			FundingRate: item.LastFundingRate,
			StdSymbol:   stdSymbol,
			Symbol:      item.Symbol,
		})
	}
}

var fundingRateIns stdmarket.StdFundingRate

func GetFundingRateIns() stdmarket.StdFundingRate {
	fundingRateOnce.Do(func() {
		fundingRateIns = &FundingRate{}
	})
	return fundingRateIns
}
