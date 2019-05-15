package mongo_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
	"gopkg.in/mgo.v2/bson"
)

type NoSqlCalendarStore struct {
	NoSqlStore
}

func NewNoSqlCalendarStore(noSqlStore NoSqlStore) store.CalendarStore {
	return &NoSqlCalendarStore{noSqlStore}
}

func (self NoSqlCalendarStore) Get(domain, name string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var calendar *model.Calendar

		err := self.GetCollection("calendar").Find(bson.M{
			"name":   name,
			"domain": domain,
		}).One(&calendar)

		if err != nil {
			result.Err = err
		} else {
			result.Data = calendar
		}
	})
}
