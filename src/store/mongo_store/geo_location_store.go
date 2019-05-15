package mongo_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
	"gopkg.in/mgo.v2/bson"
)

type NoSqlGeoLocationStore struct {
	NoSqlStore
}

func NewNoSqlGeoLocationStore(noSqlStore NoSqlStore) store.GeoLocationStore {
	return &NoSqlGeoLocationStore{noSqlStore}
}

func (self NoSqlGeoLocationStore) Find(sysLength int, numbers []string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var location *model.GeoLocation

		err := self.GetCollection("location").Pipe([]bson.M{
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
		}).One(&location)

		if err != nil {
			result.Err = err
		} else {
			result.Data = location
		}
	})
}
