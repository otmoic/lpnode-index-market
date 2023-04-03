package database

import (
	"context"
	"errors"

	"github.com/aws/smithy-go/ptr"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FindOneAndUpdateSession(dbKey string, collName string, filter interface{}, update interface{}) (commit *struct{ Commit func(commit bool) error }, ret *mongo.UpdateResult, err error) {
	if !IsInit(dbKey) {
		err = errors.New("数据库未初始化")
		return
	}
	client := DbList[dbKey].Client
	dbName := DbList[dbKey].DbName
	sess, err := client.StartSession()
	if err != nil {
		return
	}
	sessCtx := mongo.NewSessionContext(context.TODO(), sess)
	if err = sess.StartTransaction(); err != nil {
		return
	}
	coll := client.Database(dbName).Collection(collName)
	ret, err = coll.UpdateOne(sessCtx, filter, update, &options.UpdateOptions{Upsert: ptr.Bool(true)})
	commit = &struct {
		Commit func(commit bool) error
	}{
		Commit: func(commit bool) error {
			if commit {
				sess.CommitTransaction(context.Background())
				return nil
			}
			sess.AbortTransaction(context.Background())
			return nil
		},
	}
	return
}
