package binance

import "lp_market/logger"

// 币安现货配置

/*
本篇所列出的所有wss接口的baseurl为: wss://stream.binance.com:9443 或者 wss://stream.binance.com:443
Streams有单一原始 stream 或组合 stream
单一原始 streams 格式为 /ws/<streamName>
组合streams的URL格式为 /stream?streams=<streamName1>/<streamName2>/<streamName3>
订阅组合streams时，事件payload会以这样的格式封装: {"stream":"<streamName>","data":<rawPayload>}
stream名称中所有交易对均为 小写
每个到 stream.binance.com 的链接有效期不超过24小时，请妥善处理断线重连。
每3分钟，服务端会发送ping帧，客户端应当在10分钟内回复pong帧，否则服务端会主动断开链接。允许客户端发送不成对的pong帧(即客户端可以以高于10分钟每次的频率发送pong帧保持链接)。
wss://data-stream.binance.com 可以用来订阅市场信息的数据流。 用户信息无法从此URL获得。
*/

var UseSpotTestnet = false

// 现货相关配置
var SpotMarketWssBaseUrl string = "wss://stream.binance.com:9443/stream?streams="
var SpotMarketHttpsBaseUrl string = "https://api.binance.com"
var ExchangeInfoPath string = "/api/v3/exchangeInfo"
var SpotLevels = 5
var SpotOrderbookUpdateInterval = 1000 // Update Speed: 1000ms 或 100ms

func init() {
	logger.MainMessage.Debug("💢💢", "Market Var Init...")
	if UseSpotTestnet {
		SpotMarketWssBaseUrl = "wss://testnet.binance.vision/stream?streams="
		SpotMarketHttpsBaseUrl = "https://testnet.binance.vision"
	}
}

// U本位合约相关配置

var UsdtSwapLevels = 5
var UsdtSwapMarketWssBaseUrl string = "wss://fstream.binance.com/ws"
var UsdtSwapMarketHttpsBaseUrl string = "https://fapi.binance.com"
var UsdtSwapExchangeInfoPath string = "/fapi/v1/exchangeInfo"
var UsdtSwapFundingRate string = "/fapi/v1/premiumIndex" // 获取FundingRate的接口
var UsdtSwapFundingRateUpdateInterval = 10               // FundingRate更新的频率，单位Sec
var UsdtSwapOrderbookUpdateInterval = 500

// 币本位合约相关配置

var CoinSwapLevels = 5
var CoinSwapMarketWssBaseUrl string = "wss://dstream.binance.com/stream?streams="
var CoinSwapExchangeInfoPath string = "/dapi/v1/exchangeInfo"

var CoinSwapMarketHttpsBaseUrl string = "https://dapi.binance.com"
var CoinSwapFundingRatePath string = "/dapi/v1/premiumIndex"
var CoinSwapFundingRateUpdateInterval = 10 // FundingRate更新的频率，单位Sec
var CoinSwapOrderbookUpdateInterval = 500

// 系统配置相关

var ShowSpotOrderbookInfo = false
var ShowUSwapOrderbookInfo = false
var ShowCSwapOrderbookInfo = false
