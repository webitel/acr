package pg_store

import (
	"github.com/go-gorp/gorp"
	_ "github.com/lib/pq"
	"github.com/webitel/acr/src/store"
)

type SqlStore interface {
	GetMaster() *gorp.DbMap
	GetReplica() *gorp.DbMap
	GetAllConns() []*gorp.DbMap

	DefaultRoute() store.DefaultRouteStore
	ExtensionRoute() store.ExtensionRouteStore
	PublicRoute() store.PublicRouteStore
	PrivateRoute() store.PrivateRouteStore
	RouteVariables() store.RouteVariablesStore
	InboundQueue() store.InboundQueueStore
	Call() store.CallStore
	CallbackQueue() store.CallbackQueueStore
}
