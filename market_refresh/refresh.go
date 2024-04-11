package marketrefresh

import (
	"context"
	"fmt"
	"log"
	"lp_market/logger"
	"lp_market/market"
	database "lp_market/mongo_database"
	redisbus "lp_market/redis_bus"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
)

var marketCenter *market.MarketCenter

func Init() {
	marketCenter = &market.MarketCenter{}

}
func RefreshSpot() {
	for {
		select {
		case v := <-redisbus.GetRedisBus().EventList:
			logger.MainMessage.Warn("value to be processed", v)
			if v != nil {
				eventName := gjson.Get(v.Str, "type")
				logger.MainMessage.Warn("	eventName to be handled:", eventName)
			}
			err := DoRefreshSpot()
			if err != nil {
				logger.MainMessage.Error("an error occurred while refreshing the spot market", err)
			}
			time.Sleep(time.Second * 10)
		}
	}

}
func DoRefreshSpot() (err error) {
	readyList := make(map[string]bool, 0)
	var stdSymbolList []string = make([]string, 0)
	var results []struct {
		TokenAddress string `bson:"tokenAddressStr"`
		MarketName   string `bson:"marketName"`
	}
	mdb, err := database.GetSession("main")
	if err != nil {
		return
	}
	err, chainCursor := database.FindAll("main", "chainList", bson.M{})
	if err != nil {
		err = errors.WithMessage(err, "an error occurred while loading native currency pairs for chains from the database")
		return
	}
	var chainResults []struct {
		TokenName string `bson:"tokenName"`
	}
	if err = chainCursor.All(context.TODO(), &chainResults); err != nil {
		err = errors.WithMessage(err, "cursor processing error ,chainList")
		return
	}
	for _, result := range chainResults {
		chainCursor.Decode(&result)
		stdSymbol := fmt.Sprintf("%s/USDT", result.TokenName)
		readyList[stdSymbol] = true
		stdSymbolList = append(stdSymbolList, stdSymbol)
		logger.MainMessage.Warnf("adding currency pair to ChainList %s", stdSymbol)
	}

	matchFilter := bson.M{"$match": bson.M{"coinType": "coin"}}
	cursor, err := mdb.Collection("tokens").Aggregate(context.TODO(), bson.A{
		matchFilter,
		bson.M{
			"$group": bson.M{
				"_id": "$marketName",
				"tokenAddress": bson.M{
					"$addToSet": "$marketName",
				},
				"tokenAddressStr": bson.M{"$first": "$$ROOT.address"},
				"marketName":      bson.M{"$first": "$$ROOT.marketName"},
			},
		},
	})
	if err != nil {
		log.Println(errors.WithMessage(err, "an error occurred while executing Aggregate."))
		return
	}
	if err = cursor.All(context.TODO(), &results); err != nil {
		err = errors.WithMessage(err, "cursor processing error")
		return
	}
	for _, result := range results {
		cursor.Decode(&result)
	}
	for _, v := range results {
		stdSymbol := fmt.Sprintf("%s/USDT", v.MarketName)
		logger.MainMessage.Warnf("add Aggregate Stdsymbol %s", stdSymbol)
		_, exist := readyList[stdSymbol]
		if exist {
			logger.MainMessage.Warnf("existing currency pair [%s] skip", stdSymbol)
			continue
		}
		stdSymbolList = append(stdSymbolList, stdSymbol)
	}

	logger.MainMessage.Warn("starting to refresh spot currency pair list", stdSymbolList)
	marketCenter.GetSpotIns().SetUsedSymbol(stdSymbolList)
	marketCenter.GetSpotIns().RefreshMarket()
	return
}
