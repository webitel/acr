package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlPrivateRouteStore struct {
	SqlStore
}

func NewSqlPrivateRouteStore(sqlStore SqlStore) store.PrivateRouteStore {
	st := &SqlPrivateRouteStore{sqlStore}
	return st
}

func (self SqlPrivateRouteStore) Get(callId, domain string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var route *model.CallFlow
		err := self.GetMaster().SelectOne(&route, `WITH deadline as (
	  DELETE FROM callflow_private
	  WHERE created_on + callflow_private.deadline <= extract(EPOCH FROM now() at time zone 'utc')::INT
	), current as (
		  DELETE FROM callflow_private
			WHERE uuid = :CallId AND domain = :Domain
			RETURNING domain as domain, fs_timezone as fs_timezone, callflow as callflow,
				(select variables::JSON from callflow_variables where domain = :Domain LIMIT 1)
	)
	SELECT * from current limit 1`, map[string]interface{}{"CallId": callId, "Domain": domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = route
		}
	})
}

func (self SqlPrivateRouteStore) Create(callId, domain, timeZone string, deadline int, apps model.ArrayApplications) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := self.GetMaster().Exec(`INSERT INTO callflow_private ("uuid", "domain", "deadline", "fs_timezone", "callflow")
VALUES (:CallId, :Domain, :Deadline, :TimeZone, :CallFlow)`, map[string]interface{}{"CallId": callId, "Domain": domain,
			"Deadline": deadline, "TimeZone": timeZone, "CallFlow": apps}); err != nil {
			result.Err = err
		}
	})
}

func (self SqlPrivateRouteStore) Remove(domain, callId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		_, result.Err = self.GetMaster().Exec(`delete from callflow_private where domain = :Domain and uuid = :CallId`,
			map[string]interface{}{"Domain": domain, "CallId": callId})

	})
}
