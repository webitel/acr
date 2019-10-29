package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlCallbackQueueStore struct {
	SqlStore
}

func NewSqlCallbackQueueStore(sqlStore SqlStore) store.CallbackQueueStore {
	st := &SqlCallbackQueueStore{sqlStore}
	return st
}

func (self SqlCallbackQueueStore) Exists(domain, name string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		v, err := self.GetReplica().
			SelectNullInt(`SELECT 1 FROM callback_queue WHERE domain = :Domain and name like :Queue LIMIT 1`,
				map[string]interface{}{"Queue": name, "Domain": domain})

		if err != nil {
			result.Err = err
		} else {
			result.Data = v.Int64 > 0
		}
	})
}

func (s SqlCallbackQueueStore) ExistsMember(domain, queueName string, r *model.ExistsCallbackMemberRequest) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		v, err := s.GetReplica().SelectNullInt(`select 1
from callback_queue q
where q.name = :QueueName
  and q.domain = :DomainName
  and exists(
        select 1
        from callback_members m
        where m.queue_id = q.id
          and (m.number = coalesce(:Number, m.number) and m.done = coalesce(:Done, m.done))
  )`, map[string]interface{}{
			"QueueName":  queueName,
			"DomainName": domain,
			"Number":     r.Number,
			"Done":       r.Done,
		})

		if err != nil {
			result.Err = err
		} else {
			result.Data = v.Int64 > 0
		}
	})
}

func (self SqlCallbackQueueStore) CreateMember(domain, queue, number, widgetName string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var id int64
		err := self.GetMaster().SelectOne(&id, `INSERT INTO callback_members (domain, queue_id, number, widget_id)
VALUES (:Domain, (SELECT id FROM callback_queue WHERE name = :Queue AND domain = :Domain LIMIT 1), :Number, (SELECT id
																					 FROM widget WHERE name = :Widget AND domain = :Domain LIMIT 1) )
returning id`, map[string]interface{}{"Domain": domain, "Queue": queue, "Number": number, "Widget": widgetName})

		if err != nil {
			result.Err = err
		} else {
			result.Data = id
		}
	})
}

func (self SqlCallbackQueueStore) CreateMemberComment(memberId int64, domain, createdBy, text string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var id int
		err := self.GetMaster().SelectOne(&id, `insert into callback_members_comment (member_id, created_by, text, created_on) 
select :MemberId, :CreatedBy, :Text, :CreatedOn
where exists (
 select *
 from callback_members
   inner join callback_queue c2 on callback_members.queue_id = c2.id
 where callback_members.id = :MemberId AND c2.domain = :Domain
) returning id`, map[string]interface{}{"MemberId": memberId, "Domain": domain, "CreatedBy": createdBy, "Text": text, "CreatedOn": model.GetMillis()})

		if err != nil {
			result.Err = err
		} else {
			result.Data = id
		}

	})
}
