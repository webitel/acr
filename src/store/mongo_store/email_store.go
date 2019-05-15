package mongo_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
	"gopkg.in/mgo.v2/bson"
)

type NoSqlEmailStore struct {
	NoSqlStore
}

func NewNoSqlEmailStore(noSqlStore NoSqlStore) store.EmailStore {
	return &NoSqlEmailStore{noSqlStore}
}

func (self NoSqlEmailStore) Config(domain string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var config *model.EmailConfig

		err := self.GetCollection("emailConfig").Pipe([]bson.M{
			{
				"$match": bson.M{
					"domain": domain,
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
		}).One(&config)

		if err != nil {
			result.Err = err
		} else {
			result.Data = config
		}
	})
}
