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
                d.name as domain_name,
                d.timezone_id,
                ct.name as timezone_name,
                sg.scheme_id,
                ars.name as scheme_name,
                ars.scheme,
                ars.debug,
                null as variables
        from directory.sip_gateway sg
                inner join directory.wbt_domain d on sg.dc = d.dc
                inner join calendar_timezones ct on d.timezone_id = ct.id
                inner join acr_routing_scheme ars on ars.id = sg.scheme_id
        where sg.id = :GatewayId and sg.dc = :DomainId`, map[string]interface{}{"GatewayId": gatewayId, "DomainId": domainId})

	if err != nil {
		return nil, err
	}
	return routing, nil
}
