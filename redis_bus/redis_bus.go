package redisbus

import (
	"fmt"
	"log"
	"lp_market/redis_database"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/tidwall/gjson"
)

var redisBusOnce sync.Once

var redisBusIns *RedisBus

type LPSystemNoticeEventItem struct {
	Str string
}
type RedisBus struct {
	redisDB   *redis_database.RedisDb
	EventList chan *LPSystemNoticeEventItem
}

func GetRedisBus() *RedisBus {
	redisBusOnce.Do(func() {
		redisBusIns = &RedisBus{}
		redisDB := redis_database.NewRedis("main")
		redisBusIns.redisDB = redisDB
		redisBusIns.EventList = make(chan *LPSystemNoticeEventItem, 100)
	})
	return redisBusIns
}

func (rb *RedisBus) PublishEvent(channel string, val string) {
	go func() {
		event := gjson.Get(val, "type").String()
		payload := gjson.Get(val, "payload").Raw
		log.Println("publishing subscription event", channel, event, payload)
		rb.redisDB.Publish(channel, val)
	}()
}
func (rb *RedisBus) SubEvent() {

	redisConn := redis_database.GetDataRedis().PoolPtr.Get()
	psc := redis.PubSubConn{Conn: redisConn}
	redisKey := "LP_SYSTEM_Notice"
	log.Println("starting subscription", redisKey)
	psc.Subscribe(redisKey)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			//fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			log.Println("retrieved data from redis:", v.Channel, string(v.Data))
			rb.EventList <- &LPSystemNoticeEventItem{Str: string(v.Data)}
			log.Println("successfully written to queue...")

		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Println("redis 	subscription was interrupted", v, "reconnecting in three seconds....")
			time.Sleep(time.Second * 3)
			go rb.SubEvent() // restarting subscription information process
			return
		}
	}
}
