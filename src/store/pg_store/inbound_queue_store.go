package pg_store

import (
	"github.com/webitel/acr/src/model"
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

func (self SqlInboundQueueStore) DistributeMember(domainId int64, queueName string, member *model.InboundMember) store.StoreChannel {

	return store.Do(func(result *store.StoreResult) {
		var res *model.MemberAttempt
		if err := self.GetMaster().SelectOne(&res, `select q.id as queue_id, q.enabled,  case when q.enabled is true
  then (select attempt_id
     from cc_add_to_queue(:Provider::varchar(50), q.id::bigint, :CallId::varchar(36), :Number::varchar(50), :Name::varchar(50), :Priority::integer) attempt_id)
  else null
  end as attempt_id
from cc_queue q
where q.name = :QueueName and q.type = 0 
limit 1`, map[string]interface{}{"CallId": member.CallId, "Number": member.Number, "Name": member.Name, "Priority": member.Priority,
			"QueueName": queueName, "Provider": member.ProviderId}); err != nil {
			result.Err = err
		} else {
			result.Data = res
		}
	})
}

func (self SqlInboundQueueStore) CancelIfDistributing(attemptId int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		_, err := self.GetMaster().Exec(`update cc_member_attempt a
set state = 7 --, result = 'TIMEOUT'
where a.id = :AttemptId and hangup_at = 0 and state > -1 and state != 5`, map[string]interface{}{"AttemptId": attemptId})

		if err != nil {
			result.Err = err
		}
	})
}

func (self SqlInboundQueueStore) InboundInfo(domainId int64, name string) (*model.InboundQueueInfo, error) {
	var queueInfo *model.InboundQueueInfo

	err := self.GetReplica().SelectOne(&queueInfo, `select id, name, timeout, updated_at, max_calls, active_calls.cnt as active_calls, enabled, calendar.ready as calendar_ready
from cc_queue q,
 lateral (
   select exists(
     select *
     from calendar_accept_of_day d
       inner join calendar c2 on d.calendar_id = c2.id
     where d.calendar_id = q.calendar_id AND
           (to_char(current_timestamp AT TIME ZONE c2.timezone, 'SSSS') :: int / 60)
           between d.start_time_of_day AND d.end_time_of_day
     ) as ready
 ) calendar
 left join lateral (
    select count(*) cnt
    from cc_member_attempt a
    where a.hangup_at = 0 and a.queue_id = q.id
 ) active_calls on true
where q.type = 0 and q.domain_id = :DomainId and q.name = :QueueName
limit 1`, map[string]interface{}{"DomainId": domainId, "QueueName": name})

	return queueInfo, err
}
