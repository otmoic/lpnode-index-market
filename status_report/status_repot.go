package statusreport

import (
	"encoding/json"
	"fmt"
	"lp_market/market"
	"lp_market/redis_database"
	"lp_market/types"
	"os"
	"time"

	"log"
)

type Store struct {
	Lasttime          int64    `json:"lasttime"`
	SpotMarketSymbols []string `json:"spotMarketSymbols"`
}
type StatusReport struct {
	StoreData *Store
}

var marketCenter *market.MarketCenter

func NewStatusReport() *StatusReport {
	marketCenter = &market.MarketCenter{}
	sr := &StatusReport{StoreData: &Store{}}
	return sr
}
func (sr *StatusReport) UpdateStatus() {
	go func() {
		for {
			sr.StoreData.Lasttime = time.Now().UnixNano() / 1e6
			time.Sleep(time.Second * 5)
		}
	}()

	go func() {
		for {
			sr.StoreData.SpotMarketSymbols = make([]string, 0)
			list := marketCenter.GetSpotIns().GetGlobalUsedSymbolList()
			store := marketCenter.GetSpotOrderbookIns().GetOrderbook()

			list.Range(func(key, value any) bool {
				v := value.(types.ExchangeInfoSymbolApiResult)
				orderbookItem, ok := store.Load(v.StdSymbol)
				var lastUpdate int64 = 0
				if ok {
					orderbook := orderbookItem.(*types.OrderBookItem)
					lastUpdate = orderbook.Timestamp
				}
				sr.StoreData.SpotMarketSymbols = append(sr.StoreData.SpotMarketSymbols, fmt.Sprintf("%s 最后更新时间:%d", v.StdSymbol, lastUpdate))
				return true
			})
			time.Sleep(time.Second * 10)
		}
	}()
}
func (sr *StatusReport) IntervalReport() {
	for {
		sr.Save()
		time.Sleep(time.Second * 30)
	}
}
func (sr *StatusReport) Save() {
	statusKey := os.Getenv("STATUS_KEY")
	if statusKey == "" {
		log.Println("没有找到statusKey")
		return
	}
	bodyByte, err := json.Marshal(sr.StoreData)
	if err != nil {
		log.Println("Json序列化出现了问题", err.Error())
		return
	}
	log.Println("写入Status状态.....")
	_, writeErr := redis_database.GetStatusDb().Set(statusKey, string(bodyByte))
	if writeErr != nil {
		log.Println("写入Status 状态发生了错误", err)
	}
}
