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
		redisBusIns.EventList = make(chan *LPSystemNotice, 100) // 最多Limit 100个Event事件
	})
	return redisBusIns
}

func (rb *RedisBus) PublishEvent(channel string, val string) {
	go func() {
		event := gjson.Get(val, "type").String()
		payload := gjson.Get(val, "payload").Raw
		log.Println("发布订阅事件", channel, event, payload)
		rb.redisDB.Publish(channel, val)
	}()
}

func (rb *RedisBus) SubEvent() {
	redisConn := redis_database.GetDataRedis().PoolPtr.Get()
	psc := redis.PubSubConn{Conn: redisConn}
	redisKey := "LP_SYSTEM_Notice"
	log.Println("开始订阅", redisKey)
	psc.Subscribe(redisKey)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			//fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			log.Println("从redis中获取到数据:💹", v.Channel, string(v.Data))
			rb.EventList <- &LPSystemNotice{eventStr: string(v.Data)}
			log.Println("写入到Chann中成功")
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Println("redis 订阅发生了中断", v, "三秒后重新链接....")
			time.Sleep(time.Second * 3)
			go rb.SubEvent() // 重新运行订阅信息进程
			return
		}
	}
}
