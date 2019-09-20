package pg_store

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
)

type SqlRoutingInboundCallStore struct {
	SqlStore
}

func NewSqlRoutingInboundCallStore(sqlStore SqlStore) store.RoutingInboundCallStore {
	st := &SqlRoutingInboundCallStore{sqlStore}
	return st
}

func (s SqlRoutingInboundCallStore) FromGateway(domainId, gatewayId int) (*model.Routing, error) {
	var routing *model.Routing
	err := s.GetReplica().SelectOne(&routing, `select
		sg.id as source_id,
		sg.name as source_name,
		'' as source_data,
		d.dc as domain_id,
		d.dn as domain_name,
		d.timezone_id,
		ct.name as timezone_name,
		sg.scheme_id,
		ars.name as scheme_name,
		ars.scheme,
		ars.debug,
		null as variables
	from wbt_domain d
		inner join calendar_timezones ct on d.timezone_id = ct.id
		inner join sip_gateway sg on sg.dc = d.dc and sg.id = :GatewayId
		inner join acr_routing_scheme ars on ars.id = sg.scheme_id
	where d.dc = :DomainId`, map[string]interface{}{"GatewayId": gatewayId, "DomainId": domainId})

	if err != nil {
		return nil, err
	}
	return routing, nil
}
