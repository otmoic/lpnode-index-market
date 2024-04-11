package binance

import (
	"context"
	"errors"
	"fmt"
	"lp_market/logger"
	stdmarket "lp_market/std_market"
	"lp_market/types"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/parnurzeal/gorequest"
)

var UsdtSwapSymbolList_Global sync.Map     // globally shared u-based contract currency pair information
var UsdtSwapUsedSymbolList_Global sync.Map // globally shared current subscribed currency pair information
type USwapMarket struct {
	symbolList                    sync.Map
	symbolSetList                 map[string]bool
	usedSymbolList                sync.Map
	UsdtSwapGroupManagerListStore sync.Map // stores u-perpetual group subscription instances
}

var uSwapMarketInstance stdmarket.StdUsdtSwapMarket

func (usm *USwapMarket) GetGlobalUsedSymbolList() sync.Map {
	return UsdtSwapUsedSymbolList_Global
}
func (usm *USwapMarket) GetStdSymbol(symbol string) string {
	v, ok := UsdtSwapSymbolList_Global.Load(strings.ToLower(symbol))
	if !ok {
		return ""
	}
	symbolInfo := v.(types.ExchangeInfoSymbolApiResult)
	return symbolInfo.StdSymbol
}

// set only if currency pair is in whitelist, preparing subscribable pairs
func (usm *USwapMarket) PreMarketSymbol() {
	usm.symbolList.Range(func(key, value interface{}) bool {
		stdSymbolStruct := value.(types.ExchangeInfoSymbolApiResult)
		_, ok := usm.symbolSetList[stdSymbolStruct.StdSymbol]
		if ok {
			// logger.USwapMarket.Debugf("this can handle%s", stdSymbolStruct.StdSymbol)
			usm.usedSymbolList.Store(stdSymbolStruct.Symbol, stdSymbolStruct)
		}
		return true
	})
	UsdtSwapUsedSymbolList_Global = usm.usedSymbolList
}

// deletes current subscriptions and restarts
func (usm *USwapMarket) RefreshMarket() error {
	usm.UsdtSwapGroupManagerListStore.Range(func(key, value any) bool {
		logger.USwapMarket.Debug("detected a [GroupManager]", key.(int64), value.(*UsdtSwapMarketGroupConn))
		smgc := value.(*UsdtSwapMarketGroupConn)
		logger.USwapMarket.Debug("removing reference.......", smgc.GetSymbolsString())
		smgc.Drop()
		usm.UsdtSwapGroupManagerListStore.Delete(key)
		return true
	})
	go func() {
		for x := 0; x < 5; x++ {
			time.Sleep(time.Second * 2)
			runtime.GC()
		}
		usm.ProcessMarket() // resubscribing
	}()
	return nil
}
func (usm *USwapMarket) ProcessMarket() error {

	usm.PreMarketSymbol()
	var sIndex int64 = 0
	var processIndex int64 = 0
	list := []types.ExchangeInfoSymbolApiResult{}
	clearList := func() {
		list = []types.ExchangeInfoSymbolApiResult{}
	}
	usm.usedSymbolList.Range(func(key interface{}, value interface{}) bool {
		sIndex++
		list = append(list, value.(types.ExchangeInfoSymbolApiResult))
		if sIndex%5 == 0 {
			processIndex++
			logger.USwapMarket.Debug("begin processing Type[Uswap]", processIndex)
			usm.RunList(list, processIndex)
			clearList()
		}
		return true
	})
	if len(list) > 0 {
		processIndex++
		usm.RunList(list, processIndex)
		clearList()
	}

	return nil
}
func (usm *USwapMarket) RunList(list []types.ExchangeInfoSymbolApiResult, processIndex int64) {
	smgc := &UsdtSwapMarketGroupConn{}
	smgc.Init()
	smgc.SetSymbolList(list)
	smgc.Run()
	runtime.SetFinalizer(smgc, func(smgc *UsdtSwapMarketGroupConn) {
		logger.SpotMarket.Debugf("💢💢💢💢💢 Gc SymbolList Is:%s", smgc.GetSymbolsString())
	})
	usm.UsdtSwapGroupManagerListStore.Store(processIndex, smgc)
}

func (usm *USwapMarket) GetSymbols() error {
	url := fmt.Sprintf("%s%s", UsdtSwapMarketHttpsBaseUrl, UsdtSwapExchangeInfoPath)
	logger.USwapMarket.Debugf("path for trading standard information is :%s", url)
	_, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 {
		return errors.New(fmt.Errorf("request encountered an error:%s", errs[0]).Error())
	}
	var ret types.ExchangeInfoApiResult
	err := sonic.Unmarshal([]byte(body), &ret) // decode json content
	if err != nil {
		return errors.New(fmt.Errorf("decoding encountered an error:%s", err).Error())
	}
	symbolList := FormatterUsdtSwapExchangeInfo(ret)
	for _, v := range symbolList {
		usm.symbolList.Store(strings.ToLower(v.Symbol), v) // storing all currency pairs
	}
	UsdtSwapSymbolList_Global = usm.symbolList
	logger.USwapMarket.Debugf("u-based contract currency pair information retrieval completed with a total of [%d] pairs", len(ret.Symbols))
	return nil

}
func (usm *USwapMarket) SetUsedSymbol(symbolList []string) {
	for _, v := range symbolList {
		usm.symbolSetList[v] = true
	}
}
func (usm *USwapMarket) Init(ctx context.Context) error {
	logger.USwapMarket.Debug("begin initializing Uswap symbols..!")
	err := usm.GetSymbols()
	if err != nil {
		return err
	}
	err = usm.ProcessMarket() // begin processing market data
	if err != nil {
		logger.USwapMarket.Errorf("error processing market data%s", err)
	}

	<-ctx.Done()
	logger.USwapMarket.Debug(".......UsdtSwapMarket  Manager.goroutine exiting")

	return nil
}
func (usm *USwapMarket) InitSymbolList() {
	usm.symbolSetList = make(map[string]bool)
}

// return singleton instance of U-based Swap market
func GetUSwapMarketInstance() stdmarket.StdUsdtSwapMarket {
	usdtSwapMarketOnce.Do(func() {
		uSwapMarketInstance = &USwapMarket{}
		uSwapMarketInstance.InitSymbolList()
	})
	return uSwapMarketInstance

}
