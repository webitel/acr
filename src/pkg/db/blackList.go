/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package db

import (
	"github.com/webitel/acr/src/pkg/config"
	"gopkg.in/mgo.v2/bson"
)

var COLLECTION_BLACK_LIST = config.Conf.Get("mongodb:blackListCollection")

func (db *DB) CheckBlackList(domainName, name, number string) (err error, count int) {
	c := db.db.C(COLLECTION_BLACK_LIST)

	count, err = c.Find(bson.M{
		"domain": domainName,
		"name":   name,
		"number": number,
	}).Count()

	return
}
