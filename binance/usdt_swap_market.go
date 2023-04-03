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

var UsdtSwapSymbolList_Global sync.Map     // 全局共享的u本位合约币对信息
var UsdtSwapUsedSymbolList_Global sync.Map // 全局共享的，当前设置的订阅的币对信息
type USwapMarket struct {
	symbolList                    sync.Map
	symbolSetList                 map[string]bool
	usedSymbolList                sync.Map
	UsdtSwapGroupManagerListStore sync.Map // 存储 u 永续 分组订阅实例
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

// 如果币对在白名单设置中，才设置，准备可以订阅的币对
func (usm *USwapMarket) PreMarketSymbol() {
	usm.symbolList.Range(func(key, value interface{}) bool {
		stdSymbolStruct := value.(types.ExchangeInfoSymbolApiResult)
		_, ok := usm.symbolSetList[stdSymbolStruct.StdSymbol]
		if ok {
			// logger.USwapMarket.Debugf("🍅🍅🍅🍅🍅🍅🍅🍅🍅这个可以处理%s", stdSymbolStruct.StdSymbol)
			usm.usedSymbolList.Store(stdSymbolStruct.Symbol, stdSymbolStruct)
		}
		return true
	})
	UsdtSwapUsedSymbolList_Global = usm.usedSymbolList
}

// 删除目前内存中的订阅，并重新开始
func (usm *USwapMarket) RefreshMarket() error {
	usm.UsdtSwapGroupManagerListStore.Range(func(key, value any) bool {
		logger.USwapMarket.Debug("发现了一个【GroupManager】", key.(int64), value.(*UsdtSwapMarketGroupConn))
		smgc := value.(*UsdtSwapMarketGroupConn)
		logger.USwapMarket.Debug("删除引用.......", smgc.GetSymbolsString())
		smgc.Drop()
		usm.UsdtSwapGroupManagerListStore.Delete(key)
		return true
	})
	go func() {
		for x := 0; x < 5; x++ {
			time.Sleep(time.Second * 2)
			runtime.GC()
		}
		usm.ProcessMarket() // 重新开始订阅
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
			logger.USwapMarket.Debug("开始处理Type[Uswap]", processIndex)
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

// @todo 这里需要一个定时更新币对的功能，暂时先不加上
func (usm *USwapMarket) GetSymbols() error {
	url := fmt.Sprintf("%s%s", UsdtSwapMarketHttpsBaseUrl, UsdtSwapExchangeInfoPath)
	logger.USwapMarket.Debugf("交易标准信息路径是:%s", url)
	_, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 {
		return errors.New(fmt.Errorf("请求发生了错误:%s", errs[0]).Error())
	}
	var ret types.ExchangeInfoApiResult
	err := sonic.Unmarshal([]byte(body), &ret) // 解码json 内容
	if err != nil {
		return errors.New(fmt.Errorf("解码发生了错误:%s", err).Error())
	}
	symbolList := FormatterUsdtSwapExchangeInfo(ret)
	for _, v := range symbolList {
		usm.symbolList.Store(strings.ToLower(v.Symbol), v) // 把所有的币对存起来
	}
	UsdtSwapSymbolList_Global = usm.symbolList
	logger.USwapMarket.Debugf("U本位合约的币对信息已经请求完毕共计【%d】个币对", len(ret.Symbols))
	return nil

}
func (usm *USwapMarket) SetUsedSymbol(symbolList []string) {
	for _, v := range symbolList {
		usm.symbolSetList[v] = true
	}
}
func (usm *USwapMarket) Init(ctx context.Context) error {
	logger.USwapMarket.Debug("开始初始化Uswap的Symbol..!")
	err := usm.GetSymbols()
	if err != nil {
		return err
	}
	err = usm.ProcessMarket() // 开始处理行情
	if err != nil {
		logger.USwapMarket.Errorf("处理行情发生了错误%s", err)
	}

	<-ctx.Done() // 监听启动器的退出 和cancel
	logger.USwapMarket.Debug(".......UsdtSwapMarket  Manager.协程退出")

	return nil
}
func (usm *USwapMarket) InitSymbolList() {
	usm.symbolSetList = make(map[string]bool)
}

// 单例返回  U本位 Swap市场实例
func GetUSwapMarketInstance() stdmarket.StdUsdtSwapMarket {
	usdtSwapMarketOnce.Do(func() {
		uSwapMarketInstance = &USwapMarket{}
		uSwapMarketInstance.InitSymbolList()
	})
	return uSwapMarketInstance

}
