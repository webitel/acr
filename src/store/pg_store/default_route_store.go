package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlDefaultRouteStore struct {
	SqlStore
}

func NewSqlDefaultRouteStore(sqlStore SqlStore) store.DefaultRouteStore {
	st := &SqlDefaultRouteStore{sqlStore}
	return st
}

func (self SqlDefaultRouteStore) Get(domain, destination string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var cf *model.CallFlow
		err := self.GetReplica().SelectOne(&cf, `SELECT 
       callflow_default.id                                        as id,
       callflow_default.destination_number                        as destination_number,
       callflow_default.name                                      as name,
       callflow_default.debug                                     as debug,
       callflow_default.domain                                    as domain,
       callflow_default.fs_timezone                               as fs_timezone,
       callflow_default.callflow                                  as callflow,
       callflow_default.callflow_on_disconnect                    as callflow_on_disconnect,
       callflow_default.version                                   as version,
       cv.variables::JSON                                         as variables
FROM "callflow_default"
       LEFT JOIN callflow_variables cv on cv.domain = callflow_default.domain
WHERE callflow_default.domain = :Domain AND callflow_default.disabled IS NOT TRUE AND  :Destination ~ callflow_default.destination_number 
ORDER BY callflow_default."order" ASC
LIMIT 1`, map[string]interface{}{"Destination": destination, "Domain": domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = cf
		}
	})
}
