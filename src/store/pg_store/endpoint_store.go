package pg_store

import "github.com/webitel/acr/src/store"

type SqlEndpointStore struct {
	SqlStore
}

func NewSqlEndpointStore(sqlStore SqlStore) store.EndpointStore {
	st := &SqlEndpointStore{sqlStore}
	return st
}
