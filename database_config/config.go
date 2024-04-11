package database_config

import (
	"fmt"
	"log"
	"os"
)

type RedisDbConnectInfoItem struct {
	DbIndex  int
	RedisUrl string
	RedisPwd string
}
type RedisDataDataBaseConfig map[string]RedisDbConnectInfoItem

var RedisDataDataBaseConfigIns = make(map[string]RedisDbConnectInfoItem)

func Init() {
	InitMongoConfig()
	InitRedisConfig()
}

func InitRedisConfig() {
	prodRedisHost := os.Getenv("OBRIDGE_LPNODE_DB_REDIS_MASTER_SERVICE_HOST")
	if prodRedisHost != "" {
		log.Println("using redis configuration from environment variables")
		prodRedisPort := 6379
		prodRedisPass := os.Getenv("REDIS_PASSWORD")
		RedisDataDataBaseConfigIns["main"] = RedisDbConnectInfoItem{
			RedisUrl: fmt.Sprintf("%s:%d", prodRedisHost, prodRedisPort),
			RedisPwd: prodRedisPass,
			DbIndex:  0,
		}
		RedisDataDataBaseConfigIns["statusDb"] = RedisDbConnectInfoItem{
			RedisUrl: fmt.Sprintf("%s:%d", prodRedisHost, prodRedisPort),
			RedisPwd: prodRedisPass,
			DbIndex:  0,
		}
		return
	}
	redisPass := os.Getenv("REDIS_PASSWORD")
	RedisDataDataBaseConfigIns["main"] = RedisDbConnectInfoItem{
		RedisUrl: "127.0.0.1:6379",
		RedisPwd: redisPass,
		DbIndex:  0,
	}
	RedisDataDataBaseConfigIns["statusDb"] = RedisDbConnectInfoItem{
		RedisUrl: "127.0.0.1:6379",
		RedisPwd: redisPass,
		DbIndex:  0,
	}
}
