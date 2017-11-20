/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package db

import (
	"github.com/webitel/acr/src/pkg/config"
	"gopkg.in/mgo.v2/bson"
)

var COLLECTION_EMAIL = config.Conf.Get("mongodb:emailCollection")

func (db *DB) GetEmailConfig(domainName string, dataStructure interface{}) error {
	c := db.db.C(COLLECTION_EMAIL)
	//TODO use Find
	return db.observeError(c.Pipe([]bson.M{
		{
			"$match": bson.M{
				"domain": domainName,
			},
		},
		{
			"$limit": 1,
		},
		{
			"$project": bson.M{
				"provider": "$provider",
				"from":     "$from",
				"host":     "$options.host",
				"user":     "$options.auth.user",
				"pass":     "$options.auth.pass",
				"secure":   "$options.secure",
				"port":     "$options.port",
			},
		},
	}).One(dataStructure))
}
