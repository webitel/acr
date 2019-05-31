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

func (ss SqlEndpointStore) GetDistinctDevices(request []*model.EndpointsRequest) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		b, _ := json.Marshal(request)
		var endpoints []*model.EndpointsResponse
		_, err := ss.GetReplica().Select(&endpoints, `with ss as (
  select a.el->>'type' as type, a.el->>'name' as name, a.el->>'key' pos
  from (
    select json_array_elements(:Request::json) as el
  ) a
)
select t.pos, json_agg(json_build_object('id', t.id, 'name', t.name, 'devices', t.devices)) as endpoints
from (
  select
     distinct on(d.id) ss.pos, d.*
  from ss
  left join lateral (
    WITH RECURSIVE groups AS
     (
       SELECT id, u.canlogin, null::name as name
       FROM wbt_auth u
       WHERE u.dc = 1
         and u.rolname = ss.name
         and u.canlogin = false

       UNION DISTINCT

       SELECT rel.member, a.canlogin, a.rolname
       FROM wbt_auth_member rel
              inner join wbt_auth a on a.dc = rel.dc and a.id = rel.member
              inner join groups g on g.id = rel.role
       WHERE rel.dc = 1
     )
     select g.id, g.name, d.devices
     from groups g,
      lateral ( select array(select d.device_id as dev
        from wbt_device d
        where d.dc = 1 and d.owner_id = g.id) as devices
      ) d
     where ss.type = 'group' and g.canlogin = true

    union all

    select u.id, u.rolname, d.devices
    from wbt_auth u
    left join lateral (
      select array(select d.device_id as d
        from wbt_device d
        where d.dc = u.dc and d.owner_id = u.id) as devices
    ) d on true
    where ss.type = 'user' and u.dc = 1 and u.canlogin = true and u.rolname = ss.name

  ) d on true
) t
group by t.pos`, map[string]interface{}{"Request": string(b)})

		if err != nil {
			result.Err = err
		} else {
			result.Data = endpoints
		}

	})
}
