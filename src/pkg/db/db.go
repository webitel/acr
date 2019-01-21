package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2"
	"strconv"
	"time"
)

type DB struct {
	reconnecting bool
	connected    bool
	pg           *gorm.DB
	session      *mgo.Session
	db           *mgo.Database
}

func (db *DB) observeError(err error) error {
	if err != nil && err != mgo.ErrNotFound {
		db.reconnect()
	}
	return err
}

// todo mutex
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

func (db *DB) connectToPg() {
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s application_name=ACR.v1",
		config.Conf.Get("pg:user"), config.Conf.Get("pg:password"), config.Conf.Get("pg:dbName"), config.Conf.Get("pg:host"))

	pg, err := gorm.Open("postgres", dbInfo)
	if err != nil {
		logger.Debug("Connect to PG %s error: %s", dbInfo, err.Error())
		time.Sleep(time.Second)
		db.connectToPg()
		return
	}
	logger.Debug("Connect to PG %s - success", dbInfo)
	pg.Debug()

	if config.Conf.Get("pg:max") != "" {
		var maxConnection int
		maxConnection, _ = strconv.Atoi(config.Conf.Get("pg:max"))
		pg.DB().SetMaxOpenConns(maxConnection)
	}
	db.pg = pg
	db.migrateMongoToPg()
}

func NewDB(uri string) *DB {
	session, err := mgo.Dial(config.Conf.Get("mongodb:uri"))
	if err != nil {
		logger.Error("Connect to %v mongo error: %v", config.Conf.Get("mongodb:uri"), err.Error())
		return NewDB(uri)
	}

	logger.Debug("Connect to mongo: %s success", config.Conf.Get("mongodb:uri"))
	db := &DB{
		session: session,
		db:      session.DB(""),
	}
	db.connectToPg()
	return db
}

// GetMillis is a convience method to get milliseconds since epoch.
func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
