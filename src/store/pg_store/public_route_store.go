package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlPublicRouteStore struct {
	SqlStore
}

func NewSqlPublicRouteStore(sqlStore SqlStore) store.PublicRouteStore {
	st := &SqlPublicRouteStore{sqlStore}
	return st
}

func (self SqlPublicRouteStore) Get(destination string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var cf *model.CallFlow
		err := self.GetReplica().SelectOne(&cf, `SELECT callflow_public.id                     as id,
       callflow_public.destination_number     as destination_number,
       callflow_public.name                   as name,
       callflow_public.debug                  as debug,
       callflow_public.domain                 as domain,
       callflow_public.fs_timezone            as fs_timezone,
       callflow_public.callflow               as callflow,
       callflow_public.callflow_on_disconnect as callflow_on_disconnect,
       callflow_public.version                as version,
       cv.variables::JSON                     as variables
FROM "callflow_public"
       LEFT JOIN callflow_variables cv on cv.domain = callflow_public.domain
WHERE (:Destination = ANY (callflow_public.destination_number) AND disabled IS NOT TRUE)
LIMIT 1`, map[string]interface{}{"Destination": destination})

		if err != nil {
			result.Err = err
		} else {
			result.Data = cf
		}
	})
}
