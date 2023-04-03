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
	symbolList                sync.Map        // 保存已经获取好的 symbol 列表
	symbolSetList             map[string]bool // 设置好的白名单
	usedSymbolList            sync.Map        // 需要订阅的币对列表 GRSBTC 内容 {StdSymbol:GRS/BTC}
	SpotGroupManagerListStore sync.Map
}

var spotMarketInstance stdmarket.StdSpotMarket

// 单例返回  现货市场实例
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
			logger.SpotMarket.Debugf("这个可以处理%s", stdSymbolStruct.StdSymbol)
		}
		return true
	})
	SpotUsedSymbolList_Global = spotMarket.usedSymbolList

}
func (spotMarket *SpotMarket) RefreshMarket() error {
	spotMarket.SpotGroupManagerListStore.Range(func(key, value any) bool {
		logger.SpotMarket.Debug("发现了一个【GroupManager】", key.(int64), value.(*SpotMarketGroupConn))
		smgc := value.(*SpotMarketGroupConn)
		logger.SpotMarket.Debug("删除引用.......", smgc.GetSymbolsString())
		smgc.Drop()
		spotMarket.SpotGroupManagerListStore.Delete(key)
		return true
	})
	go func() {
		for x := 0; x < 5; x++ {
			time.Sleep(time.Second * 2)
			runtime.GC()
		}
		spotMarket.ProcessMarket() // 重新开始订阅
	}()
	return nil
}

// 开始处理行情数据
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
			logger.SpotMarket.Debug("开始处理", processIndex)
			spotMarket.RunList(list, processIndex)
			clearList()
		}
		return true
	})
	if len(list) != 0 { //剩下的没有处理的
		processIndex++
		logger.SpotMarket.Debug("开始处理", processIndex)
		spotMarket.RunList(list, processIndex)
		clearList()
	}

	return nil
}
func (spotMarket *SpotMarket) RunList(list []types.ExchangeInfoSymbolApiResult, processIndex int64) {
	smgc := &SpotMarketGroupConn{}
	smgc.Init()
	smgc.SetSymbolList(list)
	smgc.Run() // 运行
	runtime.SetFinalizer(smgc, func(smgc *SpotMarketGroupConn) {
		logger.SpotMarket.Debugf("💢💢💢💢💢 Gc SymbolList Is:%s", smgc.GetSymbolsString())
	})
	spotMarket.SpotGroupManagerListStore.Store(processIndex, smgc)
}

// 获取所有现货的币对,并Set 到List中
func (spotMarket *SpotMarket) GetSpotSymbols() error {
	url := fmt.Sprintf("%s%s", SpotMarketHttpsBaseUrl, ExchangeInfoPath)
	logger.SpotMarket.Debugf("交易标准信息路径是:%s", url)
	_, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 {
		return errors.New(fmt.Errorf("请求发生了错误:%s", errs[0]).Error())
	}
	var ret types.ExchangeInfoApiResult
	err := sonic.Unmarshal([]byte(body), &ret)
	if err != nil {
		return errors.New(fmt.Errorf("解码发生了错误:%s", err).Error())
	}
	symbolList := FormatterSpotExchangeInfo(ret)
	for _, v := range symbolList {
		// logger.SpotMarket.Debugf("开始存储%s", v.StdSymbol)
		spotMarket.symbolList.Store(strings.ToLower(v.Symbol), v)
	}
	SpotSymbolList_Global = spotMarket.symbolList // 把取到的Market symbol 放到公开的地方
	logger.SpotMarket.Debugf("请求已经正常完成....,一共有【%d】个币对", len(symbolList))
	go func() {
		time.Sleep(time.Minute * 10)
		logger.SpotMarket.Debugf("重新更新现货币对........")
		spotMarket.GetSpotSymbols()
	}()
	return nil
}

// 初始化现货
func (spotMarket *SpotMarket) Init(ctx context.Context) error {
	logger.SpotMarket.Debug("开始初始化现货的Symbol..!")
	err := spotMarket.GetSpotSymbols() // 获取现货币对
	if err != nil {
		return err
	}
	_ = spotMarket.ProcessMarket() // 开始处理行情

	<-ctx.Done() // 监听启动器的退出 和cancel
	logger.SpotMarket.Debug(".......Spot Market Manager.协程退出")
	return nil
}
