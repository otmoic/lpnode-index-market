package binance

import "lp_market/logger"

var UseSpotTestnet = false

var SpotMarketWssBaseUrl string = "wss://stream.binance.com:9443/stream?streams="
var SpotMarketHttpsBaseUrl string = "https://api.binance.com"
var ExchangeInfoPath string = "/api/v3/exchangeInfo"
var SpotLevels = 5
var SpotOrderbookUpdateInterval = 1000 // Update Speed: 1000ms  100ms

func init() {
	logger.MainMessage.Debug("💢💢", "Market Var Init...")
	if UseSpotTestnet {
		SpotMarketWssBaseUrl = "wss://testnet.binance.vision/stream?streams="
		SpotMarketHttpsBaseUrl = "https://testnet.binance.vision"
	}
}

var UsdtSwapLevels = 5
var UsdtSwapMarketWssBaseUrl string = "wss://fstream.binance.com/ws"
var UsdtSwapMarketHttpsBaseUrl string = "https://fapi.binance.com"
var UsdtSwapExchangeInfoPath string = "/fapi/v1/exchangeInfo"
var UsdtSwapFundingRate string = "/fapi/v1/premiumIndex"
var UsdtSwapFundingRateUpdateInterval = 10
var UsdtSwapOrderbookUpdateInterval = 500

var CoinSwapLevels = 5
var CoinSwapMarketWssBaseUrl string = "wss://dstream.binance.com/stream?streams="
var CoinSwapExchangeInfoPath string = "/dapi/v1/exchangeInfo"

var CoinSwapMarketHttpsBaseUrl string = "https://dapi.binance.com"
var CoinSwapFundingRatePath string = "/dapi/v1/premiumIndex"
var CoinSwapFundingRateUpdateInterval = 10
var CoinSwapOrderbookUpdateInterval = 500

var ShowSpotOrderbookInfo = false
var ShowUSwapOrderbookInfo = false
var ShowCSwapOrderbookInfo = false
