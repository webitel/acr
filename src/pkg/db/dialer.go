/**
 * Created by I. Navrotskyj on 31.08.17.
 */

package db

import (
	"github.com/pkg/errors"
	"github.com/webitel/acr/src/pkg/config"
	"gopkg.in/mgo.v2/bson"
)

var COLLECTION_DIALER = config.Conf.Get("mongodb:dialerCollection")
var COLLECTION_MEMBERS = config.Conf.Get("mongodb:membersCollection")
var errDialerContextBadDialerId = errors.New("Bad dialer objectId")

func (db *DB) FindDialerCallFlow(id, domainName string, dataStructure interface{}) error {

	if !bson.IsObjectIdHex(id) {
		return errDialerContextBadDialerId
	}

	c := db.db.C(COLLECTION_DIALER)
	return c.Find(bson.M{
		"_id":    bson.ObjectIdHex(id),
		"domain": domainName,
	}).Select(bson.M{"_cf": 1, "amd": 1}).One(dataStructure)
}

func (db *DB) AddMember(data interface{}) error {
	c := db.db.C(COLLECTION_MEMBERS)
	return c.Insert(data)
}

func (db *DB) UpdateMember(id string, data interface{}) error {

	if !bson.IsObjectIdHex(id) {
		return errors.New("Bad objectId " + id)
	}

	c := db.db.C(COLLECTION_MEMBERS)
	return c.UpdateId(bson.ObjectIdHex(id), bson.M{
		"$set": data,
	})
}
