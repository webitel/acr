package db

import (
	"encoding/json"
	"fmt"
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
	}).Limit(1).Count()

	db.observeError(err)

	return count > 0
}

type memberRequest struct {
	Name           string      `json:"name"`
	EndCause       interface{} `json:"end_cause"`
	Communications struct {
		Number      string      `json:"number"`
		Type        interface{} `json:"type"`
		State       interface{} `json:"state"`
		Description string      `json:"description"`
	} `json:"communications"`
	Variables map[string]interface{} `json:"variables"`
}

func (db *DB) ExistsMemberInDialer(dialer, domain string, data []byte) bool {
	c := db.db.C(COLLECTION_MEMBERS)

	filter := bson.M{
		"dialer": dialer,
		"domain": domain,
	}
	var r memberRequest
	var err error

	if err = json.Unmarshal(data, &r); err != nil {
		return false
	}

	if r.Name != "" {
		filter["name"] = r.Name
	}

	if r.EndCause != nil {
		filter["_endCause"] = r.EndCause
	}

	if r.Communications.Type != nil {
		filter["communications.type"] = r.Communications.Type
	}

	if r.Communications.Number != "" {
		filter["communications.number"] = r.Communications.Number
	}

	if r.Communications.Description != "" {
		filter["communications.description"] = r.Communications.Description
	}

	if r.Communications.State != nil {
		filter["communications.state"] = r.Communications.State
	}

	for k, v := range r.Variables {
		filter[fmt.Sprintf("variables.%s", k)] = v
	}

	var count int
	count, err = c.Find(filter).Limit(1).Count()

	db.observeError(err)

	return count > 0
}

func (db *DB) ExistsQueue(name, domain string) bool {
	var count int
	db.pg.Raw(`
		SELECT 1
		WHERE EXISTS(SELECT 1 FROM tiers WHERE queue = $1)
	`, name+"@"+domain).Count(&count)

	return count > 0
}
