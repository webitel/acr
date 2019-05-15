package mongo_store

import (
	"github.com/webitel/acr/src/store"
	"gopkg.in/mgo.v2"
)

type NoSqlStore interface {
	GetCollection(name string) *mgo.Collection

	GeoLocation() store.GeoLocationStore
	Calendar() store.CalendarStore
	BlackList() store.BlackListStore
	OutboundQueue() store.OutboundQueueStore
	Email() store.EmailStore
	Media() store.MediaStore
}
