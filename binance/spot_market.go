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

var SpotSymbolList_Global sync.Map
var SpotUsedSymbolList_Global sync.Map

type SpotMarket struct {
	symbolList                sync.Map        // saved list of symbols already retrieved
	symbolSetList             map[string]bool // configured whitelist
	usedSymbolList            sync.Map        // list of currency pairs to subscribe to, e.g. {StdSymbol: GRS/BTC}
	SpotGroupManagerListStore sync.Map
}

var spotMarketInstance stdmarket.StdSpotMarket

// return singleton instance of spot market
func GetSpotMarketInstance() stdmarket.StdSpotMarket {
	spotMarketOnce.Do(func() {
		spotMarketInstance = &SpotMarket{}
		spotMarketInstance.InitSymbolList()
	})
	return spotMarketInstance
}

func (spotMarket *SpotMarket) GetGlobalUsedSymbolList() sync.Map {
	return SpotUsedSymbolList_Global
}

func (spotMarket *SpotMarket) SetUsedSymbol(symbolList []string) {
	for _, v := range symbolList {
		spotMarket.symbolSetList[v] = true
	}
}
func (spotMarket *SpotMarket) InitSymbolList() {
	spotMarket.symbolSetList = make(map[string]bool)
}
func (spotMarket *SpotMarket) PreMarketSymbol() {
	spotMarket.symbolList.Range(func(key, value interface{}) bool {
		stdSymbolStruct := value.(types.ExchangeInfoSymbolApiResult)
		_, ok := spotMarket.symbolSetList[stdSymbolStruct.StdSymbol]
		if ok {
			spotMarket.usedSymbolList.Store(stdSymbolStruct.Symbol, stdSymbolStruct)
			logger.SpotMarket.Debugf("this can handle %s", stdSymbolStruct.StdSymbol)
		}
		return true
	})
	SpotUsedSymbolList_Global = spotMarket.usedSymbolList

}
func (spotMarket *SpotMarket) RefreshMarket() error {
	spotMarket.SpotGroupManagerListStore.Range(func(key, value any) bool {
		logger.SpotMarket.Debug("detected a [GroupManager]", key.(int64), value.(*SpotMarketGroupConn))
		smgc := value.(*SpotMarketGroupConn)
		logger.SpotMarket.Debug("removing reference.......", smgc.GetSymbolsString())
		smgc.Drop()
		spotMarket.SpotGroupManagerListStore.Delete(key)
		return true
	})
	go func() {
		for x := 0; x < 5; x++ {
			time.Sleep(time.Second * 2)
			runtime.GC()
		}
		spotMarket.ProcessMarket() // resubscribing
	}()
	return nil
}

// begin processing market data
func (spotMarket *SpotMarket) ProcessMarket() error {
	spotMarket.PreMarketSymbol()
	var sIndex int64 = 0
	var processIndex int64 = 0
	list := []types.ExchangeInfoSymbolApiResult{}
	clearList := func() {
		list = []types.ExchangeInfoSymbolApiResult{}
	}
	spotMarket.usedSymbolList.Range(func(key interface{}, value interface{}) bool {
		sIndex++
		list = append(list, value.(types.ExchangeInfoSymbolApiResult))
		if sIndex%5 == 0 {
			processIndex++
			logger.SpotMarket.Debug("begin processing", processIndex)
			spotMarket.RunList(list, processIndex)
			clearList()
		}
		return true
	})
	if len(list) != 0 { //remaining unprocessed items
		processIndex++
		logger.SpotMarket.Debug("begin processing", processIndex)
		spotMarket.RunList(list, processIndex)
		clearList()
	}

	return nil
}
func (spotMarket *SpotMarket) RunList(list []types.ExchangeInfoSymbolApiResult, processIndex int64) {
	smgc := &SpotMarketGroupConn{}
	smgc.Init()
	smgc.SetSymbolList(list)
	smgc.Run()
	runtime.SetFinalizer(smgc, func(smgc *SpotMarketGroupConn) {
		logger.SpotMarket.Debugf("💢💢💢💢💢 Gc SymbolList Is:%s", smgc.GetSymbolsString())
	})
	spotMarket.SpotGroupManagerListStore.Store(processIndex, smgc)
}

// obtaining all spot currency pairs and setting to list
func (spotMarket *SpotMarket) GetSpotSymbols() error {
	url := fmt.Sprintf("%s%s", SpotMarketHttpsBaseUrl, ExchangeInfoPath)
	logger.SpotMarket.Debugf("path for trading standard information is:%s", url)
	_, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 {
		return errors.New(fmt.Errorf("request err:%s", errs[0]).Error())
	}
	var ret types.ExchangeInfoApiResult
	err := sonic.Unmarshal([]byte(body), &ret)
	if err != nil {
		return errors.New(fmt.Errorf("decoding encountered an error:%s", err).Error())
	}
	symbolList := FormatterSpotExchangeInfo(ret)
	for _, v := range symbolList {
		// logger.SpotMarket.Debugf("start store%s", v.StdSymbol)
		spotMarket.symbolList.Store(strings.ToLower(v.Symbol), v)
	}
	SpotSymbolList_Global = spotMarket.symbolList
	logger.SpotMarket.Debugf("request completed successfully... containing a total of [%d] currency pairs", len(symbolList))
	go func() {
		time.Sleep(time.Minute * 10)
		logger.SpotMarket.Debugf("re-updating spot currency pairs........")
		spotMarket.GetSpotSymbols()
	}()
	return nil
}

// initialize spot market
func (spotMarket *SpotMarket) Init(ctx context.Context) error {
	logger.SpotMarket.Debug("starting initialization of spot market symbols...")
	err := spotMarket.GetSpotSymbols() // retrieve spot currency pairs
	if err != nil {
		return err
	}
	_ = spotMarket.ProcessMarket() // commence processing market data

	<-ctx.Done() // listen for launcher exit or cancellation
	logger.SpotMarket.Debug(".......Spot Market Manager.goroutine exit")
	return nil
}
