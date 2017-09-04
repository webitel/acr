/**
 * Created by I. Navrotskyj on 31.08.17.
 */

package db

import (
	"github.com/webitel/acr/src/pkg/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var COLLECTION_DOMAIN_VARIABLES = config.Conf.Get("mongodb:variablesCollection")

func (db *DB) GetDomainVariables(domainName string, dataStructure interface{}) (err error) {

	err = db.db.C(COLLECTION_DOMAIN_VARIABLES).Find(bson.M{
		"domain": domainName,
	}).Select(bson.M{"variables": 1, "_id": 0}).One(dataStructure)

	if err == mgo.ErrNotFound {
		err = nil
	}

	return
}

func (db *DB) SetDomainVariable(domainName, key, value string) error {
	_, e := db.db.C(COLLECTION_DOMAIN_VARIABLES).
		Upsert(bson.M{"domain": domainName}, bson.M{"$set": bson.M{"variables." + key: value}})
	return e
}
