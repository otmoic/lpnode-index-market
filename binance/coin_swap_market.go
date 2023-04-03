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
	symbolList                    sync.Map        // 保存已经获取好的 symbol 列表
	symbolSetList                 map[string]bool // 设置好的白名单
	usedSymbolList                sync.Map        // 需要订阅的币对列表 GRSBTC 内容 {StdSymbol:GRS/BTC}
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

// PreMarketSymbol 如果币对在白名单设置中，才设置，准备可以订阅的币对
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
		logger.SpotMarket.Debug("发现了一个【GroupManager】", key.(int64), value.(*CoinSwapMarketGroupConn))
		smgc := value.(*CoinSwapMarketGroupConn)
		logger.SpotMarket.Debugf("删除引用...id%d....%s,", key.(int64), smgc.GetSymbolsString())
		smgc.Drop()
		csm.CoinSwapGroupManagerListStore.Delete(key)
		return true
	})
	go func() {
		for x := 0; x < 5; x++ {
			time.Sleep(time.Second * 2)
			runtime.GC()
		}
		csm.ProcessMarket() // 重新开始订阅
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
			logger.USwapMarket.Debug("开始处理Type[Coin_swap]", processIndex)
			csm.RunList(list, processIndex)
			clearList()
		}
		return true
	})
	if len(list) != 0 { //剩下的没有处理的
		processIndex++
		logger.SpotMarket.Debug("开始处理", processIndex)
		csm.RunList(list, processIndex)
		clearList()
	}

	return nil
}
func (csm *CoinSwapMarket) RunList(list []types.ExchangeInfoSymbolApiResult, processIndex int64) {
	smgc := &CoinSwapMarketGroupConn{}
	smgc.Init()
	smgc.SetSymbolList(list)
	smgc.Run()                                                        // 运行
	runtime.SetFinalizer(smgc, func(csmgc *CoinSwapMarketGroupConn) { // 设置一个Gc 观察期，开发期间用来观察Gc
		logger.SpotMarket.Debugf("💢💢💢💢💢 Gc SymbolList Is:%s", csmgc.GetSymbolsString())
	})
	csm.CoinSwapGroupManagerListStore.Store(processIndex, smgc)
}
func (csm *CoinSwapMarket) GetSymbols() error {
	logger.CSwapMarket.Debug("准备获取币本位合约的基本币对信息")
	url := fmt.Sprintf("%s%s", CoinSwapMarketHttpsBaseUrl, CoinSwapExchangeInfoPath)
	logger.CSwapMarket.Debugf("交易标准信息路径是:%s", url)
	_, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 {
		return errors.New(fmt.Errorf("请求发生了错误:%s", errs[0]).Error())
	}
	var ret types.ExchangeInfoApiResult
	err := sonic.Unmarshal([]byte(body), &ret) // 解码json 内容
	if err != nil {
		return errors.New(fmt.Errorf("解码发生了错误:%s", err).Error())
	}
	symbolList := FormatterCoinSwapExchangeInfo(ret)
	for _, v := range symbolList {
		// logger.CSwapMarket.Debugf("%s", v.StdSymbol)
		csm.symbolList.Store(strings.ToLower(v.Symbol), v) // 把所有的币对存起来
	}
	CoinSwapSymbolList_Global = csm.symbolList
	logger.USwapMarket.Debugf("Coin本位合约的币对信息已经请求完毕共计【%d】个币对", len(symbolList))
	return nil
}
func (csm *CoinSwapMarket) SetUsedSymbol(symbolList []string) {
	for _, v := range symbolList {
		csm.symbolSetList[v] = true
	}
}
func (csm *CoinSwapMarket) Init(ctx context.Context) error {
	logger.USwapMarket.Debug("开始初始化Cswap的Symbol..!")
	err := csm.GetSymbols()
	if err != nil {
		return err
	}
	_ = csm.ProcessMarket() // 开始处理行情

	<-ctx.Done() // 监听启动器的退出 和cancel
	logger.USwapMarket.Debug(".......CoinSwapMarket  Manager.协程退出")
	return nil
}

// GetCoinSwapMarketInstance 单例返回  币本位Swap市场实例
func GetCoinSwapMarketInstance() stdmarket.StdCoinSwapMarket {
	coinSwapMarketOnce.Do(func() {
		coinSwapMarketInstance = &CoinSwapMarket{}
		coinSwapMarketInstance.InitSymbolList()
	})
	return coinSwapMarketInstance

}
