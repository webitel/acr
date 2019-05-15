package mongo_store

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
	"github.com/webitel/wlog"
	"gopkg.in/mgo.v2"
	"os"
	"time"
)

const (
	MAX_RECONNECT_COUNT = 100
)

type NoSqlSupplierOldStores struct {
	geoLocation   store.GeoLocationStore
	calendar      store.CalendarStore
	blackList     store.BlackListStore
	outboundQueue store.OutboundQueueStore
	email         store.EmailStore
	media         store.MediaStore
}

//TODO remove this store
type NoSqlSupplier struct {
	uri       string
	rCnt      int
	database  *mgo.Database
	oldStores NoSqlSupplierOldStores
}

func NewNoSqlSupplier(settings model.NoSqlSettings) NoSqlStore {
	supplier := &NoSqlSupplier{
		uri: settings.Uri,
	}

	supplier.initConnection()

	supplier.oldStores.geoLocation = NewNoSqlGeoLocationStore(supplier)
	supplier.oldStores.calendar = NewNoSqlCalendarStore(supplier)
	supplier.oldStores.blackList = NewNoSqlBlackListStore(supplier)
	supplier.oldStores.outboundQueue = NewNoSqlOutboundQueueStore(supplier)
	supplier.oldStores.email = NewNoSqlEmailStore(supplier)
	supplier.oldStores.media = NewNoSqlMediaStore(supplier)

	return supplier
}

func (self *NoSqlSupplier) GetCollection(name string) *mgo.Collection {
	self.database.Session.Refresh()
	return self.database.C(name)
}

func (self *NoSqlSupplier) initConnection() {
	self.database = self.setupConnection()

}

func (self *NoSqlSupplier) setupConnection() *mgo.Database {
	self.rCnt = 0
	for {
		if self.rCnt > MAX_RECONNECT_COUNT {
			wlog.Critical(fmt.Sprintf("max reconnect to open Mongo connection"))
			time.Sleep(time.Second)
			os.Exit(101)
		}

		session, err := mgo.Dial(self.uri)
		if err != nil {
			wlog.Critical(fmt.Sprintf("failed to open Mongo connection to err:%v", err.Error()))
			time.Sleep(time.Second)
			continue
		}

		wlog.Info(fmt.Sprintf("pinging Mongo database"))
		if err = session.Ping(); err != nil {
			wlog.Critical(fmt.Sprintf("failed to ping Mongo connection to err:%v", err.Error()))
			time.Sleep(time.Second)
			continue
		}

		return session.DB("")
	}

	return nil
}

func (self *NoSqlSupplier) OutboundQueue() store.OutboundQueueStore {
	return self.oldStores.outboundQueue
}

func (self *NoSqlSupplier) GeoLocation() store.GeoLocationStore {
	return self.oldStores.geoLocation
}

func (self *NoSqlSupplier) Calendar() store.CalendarStore {
	return self.oldStores.calendar
}

func (self *NoSqlSupplier) BlackList() store.BlackListStore {
	return self.oldStores.blackList
}

func (self *NoSqlSupplier) Email() store.EmailStore {
	return self.oldStores.email
}

func (self *NoSqlSupplier) Media() store.MediaStore {
	return self.oldStores.media
}
