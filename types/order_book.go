package types

type BidOrAsk []string

// orderbook 类型
type OrderBookItem struct {
	StdSymbol         string     `json:"stdSymbol"`
	Symbol            string     `json:"symbol"`
	LastUpdateId      int64      `json:"lastUpdateId"`
	Timestamp         int64      `json:"timestamp"`         // spot 为存储时间 swap 为推送时间
	IncomingTimestamp int64      `json:"incomingTimestamp"` // 保存的时间
	StreamName        string     `json:"stream"`
	Bids              []BidOrAsk `json:"bids"`
	Asks              []BidOrAsk `json:"asks"`
}
type StdSymbolInfo struct {
	Symbol     string `json:"symbol"`
	StdSymbol  string // 用于存储格式化之后的币对名称
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
}
type RetOrderbookMessage struct {
	Code int64                     `json:"code"`
	Data map[string]*OrderBookItem `json:"data"`
}
type RetFundingRateItem struct {
	Code int64
	Data *StdFundingRate
}
