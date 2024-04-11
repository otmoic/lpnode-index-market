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

func (h *Ctrl) ResetSpot(c *routing.Context) error {
	logger.Httpd.Debug(`reset spot...Orderbook`)
	marketCenter.GetSpotIns().RefreshMarket()
	return nil
}

func (h *Ctrl) ResetUsdtSwap(c *routing.Context) error {
	logger.Httpd.Debug(`reset Usdt Swap...Orderbook`)
	marketCenter.GetUsdtSwapIns().RefreshMarket()
	return nil
}

func (h *Ctrl) ResetCoinSwap(c *routing.Context) error {
	logger.Httpd.Debug(`reset Coin Swap...Orderbook`)
	marketCenter.GetCoinSwapIns().RefreshMarket()
	return nil
}

func (h *Ctrl) SetSpotSymbols(c *routing.Context) error {
	logger.Httpd.Debug("set Symbol list")
	body := c.Request.Body()
	var dataArr []string = []string{}
	sonic.Unmarshal(body, &dataArr)
	marketCenter.GetSpotIns().SetUsedSymbol(dataArr)
	marketCenter.GetSpotIns().RefreshMarket()
	return nil
}
func (h *Ctrl) SetUsdtSwapSymbols(c *routing.Context) error {
	logger.Httpd.Debug("set Symbol list")
	body := c.Request.Body()
	var dataArr []string = []string{}
	sonic.Unmarshal(body, &dataArr)
	marketCenter.GetUsdtSwapIns().SetUsedSymbol(dataArr)
	marketCenter.GetUsdtSwapIns().RefreshMarket()
	return nil
}
func (h *Ctrl) GetUsdtSwapFundingRate(c *routing.Context) error {
	symbolByte := c.QueryArgs().Peek("symbol")
	retFundingRate := &types.RetFundingRateItem{}
	symbol := strings.ToUpper(string(symbolByte))
	retFundingRate.Data = marketCenter.GetFundingRateIns().GetUsdtFundingRate(symbol)
	apiJsonStr, err := sonic.Marshal(retFundingRate)
	if err != nil {
		logger.Httpd.Errorf("error processing data %s", err)
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
		logger.Httpd.Errorf("error processing data %s", err)
	}
	fmt.Fprintln(c, string(apiJsonStr))
	return nil
}

func (h *Ctrl) SetCoinSwapSymbols(c *routing.Context) error {
	logger.Httpd.Debug("nitializing Symbol List Setup")
	body := c.Request.Body()
	var dataArr []string = []string{}
	sonic.Unmarshal(body, &dataArr)
	marketCenter.GetCoinSwapIns().SetUsedSymbol(dataArr)
	marketCenter.GetCoinSwapIns().RefreshMarket()
	return nil
}

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
		logger.Httpd.Errorf("An error occurred while processing the data %s", err)
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
		logger.Httpd.Errorf("An error occurred while processing the data %s", err)
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
		logger.Httpd.Errorf("An error occurred while processing the data %s", err)
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
			fmt.Fprintln(c, "currency pair with no market data Symbol", usedSymbolVal.StdSymbol)
			return true
		}
		v := item.(*types.OrderBookItem)
		lineInfo := fmt.Sprintf("pair 【%s】,Asks0 Price %s Ask0 Amount %s  Bids0 Price %s Bids0 Amount %s", v.StdSymbol, v.Asks[0][0], v.Asks[0][1], v.Bids[0][0], v.Bids[0][1])
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
			fmt.Fprintln(c, "currency pair with no market data Symbol", usedSymbolVal.StdSymbol)
			return true
		}
		v := item.(*types.OrderBookItem)
		lineInfo := fmt.Sprintf("symbol 【%s】,Asks0 Price %s Ask0 Amount %s  Bids0 Price %s Bids0 Amount %s", v.StdSymbol, v.Asks[0][0], v.Asks[0][1], v.Bids[0][0], v.Bids[0][1])
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
			fmt.Fprintln(c, "currency pair with no market data Symbol", usedSymbolVal.StdSymbol)
			return true
		}
		v := item.(*types.OrderBookItem)
		lineInfo := fmt.Sprintf("pair 【%s】,Asks0 Price %s Ask0 Amount %s  Bids0 Price %s Bids0 Amount %s", v.StdSymbol, v.Asks[0][0], v.Asks[0][1], v.Bids[0][0], v.Bids[0][1])
		fmt.Fprintln(c, lineInfo)
		return true
	})
	return nil
}

func (h *Ctrl) RegRouter(r *routing.Router) {
	r.Get("/resetSpotSymbolList", func(c *routing.Context) error {
		return h.ResetSpot(c)
	})
	r.Get("/resetUsdtSwapSymbolList", func(c *routing.Context) error {
		return h.ResetUsdtSwap(c)
	})
	r.Get("/resetCoinSwapSymbolList", func(c *routing.Context) error {
		return h.ResetCoinSwap(c)
	})

	r.Post("/setSpotSymbolList", func(c *routing.Context) error {
		return h.SetSpotSymbols(c)
	})
	r.Post("/setUsdtSwapSymbolList", func(c *routing.Context) error {
		return h.SetUsdtSwapSymbols(c)
	})
	r.Post("/setCoinSwapSymbolList", func(c *routing.Context) error {
		return h.SetCoinSwapSymbols(c)
	})

	// api orderbook
	r.Get("/api/spotOrderbook", func(c *routing.Context) error {
		return h.GetSpotOrderbookApiData(c)
	})
	r.Get("/api/usdtSwapOrderbook", func(c *routing.Context) error {
		return h.GetUsdtSwapOrderbookApiData(c)
	})
	r.Get("/api/coinSwapOrderbook", func(c *routing.Context) error {
		return h.GetCoinSwapOrderbookApiData(c)
	})

	// view orderbook
	r.Get("/spotOrderbook", func(c *routing.Context) error {
		return h.GetSpotOrderbook(c)
	})
	r.Get("/usdtSwapOrderbook", func(c *routing.Context) error {
		return h.GetUsdtSwapOrderbook(c)
	})
	r.Get("/coinSwapOrderbook", func(c *routing.Context) error {
		return h.GetCoinSwapOrderbook(c)
	})

	r.Get("/getUsdtSwapFundingRate", func(c *routing.Context) error {
		return h.GetUsdtSwapFundingRate(c)
	})
	r.Get("/getCoinSwapFundingRate", func(c *routing.Context) error {
		return h.GetCoinSwapFundingRate(c)
	})
}
