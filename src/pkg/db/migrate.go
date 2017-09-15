/**
 * Created by I. Navrotskyj on 15.09.17.
 */

package db

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type execMigrate func(db *DB)

var COLLECTION_EXTENSION = config.Conf.Get("mongodb:extensionsCollection")
var COLLECTION_DEFAULT = config.Conf.Get("mongodb:defaultCollection")
var COLLECTION_PUBLIC = config.Conf.Get("mongodb:publicCollection")
var COLLECTION_VARIABLES = config.Conf.Get("mongodb:variablesCollection")

func (db *DB) migrateMongoToPg() {
	migrate(db, COLLECTION_EXTENSION, migrateExtension)
	migrate(db, COLLECTION_DEFAULT, migrateDefault)
	migrate(db, COLLECTION_PUBLIC, migratePublic)
	migrate(db, COLLECTION_VARIABLES, migrateVariables)
}

func migrate(db *DB, collName string, fn execMigrate) {
	c := db.db.C(collName)

	count, err := c.Find(bson.M{
		"migrate": bson.M{
			"$ne": true,
		},
	}).Count()

	if err != nil {
		logger.Error(err.Error())
		return
	}

	if count == 0 {
		logger.Debug("Not found migrate in %s", collName)
		return
	} else {
		logger.Debug("Start migrate %d in %s", count, collName)
	}
	fn(db)
}

func migrateDefault(db *DB) {
	c := db.db.C(COLLECTION_DEFAULT)

	iter := c.Find(bson.M{
		"migrate": bson.M{
			"$ne": true,
		},
	}).Sort("-domain", "+order").Iter()

	doc := make(map[string]interface{})
	var tmp, tmp2 []byte
	var d *gorm.DB

	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"migrate": true,
			},
		},
	}

	for iter.Next(&doc) {
		tmp, _ = json.Marshal(doc["callflow"])
		tmp2, _ = json.Marshal(doc["onDisconnect"])
		d = db.pg.Debug().Exec(`INSERT INTO callflow_default (destination_number, name, disabled, domain, fs_timezone, callflow, callflow_on_disconnect)
				VALUES(?, ?, ?, ?, ?, ?, ?)`,
			doc["destination_number"], doc["name"], doc["disabled"], doc["domain"], doc["fs_timezone"], tmp, tmp2)

		if d.Error != nil {
			logger.Error(`Migrate error %s in %v`, d.Error.Error(), doc)
			continue
		}

		c.FindId(doc["_id"]).Apply(change, nil)
	}
	iter.Close()
	logger.Debug("End migrate in %s", COLLECTION_DEFAULT)
}

func migratePublic(db *DB) {
	c := db.db.C(COLLECTION_PUBLIC)

	iter := c.Find(bson.M{
		"migrate": bson.M{
			"$ne": true,
		},
	}).Sort("-domain").Iter()

	doc := make(map[string]interface{})
	var tmp, tmp2 []byte
	var d *gorm.DB

	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"migrate": true,
			},
		},
	}

	for iter.Next(&doc) {
		tmp, _ = json.Marshal(doc["callflow"])
		tmp2, _ = json.Marshal(doc["onDisconnect"])

		d = db.pg.Debug().Exec(`INSERT INTO callflow_public (destination_number, name, domain, fs_timezone, disabled, callflow, callflow_on_disconnect)
				VALUES(ARRAY[?], ?, ?, ?, ?, ?::JSON, ?::JSON)`,
			doc["destination_number"], doc["name"], doc["domain"], doc["fs_timezone"], doc["disabled"], tmp, tmp2)

		if d.Error != nil {
			logger.Error(`Migrate error %s in %v`, d.Error.Error(), doc)
			continue
		}

		c.FindId(doc["_id"]).Apply(change, nil)
	}
	iter.Close()
	logger.Debug("End migrate in %s", COLLECTION_DEFAULT)
}

func migrateExtension(db *DB) {
	c := db.db.C(COLLECTION_EXTENSION)

	iter := c.Find(bson.M{
		"migrate": bson.M{
			"$ne": true,
		},
	}).Iter()

	doc := make(map[string]interface{})
	var tmp, tmp2 []byte
	var d *gorm.DB

	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"migrate": true,
			},
		},
	}

	for iter.Next(&doc) {
		tmp, _ = json.Marshal(doc["callflow"])
		tmp2, _ = json.Marshal(doc["onDisconnect"])
		d = db.pg.Debug().Exec(`INSERT INTO callflow_extension (destination_number, domain, user_id, name, callflow, callflow_on_disconnect, fs_timezone)
				select ?, ?, ?, ?, ?, ?, ?
				WHERE NOT EXISTS (select id from callflow_extension where domain = ? AND user_id = ?)`,
			doc["destination_number"], doc["domain"], doc["userRef"], doc["name"], tmp, tmp2, doc["fs_timezone"], doc["domain"], doc["userRef"])

		if d.Error != nil {
			logger.Error(`Migrate error %s in %v`, d.Error.Error(), doc)
			continue
		}

		c.FindId(doc["_id"]).Apply(change, nil)
	}
	iter.Close()
	logger.Debug("End migrate in %s", COLLECTION_EXTENSION)
}

func migrateVariables(db *DB) {
	c := db.db.C(COLLECTION_VARIABLES)

	iter := c.Find(bson.M{
		"migrate": bson.M{
			"$ne": true,
		},
	}).Iter()

	doc := make(map[string]interface{})
	var tmp []byte
	var d *gorm.DB

	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"migrate": true,
			},
		},
	}

	for iter.Next(&doc) {
		tmp, _ = json.Marshal(doc["variables"])
		d = db.pg.Debug().Exec(`INSERT INTO callflow_variables (domain, variables)
				VALUES(?, ?)`,
			doc["domain"], tmp)

		if d.Error != nil {
			logger.Error(`Migrate error %s in %v`, d.Error.Error(), doc)
			continue
		}

		c.FindId(doc["_id"]).Apply(change, nil)
	}
	iter.Close()
	logger.Debug("End migrate in %s", COLLECTION_VARIABLES)
}
