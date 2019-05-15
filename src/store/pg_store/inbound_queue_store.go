package pg_store

import (
	"github.com/webitel/acr/src/store"
)

type SqlInboundQueueStore struct {
	SqlStore
}

func NewSqlInboundQueueStore(sqlStore SqlStore) store.InboundQueueStore {
	st := &SqlInboundQueueStore{sqlStore}
	return st
}

func (self SqlInboundQueueStore) Exists(domain, name string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		v, err := self.GetReplica().
			SelectNullInt(`SELECT 1 FROM tiers WHERE queue = :Queue LIMIT 1`, map[string]interface{}{"Queue": name + "@" + domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = v.Int64 > 0
		}
	})
}

func (self SqlInboundQueueStore) CountAvailableAgent(domain, name string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		v, err := self.GetReplica().
			SelectNullInt(`SELECT count(*) as count
FROM "tiers"
       INNER JOIN agents a ON a.name = tiers.agent
WHERE (tiers.queue = :Queue AND a.state = 'Waiting' AND a.status in ('Available', 'Available (On Demand)')
  AND a.ready_time <= extract(epoch from now() at time zone 'utc')::BIGINT AND
       a.last_bridge_end < extract(epoch from now() at time zone 'utc')::BIGINT - a.wrap_up_time)`,
				map[string]interface{}{"Queue": name + "@" + domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = int(v.Int64)
		}
	})
}

func (self SqlInboundQueueStore) CountAvailableMembers(domain, name string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		v, err := self.GetReplica().
			SelectNullInt(`SELECT count(*) as count
FROM "members"
WHERE queue = :Queue AND state in ('Trying', 'Waiting')`,
				map[string]interface{}{"Queue": name + "@" + domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = int(v.Int64)
		}
	})
}
