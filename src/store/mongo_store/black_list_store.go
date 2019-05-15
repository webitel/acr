package mongo_store

import (
	"github.com/webitel/acr/src/store"
	"gopkg.in/mgo.v2/bson"
)

type NoSqlBlackListStore struct {
	NoSqlStore
}

func NewNoSqlBlackListStore(noSqlStore NoSqlStore) store.BlackListStore {
	return &NoSqlBlackListStore{noSqlStore}
}

func (self NoSqlBlackListStore) CountNumbers(domain, name, number string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		count, err := self.GetCollection("blackList").Find(bson.M{
			"domain": domain,
			"name":   name,
			"number": number,
		}).Count()

		if err != nil {
			result.Err = err
		} else {
			result.Data = count
		}
	})
}
