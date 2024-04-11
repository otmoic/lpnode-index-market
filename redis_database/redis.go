package redis_database

import (
	"log"
	"lp_market/database_config"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisDb struct {
	PoolPtr *redis.Pool
}

var RedisDbList map[string]*RedisDb = make(map[string]*RedisDb)

func GetRedisDbIns(key string) *RedisDb {
	redisDbIns, ok := RedisDbList[key]
	if ok {
		return redisDbIns
	}
	redisDbIns = NewRedis(key)
	RedisDbList[key] = redisDbIns
	return redisDbIns
}

func NewRedis(key string) *RedisDb {
	redisDb := &RedisDb{}
	redisConf, ok := database_config.RedisDataDataBaseConfigIns[key]
	if !ok {
		log.Printf("error retrieving system base configuration RedisKey:%s", key)
		os.Exit(0)
	}
	log.Println("link", redisConf)
	redisDb.PoolPtr = &redis.Pool{
		MaxActive:   300,
		MaxIdle:     30,
		Wait:        true,
		IdleTimeout: time.Second * 100,
		Dial: func() (redis.Conn, error) {
			setPasswd := redis.DialPassword(redisConf.RedisPwd)
			setDb := redis.DialDatabase(redisConf.DbIndex)
			conn, e := redis.Dial("tcp", redisConf.RedisUrl, setPasswd, setDb)
			return conn, e
		},
	}
	log.Println("link has ended", redisConf)
	return redisDb
}

func (redisDb *RedisDb) GetString(key string) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("GET", key)
	if err == nil {

	}
	result, err := redis.String(reply, err)
	return result, err
}
func (redisDb *RedisDb) HashGet(key string, subKey string) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("HGET", key, subKey)
	result, err := redis.String(reply, err)
	return result, err
}
func (redisDb *RedisDb) HashGetAll(key string) (map[string]string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("HGETALL", key)
	result, err := redis.Values(reply, err)
	if err != nil {
		return nil, err
	}
	index := 0
	resultMap := make(map[string]string)
	var itemKey string
	var itemvalue string
	for _, v := range result {
		index++
		if index%2 == 0 {
			itemvalue = string(v.([]byte))
			resultMap[itemKey] = itemvalue
		} else {
			itemKey = string(v.([]byte))
		}
	}
	return resultMap, err
}
func (redisDb *RedisDb) Del(key string) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("DEL", key)
	if err == nil {

	}
	result, err := redis.Int64(reply, err)
	return result, err
}
func (redisDb *RedisDb) RPush(key string, value string) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	conn.Do("LTRIM", key, 0, 10000)
	reply, err := conn.Do("RPUSH", key, value)
	if err == nil {

	}
	result, err := redis.Int64(reply, err)
	return result, err
}
func (redisDb *RedisDb) HashDel(key string, subKey string) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("HDEL", key, subKey)
	if err == nil {

	}
	result, err := redis.Int64(reply, err)
	return result, err
}
func (redisDb *RedisDb) HashSet(key string, subKey string, value string) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("HSET", key, subKey, value)
	if err == nil {

	}
	result, err := redis.Int64(reply, err)
	return result, err
}
func (redisDb *RedisDb) SetString(key string, value string) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("SET", key, value)
	result, err := redis.String(reply, err)
	return result, err
}

func (redisDb *RedisDb) SetNx(key string, time int64) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("Set", key, 1, "EX", time, "NX")
	result, err := redis.String(reply, err)
	return result, err
}

func (redisDb *RedisDb) SetEx(key string, value string, time int64) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("Set", key, value, "EX", time)
	result, err := redis.String(reply, err)
	return result, err
}

func (redisDb *RedisDb) Set(key string, value string) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("Set", key, value)
	result, err := redis.String(reply, err)
	return result, err
}

func (redisDb *RedisDb) SetExpire(key string, second int64) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("EXPIRE", key, second)
	result, err := redis.Int64(reply, err)
	return result, err
}
func (redisDb *RedisDb) SetExpireAt(key string, time int64) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("EXPIREAT", key, time)
	result, err := redis.Int64(reply, err)
	return result, err
}

func (redisDb *RedisDb) LIndex(key string, index int64) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("LINDEX", key, index)
	result, err := redis.String(reply, err)
	return result, err
}
func (redisDb *RedisDb) LPop(key string) (string, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("LPOP", key)
	result, err := redis.String(reply, err)
	return result, err
}
func (redisDb *RedisDb) Publish(key string, value string) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("PUBLISH", key, value)
	result, err := redis.Int64(reply, err)
	return result, err
}
func (redisDb *RedisDb) PublishByByte(key string, value []byte) (int64, error) {
	conn := redisDb.PoolPtr.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	reply, err := conn.Do("PUBLISH", key, value)
	result, err := redis.Int64(reply, err)
	return result, err
}
