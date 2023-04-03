package state

import (
	"context"
	"lp_market/logger"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

var once sync.Once

type StateDb struct {
	db *leveldb.DB
}

func (sd *StateDb) Init(ctx context.Context) error {
	db, err := leveldb.OpenFile("state_file/", nil)
	if err != nil {
		return err
	}
	sd.db = db
	logger.StateDb.Debug("开始初始化数据库...")
	select {
	case <-ctx.Done(): // 监听启动器的退出 和cancel
		db.Close()
		logger.SpotMarket.Debug("....... 关闭db并退出")
	}
	return nil
}

var stateDbInstance *StateDb

// 单例返回  现货市场实例
func GetStateDbInstance() *StateDb {
	once.Do(func() {
		stateDbInstance = &StateDb{}
	})
	return stateDbInstance
}
