/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package db

import (
	"github.com/webitel/acr/src/pkg/config"
	"gopkg.in/mgo.v2/bson"
)

var COLLECTION_CALENDAR = config.Conf.Get("mongodb:calendarCollection")

func (db *DB) GetCalendar(name, domainName string, dataStructure interface{}) error {
	c := db.db.C(COLLECTION_CALENDAR)
	return db.observeError(c.Find(bson.M{
		"name":   name,
		"domain": domainName,
	}).One(dataStructure))
}
