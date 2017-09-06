package db

import (
	"errors"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var COLLECTION_EXTENSION = config.Conf.Get("mongodb:extensionsCollection")
var COLLECTION_DEFAULT = config.Conf.Get("mongodb:defaultCollection")
var COLLECTION_PUBLIC = config.Conf.Get("mongodb:publicCollection")

type DB struct {
	reconnecting bool
	connected    bool
	session      *mgo.Session
	db           *mgo.Database
}

func (db *DB) observeError(err error) {
	if err != nil && err != mgo.ErrNotFound {
		db.reconnect()
	}
}

// todo mutex ?
func (db *DB) reconnect() {
	db.connected = false
	if db.reconnecting {
		return
	}

	db.reconnecting = true
	go func(_db *DB) {
		time.Sleep(time.Second)
		logger.Warning("DB: try reconnect to mongodb")
		_db.session.Refresh()
		_db.reconnecting = false
		if err := _db.session.Ping(); err != nil {
			logger.Error("DB: error %s", err.Error())
			_db.reconnect()
		} else {
			logger.Info("DB: mongodb reconnect")
		}
	}(db)
}

func (db *DB) FindExtension(destinationNumber string, domainName string, dataStructure interface{}) (err error, ok bool) {
	ok = false
	if destinationNumber == "" {
		err = errors.New("destinationNumber is empty")
		return
	}

	c := db.db.C(COLLECTION_EXTENSION)

	err = c.Find(bson.M{
		"destination_number": destinationNumber,
		"domain":             domainName,
		"disabled": bson.M{
			"$ne": true,
		},
	}).Select(bson.M{
		"_id":                1,
		"debug":              1,
		"name":               1,
		"destination_number": 1,
		"fs_timezone":        1,
		"domain":             1,
		"callflow":           1,
		"onDisconnect":       1,
		"version":            1,
	}).One(dataStructure)

	if err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		db.observeError(err)
		return
	}
	ok = true
	return
}

func (db *DB) FindDefault(domainName string, dataStructure interface{}) (err error, ok bool) {
	ok = false

	c := db.db.C(COLLECTION_DEFAULT)

	err = c.Find(bson.M{
		"domain": domainName,
		"disabled": bson.M{
			"$ne": true,
		},
	}).Sort("order").Select(bson.M{
		"_id":                1,
		"debug":              1,
		"name":               1,
		"destination_number": 1,
		"fs_timezone":        1,
		"domain":             1,
		"callflow":           1,
		"onDisconnect":       1,
		"version":            1,
	}).All(dataStructure)
	if err != nil {
		if err != mgo.ErrNotFound {
			err = nil
		}
		db.observeError(err)
		return
	}
	ok = true
	return
}

func (db *DB) FindPublic(destinationNumber string, dataStructure interface{}) (err error, ok bool) {
	if destinationNumber == "" {
		err = errors.New("destination_number is undefined")
	}

	c := db.db.C(COLLECTION_PUBLIC)

	err = c.Find(bson.M{
		"destination_number": destinationNumber,
		"disabled": bson.M{
			"$ne": true,
		},
	}).Sort("version").Select(bson.M{
		"_id":                1,
		"debug":              1,
		"name":               1,
		"destination_number": 1,
		"fs_timezone":        1,
		"domain":             1,
		"callflow":           1,
		"onDisconnect":       1,
		"version":            1,
	}).One(dataStructure)

	if err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		db.observeError(err)
		return
	}
	ok = true
	return
}

//TODO RECONNECT!!!
func NewDB(uri string) *DB {
	session, err := mgo.Dial(config.Conf.Get("mongodb:uri"))
	if err != nil {
		logger.Error("Connect to %v mongo error: %v", config.Conf.Get("mongodb:uri"), err.Error())
		return NewDB(uri)
	}
	logger.Debug("Connect to mongo success")
	return &DB{
		session: session,
		db:      session.DB("webitel"),
	}
}
