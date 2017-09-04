/**
 * Created by I. Navrotskyj on 30.08.17.
 */

package db

import (
	"github.com/webitel/acr/src/pkg/config"
	"gopkg.in/mgo.v2/bson"
)

var COLLECTION_LOCATION = config.Conf.Get("mongodb:locationNumberCollection")

func (db *DB) FindLocation(sysLength int, numbers []string, dataStructure interface{}) error {
	c := db.db.C(COLLECTION_LOCATION)
	//TODO use Find
	return c.Pipe([]bson.M{
		{
			"$match": bson.M{
				"sysLength": sysLength,
				"code": bson.M{
					"$in": numbers,
				},
			},
		},
		{
			"$sort": bson.M{
				"sysOrder": -1,
			},
		},
		{
			"$limit": 1,
		},
		{
			"$unwind": "$goecode",
		},
		{
			"$limit": 1,
		},
		{
			"$project": bson.M{
				"_id":         0,
				"latitude":    "$goecode.latitude",
				"longitude":   "$goecode.longitude",
				"countryCode": "$goecode.countryCode",
				"country":     "$country",
				"type":        "$type",
				"city":        "$city",
			},
		},
	}).One(dataStructure)
}
