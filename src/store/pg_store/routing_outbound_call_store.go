package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlRoutingOutboundCallStore struct {
	SqlStore
}

func NewSqlRoutingOutboundCallStore(sqlStore SqlStore) store.RoutingOutboundCallStore {
	st := &SqlRoutingOutboundCallStore{sqlStore}
	return st
}

func (s SqlRoutingOutboundCallStore) SearchByDestination(domainId int, destination string) (*model.Routing, error) {
	var routing *model.Routing
	err := s.GetReplica().SelectOne(&routing, `select
    r.id as source_id,
    r.name as source_name,
	r.pattern as source_data,
    d.dc as domain_id,
    d.dn as domain_name,
    d.timezone_id,
    ct.name as timezone_name,
    r.scheme_id,
    ars.name as scheme_name,
    ars.scheme,
    ars.debug,
    null as variables
from acr_routing_outbound_call r
    inner join wbt_domain d on d.dc = r.domain_id
    inner join calendar_timezones ct on d.timezone_id = ct.id
    inner join acr_routing_scheme ars on ars.id = r.scheme_id
where r.domain_id = :DomainId and (not r.disabled) and :Destination::varchar(50) ~ r.pattern
order by r.priority asc
limit 1`, map[string]interface{}{"DomainId": domainId, "Destination": destination})
	if err != nil {
		return nil, err
	}
	return routing, nil
}
