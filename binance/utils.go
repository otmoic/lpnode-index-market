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
			continue // skip currency pairs currently not trading
		}
		retSymbolList = append(retSymbolList, ret.Symbols[i])
	}
	return retSymbolList
}

func FormatterUsdtSwapExchangeInfo(ret types.ExchangeInfoApiResult) []types.ExchangeInfoSymbolApiResult {
	for i := range ret.Symbols {
		ret.Symbols[i].StdSymbol = fmt.Sprintf("%s/%s", ret.Symbols[i].BaseAsset, ret.Symbols[i].QuoteAsset)
	}
	return ret.Symbols
}

func FormatterCoinSwapExchangeInfo(ret types.ExchangeInfoApiResult) []types.ExchangeInfoSymbolApiResult {
	var retSymbolList []types.ExchangeInfoSymbolApiResult = make([]types.ExchangeInfoSymbolApiResult, 0)
	for i := range ret.Symbols {
		if ret.Symbols[i].ContractType == "PERPETUAL" { // temporarily retain only perpetual contracts
			ret.Symbols[i].StdSymbol = fmt.Sprintf("%s/%s", ret.Symbols[i].BaseAsset, ret.Symbols[i].QuoteAsset)
			retSymbolList = append(retSymbolList, ret.Symbols[i])
		} else {
			// logger.CSwapMarket.Warnf("ignore settlement contracts, orPrep:%s", ret.Symbols[i].Symbol)
		}

	}
	return retSymbolList
}
