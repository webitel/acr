package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlExtensionRouteStore struct {
	SqlStore
}

func NewSqlExtensionRouteStore(sqlStore SqlStore) store.ExtensionRouteStore {
	st := &SqlExtensionRouteStore{sqlStore}
	return st
}

func (self SqlExtensionRouteStore) Get(domain, extension string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var cf *model.CallFlow
		err := self.GetReplica().SelectOne(&cf, `SELECT callflow_extension.id                     as id,
       callflow_extension.destination_number     as destination_number,
       callflow_extension.name                   as name,
       callflow_extension.callflow               as callflow,
       callflow_extension.callflow_on_disconnect as callflow_on_disconnect,
       callflow_extension.version                as version,
       callflow_extension.domain                 as domain,
       cv.variables::JSON                        as variables
FROM "callflow_extension"
       LEFT JOIN callflow_variables cv on cv.domain = callflow_extension.domain
WHERE (callflow_extension.domain = :Domain AND callflow_extension.destination_number = :Extension)
LIMIT 1`, map[string]interface{}{"Extension": extension, "Domain": domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = cf
		}
	})
}
