package types

type ExchangeInfoSymbolApiResult struct {
	Symbol       string `json:"symbol"`
	StdSymbol    string // 用于存储格式化之后的币对名称
	Status       string `json:"status"` //  swap  TRADING , 现货可以过滤一下 TRADING，否则可能不在交易
	BaseAsset    string `json:"baseAsset"`
	QuoteAsset   string `json:"quoteAsset"`
	MarginAsset  string `json:"marginAsset"`
	ContractType string `json:"contractType"` // 可以筛选出永续合约 PERPETUAL
}

// 初步看 spot u swap 返回的内容是一致的可以直接公用一个
type ExchangeInfoApiResult struct {
	Timezone   string                        `json:"timezone"`
	ServerTime int64                         `json:"serverTime"`
	Symbols    []ExchangeInfoSymbolApiResult `json:"symbols"`
}

//	{
//		"symbol": "BTCUSDT",                // 交易对
//		"markPrice": "11793.63104562",      // 标记价格
//		"indexPrice": "11781.80495970",     // 指数价格
//		"estimatedSettlePrice": "11781.16138815",  // 预估结算价,仅在交割开始前最后一小时有意义
//		"lastFundingRate": "0.00038246",    // 最近更新的资金费率
//		"nextFundingTime": 1597392000000,   // 下次资金费时间
//		"interestRate": "0.00010000",       // 标的资产基础利率
//		"time": 1597370495002               // 更新时间
//	}

type BinanceUsdtSwapFundingRate struct {
	Symbol               string `json:"symbol"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`
	EstimatedSettlePrice string `json:"estimatedSettlePrice"`
	LastFundingRate      string `json:"lastFundingRate"`
	NextFundingTime      int64  `json:"nextFundingTime"`
	InterestRate         string `json:"interestRate"`
	Time                 int64  `json:"time"`
}
type BinanceCoinSwapFundingRate struct {
	Symbol               string `json:"symbol"`
	Pair                 string `json:"pair"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`
	EstimatedSettlePrice string `json:"estimatedSettlePrice"`
	LastFundingRate      string `json:"lastFundingRate"`
	NextFundingTime      int64  `json:"nextFundingTime"`
	InterestRate         string `json:"interestRate"`
	Time                 int64  `json:"time"`
}
type BinanceUsdtSwapFundingRateApiResult []BinanceUsdtSwapFundingRate
type SpotOrderBookInDataMessage struct {
	LastUpdateId int64      `json:"lastUpdateId"`
	Bids         []BidOrAsk `json:"bids"`
	Asks         []BidOrAsk `json:"asks"`
}
type SpotOrderBookInMessage struct {
	Stream string                     `json:"stream"`
	Data   SpotOrderBookInDataMessage `json:"data"`
}

type UsdtSwapOrderBookInMessage struct {
	EventType      string     `json:"e"` // 事件类型
	EventTimestamp int64      `json:"E"` // 事件类型
	TradeTimestamp int64      `json:"T"`
	Stream         string     `json:"s"` //"s": "BTCUSDT",
	Bids           []BidOrAsk `json:"b"`
	Asks           []BidOrAsk `json:"a"`
}

type CoinSwapOrderBookInDataMessage struct {
	EventType      string     `json:"e"` // 事件类型
	EventTimestamp int64      `json:"E"` // 事件类型
	TradeTimestamp int64      `json:"T"`
	Stream         string     `json:"s"`  //"s":"BTCUSD_200626",      /
	StdStream      string     `json:"ps"` // "ps":"BTCUSD",
	Bids           []BidOrAsk `json:"b"`
	Asks           []BidOrAsk `json:"a"`
}
type CoinSwapOrderBookInMessage struct {
	Stream string                         `json:"stream"`
	Data   CoinSwapOrderBookInDataMessage `json:"data"`
}
