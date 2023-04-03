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
			logger.MainMessage.Warn("需要处理的值", v)
			if v != nil {
				eventName := gjson.Get(v.Str, "type")
				logger.MainMessage.Warn("需要处理的eventName:", eventName)
			}
			err := DoRefreshSpot()
			if err != nil {
				logger.MainMessage.Error("刷新现货市场发生了错误", err)
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
		err = errors.WithMessage(err, "从数据库中加载链的原生币对发生了错误")
		return
	}
	var chainResults []struct {
		TokenName string `bson:"tokenName"`
	}
	if err = chainCursor.All(context.TODO(), &chainResults); err != nil {
		err = errors.WithMessage(err, "cursor处理错误,chainList")
		return
	}
	for _, result := range chainResults {
		chainCursor.Decode(&result)
		stdSymbol := fmt.Sprintf("%s/USDT", result.TokenName)
		readyList[stdSymbol] = true
		stdSymbolList = append(stdSymbolList, stdSymbol)
		logger.MainMessage.Warnf("增加ChainList中的币对%s", stdSymbol)
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
		log.Println(errors.WithMessage(err, "执行Aggregate 发生了错误."))
		return
	}
	if err = cursor.All(context.TODO(), &results); err != nil {
		err = errors.WithMessage(err, "cursor处理错误")
		return
	}
	for _, result := range results {
		cursor.Decode(&result)
	}
	for _, v := range results {
		stdSymbol := fmt.Sprintf("%s/USDT", v.MarketName)
		logger.MainMessage.Warnf("添加 Aggregate Stdsymbol %s", stdSymbol)
		_, exist := readyList[stdSymbol]
		if exist {
			logger.MainMessage.Warnf("已经存在的币对 [%s] 跳过", stdSymbol)
			continue
		}
		stdSymbolList = append(stdSymbolList, stdSymbol)
	}

	logger.MainMessage.Warn("开始刷新现货币对列表", stdSymbolList)
	marketCenter.GetSpotIns().SetUsedSymbol(stdSymbolList)
	marketCenter.GetSpotIns().RefreshMarket()
	return
}
