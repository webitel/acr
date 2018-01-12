package db

import (
	"github.com/webitel/acr/src/pkg/config"
	"gopkg.in/mgo.v2/bson"
)

var COLLECTION_MEDIA_FILE = config.Conf.Get("mongodb:mediaFileCollection")

func (db *DB) ExistsMediaFile(name, typeFile, domainName string) bool {
	c := db.db.C(COLLECTION_MEDIA_FILE)
	if typeFile == "" {
		typeFile = "mp3"
	}
	count, err := c.Find(bson.M{
		"name":   name,
		"type":   typeFile,
		"domain": domainName,
	}).Count()

	db.observeError(err)

	return count > 0
}

func (db *DB) ExistsDialer(name, domain string) bool {
	c := db.db.C(COLLECTION_DIALER)

	or := []bson.M{
		{
			"name": name,
		},
	}
	if bson.IsObjectIdHex(name) {
		or = append(or, bson.M{
			"_id": bson.ObjectIdHex(name),
		})
	}

	count, err := c.Find(bson.M{
		"$or":    or,
		"domain": domain,
	}).Count()

	db.observeError(err)

	return count > 0
	return false
}

func (db *DB) ExistsQueue(name, domain string) bool {
	var count int
	db.pg.Raw(`
		SELECT 1
		WHERE EXISTS(SELECT 1 FROM tiers WHERE queue = $1)
	`, name+"@"+domain).Count(&count)

	return count > 0
}
