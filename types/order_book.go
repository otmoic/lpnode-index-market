package types

type BidOrAsk []string

// orderbook type
type OrderBookItem struct {
	StdSymbol         string     `json:"stdSymbol"`
	Symbol            string     `json:"symbol"`
	LastUpdateId      int64      `json:"lastUpdateId"`
	Timestamp         int64      `json:"timestamp"`         // spot is storage time swap is push time
	IncomingTimestamp int64      `json:"incomingTimestamp"` // saved time
	StreamName        string     `json:"stream"`
	Bids              []BidOrAsk `json:"bids"`
	Asks              []BidOrAsk `json:"asks"`
}
type StdSymbolInfo struct {
	Symbol     string `json:"symbol"`
	StdSymbol  string // used for storing formatted currency pair names
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
