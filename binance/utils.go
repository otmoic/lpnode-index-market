package binance

import (
	"fmt"
	"lp_market/types"
)

func FormatterSpotExchangeInfo(ret types.ExchangeInfoApiResult) []types.ExchangeInfoSymbolApiResult {
	var retSymbolList []types.ExchangeInfoSymbolApiResult = make([]types.ExchangeInfoSymbolApiResult, 0)

	for i := range ret.Symbols {
		ret.Symbols[i].StdSymbol = fmt.Sprintf("%s/%s", ret.Symbols[i].BaseAsset, ret.Symbols[i].QuoteAsset)
		if ret.Symbols[i].Status != "TRADING" {
			// logger.SpotMarket.Warnf("Skip Symbol【%s】Status:【%s】", ret.Symbols[i].StdSymbol, v.Status)
			continue // 跳过目前不在交易中的币对
		}
		retSymbolList = append(retSymbolList, ret.Symbols[i])
	}
	return retSymbolList
}

// @todo 这里的交割合约可能也有问题,标准币对信息可能存储了不正确的值
func FormatterUsdtSwapExchangeInfo(ret types.ExchangeInfoApiResult) []types.ExchangeInfoSymbolApiResult {
	for i := range ret.Symbols {
		ret.Symbols[i].StdSymbol = fmt.Sprintf("%s/%s", ret.Symbols[i].BaseAsset, ret.Symbols[i].QuoteAsset)
	}
	return ret.Symbols
}

func FormatterCoinSwapExchangeInfo(ret types.ExchangeInfoApiResult) []types.ExchangeInfoSymbolApiResult {
	var retSymbolList []types.ExchangeInfoSymbolApiResult = make([]types.ExchangeInfoSymbolApiResult, 0)
	for i := range ret.Symbols {
		if ret.Symbols[i].ContractType == "PERPETUAL" { // 暂时只保留永续合约
			ret.Symbols[i].StdSymbol = fmt.Sprintf("%s/%s", ret.Symbols[i].BaseAsset, ret.Symbols[i].QuoteAsset)
			retSymbolList = append(retSymbolList, ret.Symbols[i])
		} else {
			// logger.CSwapMarket.Warnf("忽略掉交割合约,或者Prep:%s", ret.Symbols[i].Symbol)
		}

	}
	return retSymbolList
}
