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

var CoinSwapSymbolList_Global sync.Map
var CoinSwapUsedSymbolList_Global sync.Map

type CoinSwapMarket struct {
	symbolList                    sync.Map        // save the already obtained symbol list
	symbolSetList                 map[string]bool // configured whitelist
	usedSymbolList                sync.Map        // content for currency pairs to subscribe: GRSBTC {StdSymbol: GRS/BTC}
	CoinSwapGroupManagerListStore sync.Map
}

var coinSwapMarketInstance stdmarket.StdCoinSwapMarket

func (csm *CoinSwapMarket) GetGlobalUsedSymbolList() sync.Map {
	return CoinSwapUsedSymbolList_Global
}
func (csm *CoinSwapMarket) GetStdSymbol(symbol string) string {
	v, ok := CoinSwapSymbolList_Global.Load(strings.ToLower(symbol) + "_perp")
	if !ok {
		return ""
	}
	symbolInfo := v.(types.ExchangeInfoSymbolApiResult)
	return symbolInfo.StdSymbol
}
func (csm *CoinSwapMarket) InitSymbolList() {
	csm.symbolSetList = make(map[string]bool)
}

// PreMarketSymbol only set and prepare subscribable currency pairs if they are in the whitelist settings
func (csm *CoinSwapMarket) PreMarketSymbol() {
	csm.symbolList.Range(func(key, value interface{}) bool {
		stdSymbolStruct := value.(types.ExchangeInfoSymbolApiResult)
		_, ok := csm.symbolSetList[stdSymbolStruct.StdSymbol]
		if ok {
			csm.usedSymbolList.Store(stdSymbolStruct.Symbol, stdSymbolStruct)
		}
		return true
	})
	CoinSwapUsedSymbolList_Global = csm.usedSymbolList
}
func (csm *CoinSwapMarket) RefreshMarket() error {
	csm.CoinSwapGroupManagerListStore.Range(func(key, value any) bool {
		logger.SpotMarket.Debug("found a [GroupManager]", key.(int64), value.(*CoinSwapMarketGroupConn))
		smgc := value.(*CoinSwapMarketGroupConn)
		logger.SpotMarket.Debugf("delete reference...id%d....%s,", key.(int64), smgc.GetSymbolsString())
		smgc.Drop()
		csm.CoinSwapGroupManagerListStore.Delete(key)
		return true
	})
	go func() {
		for x := 0; x < 5; x++ {
			time.Sleep(time.Second * 2)
			runtime.GC()
		}
		csm.ProcessMarket() // restart subscription
	}()
	return nil
}
func (csm *CoinSwapMarket) ProcessMarket() error {
	csm.PreMarketSymbol()
	var sIndex int64 = 0
	var processIndex int64 = 0
	list := []types.ExchangeInfoSymbolApiResult{}
	clearList := func() {
		list = []types.ExchangeInfoSymbolApiResult{}
	}
	csm.usedSymbolList.Range(func(key interface{}, value interface{}) bool {
		sIndex++
		list = append(list, value.(types.ExchangeInfoSymbolApiResult))
		if sIndex%5 == 0 {
			processIndex++
			logger.USwapMarket.Debug("start processing type[Coin_swap]", processIndex)
			csm.RunList(list, processIndex)
			clearList()
		}
		return true
	})
	if len(list) != 0 { //remaining unprocessed
		processIndex++
		logger.SpotMarket.Debug("start processing", processIndex)
		csm.RunList(list, processIndex)
		clearList()
	}

	return nil
}
func (csm *CoinSwapMarket) RunList(list []types.ExchangeInfoSymbolApiResult, processIndex int64) {
	smgc := &CoinSwapMarketGroupConn{}
	smgc.Init()
	smgc.SetSymbolList(list)
	smgc.Run()
	runtime.SetFinalizer(smgc, func(csmgc *CoinSwapMarketGroupConn) { //	set a gc observation period for observing gc during development
		logger.SpotMarket.Debugf("💢💢💢💢💢 Gc SymbolList Is:%s", csmgc.GetSymbolsString())
	})
	csm.CoinSwapGroupManagerListStore.Store(processIndex, smgc)
}
func (csm *CoinSwapMarket) GetSymbols() error {
	logger.CSwapMarket.Debug("prepare to retrieve basic currency pair information for coin-based contracts")
	url := fmt.Sprintf("%s%s", CoinSwapMarketHttpsBaseUrl, CoinSwapExchangeInfoPath)
	logger.CSwapMarket.Debugf("trading standard information path is:%s", url)
	_, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 {
		return errors.New(fmt.Errorf("request encountered an error:%s", errs[0]).Error())
	}
	var ret types.ExchangeInfoApiResult
	err := sonic.Unmarshal([]byte(body), &ret) // decode json content
	if err != nil {
		return errors.New(fmt.Errorf("decoding error occurred:%s", err).Error())
	}
	symbolList := FormatterCoinSwapExchangeInfo(ret)
	for _, v := range symbolList {
		// logger.CSwapMarket.Debugf("%s", v.StdSymbol)
		csm.symbolList.Store(strings.ToLower(v.Symbol), v) // store all currency pairs
	}
	CoinSwapSymbolList_Global = csm.symbolList
	logger.USwapMarket.Debugf("coin-based contract's currency pair information retrieval completed, total of 【%d】 pair", len(symbolList))
	return nil
}
func (csm *CoinSwapMarket) SetUsedSymbol(symbolList []string) {
	for _, v := range symbolList {
		csm.symbolSetList[v] = true
	}
}
func (csm *CoinSwapMarket) Init(ctx context.Context) error {
	logger.USwapMarket.Debug("start initializing Cswap's symbol..!")
	err := csm.GetSymbols()
	if err != nil {
		return err
	}
	_ = csm.ProcessMarket() // process market info

	<-ctx.Done() // listen for launcher exit and cancel
	logger.USwapMarket.Debug(".......CoinSwapMarket  Manager.fiber exit")
	return nil
}

// GetCoinSwapMarketInstance singleton returns coin-based Swap market instance
func GetCoinSwapMarketInstance() stdmarket.StdCoinSwapMarket {
	coinSwapMarketOnce.Do(func() {
		coinSwapMarketInstance = &CoinSwapMarket{}
		coinSwapMarketInstance.InitSymbolList()
	})
	return coinSwapMarketInstance

}
