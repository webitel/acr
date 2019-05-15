package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlRouteVariablesStore struct {
	SqlStore
}

func NewSqlRouteVariablesStore(sqlStore SqlStore) store.RouteVariablesStore {
	st := &SqlRouteVariablesStore{sqlStore}
	return st
}

func (self SqlRouteVariablesStore) Get(domain string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var variables *model.DomainVariables
		err := self.GetReplica().SelectOne(&variables, `select id, variables
from callflow_variables
where domain = '10.10.10.144'`, map[string]interface{}{"Domain": domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = variables
		}
	})
}

func (self SqlRouteVariablesStore) Set(domain, key, value string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		_, err := self.GetMaster().Exec(`with upsert as (
		  update callflow_variables
		  set variables = jsonb_set(variables, ARRAY[:Key], :Value, TRUE )
		  where domain = :Domain
		  returning *
		)
		INSERT INTO callflow_variables (domain, variables)
		select :Domain, jsonb_set('{}', ARRAY[:Key], :Value, TRUE )
		WHERE NOT EXISTS (SELECT * FROM upsert)`, map[string]interface{}{"Domain": domain, "Key": key, "Value": `"` + value + `"`})

		if err != nil {
			result.Err = err
		}
	})
}
