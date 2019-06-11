package pg_store

import (
	"encoding/json"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlEndpointStore struct {
	SqlStore
}

func NewSqlEndpointStore(sqlStore SqlStore) store.EndpointStore {
	st := &SqlEndpointStore{sqlStore}
	return st
}

func (ss SqlEndpointStore) GetDistinctDevices(domainId int64, request []*model.EndpointsRequest) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		b, _ := json.Marshal(request)
		var endpoints []*model.EndpointsResponse
		_, err := ss.GetReplica().Select(&endpoints, `with ss as (
  select a.el->>'type' as type, a.el->>'name' as name, a.el->>'key' pos, :DomainId::bigint as dc
  from (
    select json_array_elements(:Request::json) as el
  ) a
)
select ss.pos,
       t.id,
       t.name,
       t.number,
       t.dnd
from ss
left join (
  select
        ss.pos,
        row_number()  over (partition by d.id order by ss.pos asc) as rn,
        d.*
  from ss
  left join lateral (
     WITH RECURSIVE groups AS
      (
        SELECT id, u.canlogin, null::name as name, null::name as number, null::boolean as dnd
        FROM wbt_auth u
        WHERE u.dc = ss.dc
          and u.rolname = ss.name
          and u.canlogin = false

        UNION DISTINCT

        SELECT rel.member, a.canlogin, a.caller_name, a.caller_number, a.do_not_disturb
        FROM wbt_auth_member rel
               inner join wbt_auth a on a.dc = rel.dc and a.id = rel.member
               inner join groups g on g.id = rel.role
        WHERE rel.dc = ss.dc
      )
      select g.id, g.name, g.number, g.dnd
      from groups g
      where ss.type = 'group' and g.canlogin = true and g.number notnull

     union all

     select u.id, u.caller_name, u.caller_number as number, u.do_not_disturb as dnd
     from wbt_auth u
     where  ss.type = 'user' and u.dc = ss.dc and u.canlogin = true and u.rolname = ss.name and u.caller_number notnull

     union all

     select u.id, u.caller_name, u.caller_number as number, u.do_not_disturb as dnd
     from wbt_auth u
     where  ss.type = 'extension' and u.dc = ss.dc and u.canlogin = true and u.caller_number =  ss.name
  ) d on true
) t on t.pos = ss.pos and t.rn = 1`, map[string]interface{}{"Request": string(b), "DomainId": domainId})
		if err != nil {
			result.Err = err
		} else {
			result.Data = endpoints
		}

	})
}
