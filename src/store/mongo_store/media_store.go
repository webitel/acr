package mongo_store

import (
	"github.com/webitel/acr/src/store"
	"gopkg.in/mgo.v2/bson"
)

type NoSqlMediaStore struct {
	NoSqlStore
}

func NewNoSqlMediaStore(noSqlStore NoSqlStore) store.MediaStore {
	return &NoSqlMediaStore{noSqlStore}
}

func (self NoSqlMediaStore) ExistsFile(name, typeFile, domain string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if typeFile == "" {
			typeFile = "mp3"
		}

		count, err := self.GetCollection("mediaFile").Find(bson.M{
			"name":   name,
			"type":   typeFile,
			"domain": domain,
		}).Count()

		if err != nil {
			result.Err = err
		} else {
			result.Data = count > 0
		}
	})
}
