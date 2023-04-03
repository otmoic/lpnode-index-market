package redis_database

func GetDataRedis() *RedisDb {
	return GetRedisDbIns("main")
}
func GetStatusDb() *RedisDb {
	return GetRedisDbIns("statusDb")
}
