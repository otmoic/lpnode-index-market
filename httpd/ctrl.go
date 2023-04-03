package httpd

import (
	"fmt"
	"lp_market/logger"
	"lp_market/market"
	"lp_market/types"
	"strings"

	"github.com/bytedance/sonic"
	routing "github.com/qiangxue/fasthttp-routing"
)

type Ctrl struct {
}

var marketCenter *market.MarketCenter

func (h *Ctrl) Init() {
	marketCenter = &market.MarketCenter{}
}

// ResetSpot 重置现货的订阅
func (h *Ctrl) ResetSpot(c *routing.Context) error {
	logger.Httpd.Debug(`重置现货...Orderbook链接`)
	marketCenter.GetSpotIns().RefreshMarket() // 重新刷新现货币对列表
	return nil
}

// ResetUsdtSwap 重置u本位的订阅
func (h *Ctrl) ResetUsdtSwap(c *routing.Context) error {
	logger.Httpd.Debug(`重置Usdt Swap...Orderbook链接`)
	marketCenter.GetUsdtSwapIns().RefreshMarket() // 重新刷新u本位订阅的链接
	return nil
}

// ResetCoinSwap 重置币本位的订阅
func (h *Ctrl) ResetCoinSwap(c *routing.Context) error {
	logger.Httpd.Debug(`重置Coin Swap...Orderbook链接`)
	marketCenter.GetCoinSwapIns().RefreshMarket() // 重新刷新币本位订阅的链接
	return nil
}

// SetSportSymbols 设置现货的可用币对
func (h *Ctrl) SetSpotSymbols(c *routing.Context) error {
	logger.Httpd.Debug("开始设置Symbol列表")
	body := c.Request.Body()
	var dataArr []string = []string{}
	sonic.Unmarshal(body, &dataArr)
	marketCenter.GetSpotIns().SetUsedSymbol(dataArr)
	marketCenter.GetSpotIns().RefreshMarket() // 重新刷新现货币对列表
	return nil
}
func (h *Ctrl) SetUsdtSwapSymbols(c *routing.Context) error {
	logger.Httpd.Debug("开始设置Symbol列表")
	body := c.Request.Body()
	var dataArr []string = []string{}
	sonic.Unmarshal(body, &dataArr)
	marketCenter.GetUsdtSwapIns().SetUsedSymbol(dataArr)
	marketCenter.GetUsdtSwapIns().RefreshMarket() // 重新刷新现货币对列表
	return nil
}
func (h *Ctrl) GetUsdtSwapFundingRate(c *routing.Context) error {
	symbolByte := c.QueryArgs().Peek("symbol")
	retFundingRate := &types.RetFundingRateItem{}
	symbol := strings.ToUpper(string(symbolByte))
	retFundingRate.Data = marketCenter.GetFundingRateIns().GetUsdtFundingRate(symbol)
	apiJsonStr, err := sonic.Marshal(retFundingRate)
	if err != nil {
		logger.Httpd.Errorf("处理数据发生了错误%s", err)
	}
	fmt.Fprintln(c, string(apiJsonStr))
	return nil
}
func (h *Ctrl) GetCoinSwapFundingRate(c *routing.Context) error {
	symbolByte := c.QueryArgs().Peek("symbol")
	retFundingRate := &types.RetFundingRateItem{}
	symbol := strings.ToUpper(string(symbolByte))
	retFundingRate.Data = marketCenter.GetFundingRateIns().GetCoinFundingRate(symbol)
	apiJsonStr, err := sonic.Marshal(retFundingRate)
	if err != nil {
		logger.Httpd.Errorf("处理数据发生了错误%s", err)
	}
	fmt.Fprintln(c, string(apiJsonStr))
	return nil
}

func (h *Ctrl) SetCoinSwapSymbols(c *routing.Context) error {
	logger.Httpd.Debug("开始设置Symbol列表")
	body := c.Request.Body()
	var dataArr []string = []string{}
	sonic.Unmarshal(body, &dataArr)
	marketCenter.GetCoinSwapIns().SetUsedSymbol(dataArr)
	marketCenter.GetCoinSwapIns().RefreshMarket() // 重新刷新现货币对列表
	return nil
}

// 返回现货Orderbook的数据
func (h *Ctrl) GetSpotOrderbookApiData(c *routing.Context) error {
	orderbook := marketCenter.GetSpotOrderbookIns().GetOrderbook()
	ret := &types.RetOrderbookMessage{
		Data: make(map[string]*types.OrderBookItem, 0),
	}
	orderbook.Range(func(key, value any) bool {
		orderbookItem := value.(*types.OrderBookItem)
		ret.Data[orderbookItem.StdSymbol] = orderbookItem
		return true
	})
	apiJsonStr, err := sonic.Marshal(ret)
	if err != nil {
		logger.Httpd.Errorf("处理数据发生了错误%s", err)
	}
	fmt.Fprintln(c, string(apiJsonStr))
	return nil
}
func (h *Ctrl) GetUsdtSwapOrderbookApiData(c *routing.Context) error {
	orderbook := marketCenter.GetUsdtSwapOrderbookIns().GetOrderbook()
	ret := &types.RetOrderbookMessage{
		Data: make(map[string]*types.OrderBookItem, 0),
	}
	orderbook.Range(func(key, value any) bool {
		orderbookItem := value.(*types.OrderBookItem)
		ret.Data[orderbookItem.StdSymbol] = orderbookItem
		return true
	})
	apiJsonStr, err := sonic.Marshal(ret)
	if err != nil {
		logger.Httpd.Errorf("处理数据发生了错误%s", err)
	}
	fmt.Fprintln(c, string(apiJsonStr))
	return nil
}
func (h *Ctrl) GetCoinSwapOrderbookApiData(c *routing.Context) error {
	orderbook := marketCenter.GetCoinSwapOrderbookIns().GetOrderbook()
	ret := &types.RetOrderbookMessage{
		Data: make(map[string]*types.OrderBookItem, 0),
	}
	orderbook.Range(func(key, value any) bool {
		orderbookItem := value.(*types.OrderBookItem)
		ret.Data[orderbookItem.StdSymbol] = orderbookItem
		return true
	})
	apiJsonStr, err := sonic.Marshal(ret)
	if err != nil {
		logger.Httpd.Errorf("处理数据发生了错误%s", err)
	}
	fmt.Fprintln(c, string(apiJsonStr))
	return nil
}

func (h *Ctrl) GetSpotOrderbook(c *routing.Context) error {
	fmt.Fprintln(c, "SpotOrderBook")
	orderbook := marketCenter.GetSpotOrderbookIns().GetOrderbook()
	list := marketCenter.GetSpotIns().GetGlobalUsedSymbolList()
	list.Range(func(key, value any) bool {
		usedSymbolVal := value.(types.ExchangeInfoSymbolApiResult)
		item, ok := orderbook.Load(usedSymbolVal.StdSymbol)
		if !ok {
			fmt.Fprintln(c, "币对 无行情Symbol", usedSymbolVal.StdSymbol)
			return true
		}
		v := item.(*types.OrderBookItem)
		lineInfo := fmt.Sprintf("币对 【%s】,Asks0 Price %s Ask0 Amount %s  Bids0 Price %s Bids0 Amount %s", v.StdSymbol, v.Asks[0][0], v.Asks[0][1], v.Bids[0][0], v.Bids[0][1])
		fmt.Fprintln(c, lineInfo)
		return true
	})

	return nil
}
func (h *Ctrl) GetUsdtSwapOrderbook(c *routing.Context) error {
	fmt.Fprintln(c, "UsdtSwapOrderbook")
	orderbook := marketCenter.GetUsdtSwapOrderbookIns().GetOrderbook()
	list := marketCenter.GetUsdtSwapIns().GetGlobalUsedSymbolList()
	list.Range(func(key, value any) bool {
		usedSymbolVal := value.(types.ExchangeInfoSymbolApiResult)
		item, ok := orderbook.Load(usedSymbolVal.StdSymbol)
		if !ok {
			fmt.Fprintln(c, "币对 无行情Symbol", usedSymbolVal.StdSymbol)
			return true
		}
		v := item.(*types.OrderBookItem)
		lineInfo := fmt.Sprintf("币对 【%s】,Asks0 Price %s Ask0 Amount %s  Bids0 Price %s Bids0 Amount %s", v.StdSymbol, v.Asks[0][0], v.Asks[0][1], v.Bids[0][0], v.Bids[0][1])
		fmt.Fprintln(c, lineInfo)
		return true
	})
	return nil
}
func (h *Ctrl) GetCoinSwapOrderbook(c *routing.Context) error {
	fmt.Fprintln(c, "CoinSwapOrderbook")
	orderbook := marketCenter.GetCoinSwapOrderbookIns().GetOrderbook()
	list := marketCenter.GetCoinSwapIns().GetGlobalUsedSymbolList()
	list.Range(func(key, value any) bool {
		usedSymbolVal := value.(types.ExchangeInfoSymbolApiResult)
		item, ok := orderbook.Load(usedSymbolVal.StdSymbol)
		if !ok {
			fmt.Fprintln(c, "币对 无行情Symbol", usedSymbolVal.StdSymbol)
			return true
		}
		v := item.(*types.OrderBookItem)
		lineInfo := fmt.Sprintf("币对 【%s】,Asks0 Price %s Ask0 Amount %s  Bids0 Price %s Bids0 Amount %s", v.StdSymbol, v.Asks[0][0], v.Asks[0][1], v.Bids[0][0], v.Bids[0][1])
		fmt.Fprintln(c, lineInfo)
		return true
	})
	return nil
}

func (h *Ctrl) RegRouter(r *routing.Router) {
	// 重新订阅目前内存中的币对列表
	r.Get("/resetSpotSymbolList", func(c *routing.Context) error {
		return h.ResetSpot(c)
	})
	r.Get("/resetUsdtSwapSymbolList", func(c *routing.Context) error {
		return h.ResetUsdtSwap(c)
	})
	r.Get("/resetCoinSwapSymbolList", func(c *routing.Context) error {
		return h.ResetCoinSwap(c)
	})
	// 设置币对列表的相关接口
	r.Post("/setSpotSymbolList", func(c *routing.Context) error {
		return h.SetSpotSymbols(c)
	})
	r.Post("/setUsdtSwapSymbolList", func(c *routing.Context) error {
		return h.SetUsdtSwapSymbols(c)
	})
	r.Post("/setCoinSwapSymbolList", func(c *routing.Context) error {
		return h.SetCoinSwapSymbols(c)
	})

	// api orderbook部分
	r.Get("/api/spotOrderbook", func(c *routing.Context) error {
		return h.GetSpotOrderbookApiData(c)
	})
	r.Get("/api/usdtSwapOrderbook", func(c *routing.Context) error {
		return h.GetUsdtSwapOrderbookApiData(c)
	})
	r.Get("/api/coinSwapOrderbook", func(c *routing.Context) error {
		return h.GetCoinSwapOrderbookApiData(c)
	})

	// view orderbook部分
	r.Get("/spotOrderbook", func(c *routing.Context) error {
		return h.GetSpotOrderbook(c)
	})
	r.Get("/usdtSwapOrderbook", func(c *routing.Context) error {
		return h.GetUsdtSwapOrderbook(c)
	})
	r.Get("/coinSwapOrderbook", func(c *routing.Context) error {
		return h.GetCoinSwapOrderbook(c)
	})
	// 资金费率接口
	r.Get("/getUsdtSwapFundingRate", func(c *routing.Context) error {
		return h.GetUsdtSwapFundingRate(c)
	})
	r.Get("/getCoinSwapFundingRate", func(c *routing.Context) error {
		return h.GetCoinSwapFundingRate(c)
	})
}
