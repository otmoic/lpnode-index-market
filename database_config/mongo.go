package database_config

import (
	"fmt"
	"log"
	"os"
)

type MongoDbConnectInfoItem struct {
	Url      string `bson:"mongoUrl"`
	Database string `bson:"mongoDatabase"`
	UserName string ``
	Password string
}

var MongoDataBaseConfigIns = make(map[string]MongoDbConnectInfoItem)

func InitMongoConfig() {
	prodMongoHost := os.Getenv("OBRIDGE_LPNODE_DB_MONGODB_SERVICE_HOST")
	if prodMongoHost != "" {
		log.Println("使用环境变量中的Mongodb配置")
		prodMongoPass := os.Getenv("MONGODBPASS")
		prodMongoHost := os.Getenv("OBRIDGE_LPNODE_DB_MONGODB_SERVICE_HOST")
		prodMongoPort := os.Getenv("OBRIDGE_LPNODE_DB_MONGODB_SERVICE_PORT")
		prodMongoUser := os.Getenv("MONGODB_ACCOUNT")
		prodMongoDBNameStore := os.Getenv("MONGODB_DBNAME_LP_STORE")
		url := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s", prodMongoUser, prodMongoPass, prodMongoHost, prodMongoPort, prodMongoDBNameStore, prodMongoDBNameStore)
		fmt.Println(url)
		item := MongoDbConnectInfoItem{Url: url, Database: prodMongoDBNameStore}
		MongoDataBaseConfigIns["main"] = item
		return
	}
	item := MongoDbConnectInfoItem{Url: "mongodb://root:123456@127.0.0.1:27017/lp_store?authSource=admin", Database: "lp_store"}
	MongoDataBaseConfigIns["main"] = item
}
