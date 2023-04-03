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

	<-ctx.Done() // 监听启动器的退出 和cancel
	logger.SpotMarket.Debug(".......Spot Market Manager.协程退出")
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
		logger.FundingRate.Debugf("开始请求%s", url)
		_, body, errs := gorequest.New().Get(url).End()
		if len(errs) > 0 {
			return errors.New(fmt.Errorf("请求发生了错误:%s", errs[0]).Error())
		}
		var ret types.BinanceUsdtSwapFundingRateApiResult
		err := sonic.Unmarshal([]byte(body), &ret)
		if err != nil {
			return errors.New(fmt.Errorf("解码发生了错误:%s", err).Error())
		}
		fr.SaveUsdtFundingRate(ret)
		return nil
	}
	go func() error {
		for {
			err := sync()
			if err != nil {
				logger.FundingRate.Error("处理资金费率发生错误:%s", err.Error())
			}
			time.Sleep(time.Second * time.Duration(UsdtSwapFundingRateUpdateInterval))
		}
	}()
	return nil
}
func (fr *FundingRate) SyncCoin() error {
	sync := func() error {
		url := fmt.Sprintf("%s%s", CoinSwapMarketHttpsBaseUrl, CoinSwapFundingRatePath)
		logger.FundingRate.Debugf("开始请求%s", url)
		_, body, errs := gorequest.New().Get(url).End()
		if len(errs) > 0 {
			return errors.New(fmt.Errorf("请求发生了错误:%s", errs[0]).Error())
		}
		var ret []types.BinanceCoinSwapFundingRate
		err := sonic.Unmarshal([]byte(body), &ret)
		if err != nil {
			return errors.New(fmt.Errorf("解码发生了错误:%s", err).Error())
		}
		fr.SaveCoinFundingRate(ret)
		return nil
	}
	go func() error {
		for {
			err := sync()
			if err != nil {
				logger.FundingRate.Error("处理资金费率发生错误:%s", err.Error())
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
			// logger.FundingRate.Warnf("没有找到对应的标准设置:%s", item.Symbol)
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
			//logger.FundingRate.Warnf("没有找到对应的标准设置:%s", item.Symbol)
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
