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

type LPSystemNotice struct {
	eventStr string
}

type RedisBus struct {
	redisDB   *redis_database.RedisDb
	EventList chan *LPSystemNotice
}

func GetRedisBus() *RedisBus {
	redisBusOnce.Do(func() {
		redisBusIns = &RedisBus{}
		redisDB := redis_database.NewRedis("main")
		redisBusIns.redisDB = redisDB
		redisBusIns.EventList = make(chan *LPSystemNotice, 100) // Limit to 100 Event events at most
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
	log.Println("subscribing to", redisKey)
	psc.Subscribe(redisKey)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			//fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			log.Println("data received from redis: 💹", v.Channel, string(v.Data))
			rb.EventList <- &LPSystemNotice{eventStr: string(v.Data)}
			log.Println("successfully written to channel")
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Println("redis subscription encountered an interruption", v, "reconnecting in 3 seconds...")
			time.Sleep(time.Second * 3)
			go rb.SubEvent() // restart subscription event process
			return
		}
	}
}
