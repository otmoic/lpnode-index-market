package types

type ExchangeInfoSymbolApiResult struct {
	Symbol       string `json:"symbol"`
	StdSymbol    string
	Status       string `json:"status"`
	BaseAsset    string `json:"baseAsset"`
	QuoteAsset   string `json:"quoteAsset"`
	MarginAsset  string `json:"marginAsset"`
	ContractType string `json:"contractType"`
}

type ExchangeInfoApiResult struct {
	Timezone   string                        `json:"timezone"`
	ServerTime int64                         `json:"serverTime"`
	Symbols    []ExchangeInfoSymbolApiResult `json:"symbols"`
}

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
	EventType      string     `json:"e"`
	EventTimestamp int64      `json:"E"`
	TradeTimestamp int64      `json:"T"`
	Stream         string     `json:"s"` //"s": "BTCUSDT",
	Bids           []BidOrAsk `json:"b"`
	Asks           []BidOrAsk `json:"a"`
}

type CoinSwapOrderBookInDataMessage struct {
	EventType      string     `json:"e"`
	EventTimestamp int64      `json:"E"`
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
