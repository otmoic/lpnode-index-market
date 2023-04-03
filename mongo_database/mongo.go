package database

import (
	"context"
	"fmt"
	"log"
	"lp_market/database_config"
	sysLogger "lp_market/logger"
	"lp_market/types"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/smithy-go/ptr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoSet struct {
	Client  *mongo.Client
	Session *mongo.Database
	DbName  string
}

var DbList map[string]MongoSet = make(map[string]MongoSet)

type UrlOption struct {
	Url      string
	DataBase string
}

// IsInit 判断数据库是否已经链接
func IsInit(dbKey string) bool {
	_, ok := DbList[dbKey]
	if !ok {
		return false
	}
	return true
}
func InitConnect(dbKey string, option *UrlOption) (err error) {
	if IsInit(dbKey) {
		log.Printf("数据库已经连接过Key[%s]", dbKey)
		return
	}
	var url string
	var dbName string
	if option == nil {
		mongoConfig, ok := database_config.MongoDataBaseConfigIns[dbKey]
		url = mongoConfig.Url
		dbName = mongoConfig.Database
		if !ok {
			sysLogger.Config.Errorf("Mongo的配置文件不存在,Key[%s]", dbKey)
			os.Exit(0)
		}
	} else {
		url = option.Url
		dbName = option.DataBase
	}

	log.Println("开始链接数据库", url)
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*2)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Println("链接发生了错误", err)
		return
	}
	dbSession := client.Database(dbName)
	log.Println("选择数据库", dbName)

	selectCtx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	selectCtx.Deadline()
	collections, err := dbSession.ListCollections(selectCtx, bson.M{})
	if err != nil {
		return
	}
	var _ = collections
	DbList[dbKey] = MongoSet{
		Client:  client,
		Session: dbSession,
		DbName:  dbName,
	}
	log.Println("选择数据库结束", dbName)
	return
}

func FindOne(dbKey string, collection string, filter bson.M, v interface{}) error {
	if !IsInit(dbKey) {
		return errors.New(fmt.Sprintf("数据库没有初始化%s", dbKey))
	}
	session := DbList[dbKey].Session
	// database := DbList[dbKey].DbName
	err := session.Collection(collection).FindOne(context.Background(), filter).Decode(v)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}
	return nil
}
func MatchOne(dbKey string, collection string, filter bson.M) (res bool, err error) {
	res = false
	err = nil
	if !IsInit(dbKey) {
		err = errors.WithMessage(err, fmt.Sprintf("数据库没有初始化%s", dbKey))
		return
	}
	session := DbList[dbKey].Session
	docStruct := &struct {
		ID primitive.ObjectID `bson:"_id"`
	}{}
	err = session.Collection(collection).FindOne(context.Background(), filter).Decode(docStruct)
	if err != nil {
		return
	}
	if docStruct.ID.Hex() == types.MongoEmptyIdHex {
		res = true
		return
	}
	res = true
	return
}
func FindAll(dbKey string, collection string, filter bson.M) (error, *mongo.Cursor) {
	if !IsInit(dbKey) {
		return errors.New(fmt.Sprintf("数据库没有初始化%s", dbKey)), nil
	}
	session := DbList[dbKey].Session
	cursor, err := session.Collection(collection).Find(context.Background(), filter)
	if err != nil {
		return err, nil
	}
	return nil, cursor
}
func FindOneAndUpdate(dbKey string, collection string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	if !IsInit(dbKey) {
		return nil, errors.New(fmt.Sprintf("数据库没有初始化%s", dbKey))
	}
	session := DbList[dbKey].Session
	result, err := session.Collection(collection).UpdateOne(context.Background(), filter, update, &options.UpdateOptions{Upsert: ptr.Bool(true)})
	if err != nil {
		return nil, err
	}
	return result, nil
}
func Update(dbKey string, collection string, filter interface{}, update interface{}) error {
	if !IsInit(dbKey) {
		return errors.New(fmt.Sprintf("数据库没有初始化%s", dbKey))
	}
	session := DbList[dbKey].Session
	_, err := session.Collection(collection).UpdateOne(context.Background(), filter, update, &options.UpdateOptions{Upsert: ptr.Bool(false)})
	if err != nil {
		return err
	}
	return nil
}
func Insert(dbKey string, collection string, set interface{}) error {
	if !IsInit(dbKey) {
		return errors.New(fmt.Sprintf("数据库没有初始化%s", dbKey))
	}
	session := DbList[dbKey].Session
	session.Collection(collection).InsertOne(context.Background(), set)
	return nil
}
func DeleteOne(dbKey string, collection string, filter interface{}) (int64, error) {
	if !IsInit(dbKey) {
		return 0, errors.New(fmt.Sprintf("数据库没有初始化%s", dbKey))
	}
	session := DbList[dbKey].Session
	deleteCount, err := session.Collection(collection).DeleteOne(context.TODO(), filter)
	if err != nil {
		return 0, err
	}
	return deleteCount.DeletedCount, nil
}

func GetSession(dbKey string) (*mongo.Database, error) {
	if !IsInit(dbKey) {
		return nil, errors.New(fmt.Sprintf("数据库没有初始化%s", dbKey))
	}
	session := DbList[dbKey].Session
	return session, nil
}
