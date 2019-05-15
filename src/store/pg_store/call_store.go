package pg_store

import (
	"github.com/webitel/acr/src/store"
)

type SqlCallStore struct {
	SqlStore
}

func NewSqlCallStore(sqlStore SqlStore) store.CallStore {
	st := &SqlCallStore{sqlStore}
	return st
}

func (self SqlCallStore) GetIdByPresence(presence string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var id string
		err := self.GetReplica().SelectOne(&id, `select uuid
from channels
where presence_id = :Presence
order by created desc
limit 1`, map[string]interface{}{"Presence": presence})

		if err != nil {
			result.Err = err
		} else {
			result.Data = id
		}
	})
}
