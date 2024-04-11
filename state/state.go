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
	logger.StateDb.Debug("initializing database...")
	select {
	case <-ctx.Done():
		db.Close()
		logger.SpotMarket.Debug("....... closing db and exiting")
	}
	return nil
}

var stateDbInstance *StateDb

func GetStateDbInstance() *StateDb {
	once.Do(func() {
		stateDbInstance = &StateDb{}
	})
	return stateDbInstance
}
