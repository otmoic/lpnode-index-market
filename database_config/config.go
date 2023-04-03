package database_config

import (
	"fmt"
	"log"
	"os"
)

// RedisDbConnectInfoItem Redis db的主要配置文件
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

// InitRedisConfig 初始化redis需要的配置
func InitRedisConfig() {
	prodRedisHost := os.Getenv("OBRIDGE_LPNODE_DB_REDIS_MASTER_SERVICE_HOST")
	if prodRedisHost != "" {
		log.Println("使用环境变量中的Redis配置")
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
			DbIndex:  9,
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
		DbIndex:  9,
	}
}
