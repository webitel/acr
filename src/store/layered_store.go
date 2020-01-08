package store

import "context"

type LayeredStoreDatabaseLayer interface {
	LayeredStoreSupplier
	SqlStore //TODO delete mongo store
}

type LayeredStore struct {
	TmpContext     context.Context
	DatabaseLayer  LayeredStoreDatabaseLayer
	LayerChainHead LayeredStoreSupplier
	NoSqlStore     NoSqlStore
}

func NewLayeredStore(db LayeredStoreDatabaseLayer) Store {
	store := &LayeredStore{
		TmpContext:    context.TODO(),
		DatabaseLayer: db,
	}

	return store
}

//region mongodb TODO remove this store
func (s *LayeredStore) OutboundQueue() OutboundQueueStore {
	return s.NoSqlStore.OutboundQueue()
}

func (s *LayeredStore) GeoLocation() GeoLocationStore {
	return s.NoSqlStore.GeoLocation()
}

func (s *LayeredStore) Calendar() CalendarStore {
	return s.NoSqlStore.Calendar()
}

func (s *LayeredStore) BlackList() BlackListStore {
	return s.NoSqlStore.BlackList()
}

func (s *LayeredStore) Email() EmailStore {
	return s.NoSqlStore.Email()
}

func (s *LayeredStore) Media() MediaStore {
	return s.NoSqlStore.Media()
}

//endregion

func (s *LayeredStore) DefaultRoute() DefaultRouteStore {
	return s.DatabaseLayer.DefaultRoute()
}

func (s *LayeredStore) ExtensionRoute() ExtensionRouteStore {
	return s.DatabaseLayer.ExtensionRoute()
}

func (s *LayeredStore) PublicRoute() PublicRouteStore {
	return s.DatabaseLayer.PublicRoute()
}

func (s *LayeredStore) PrivateRoute() PrivateRouteStore {
	return s.DatabaseLayer.PrivateRoute()
}

func (s *LayeredStore) RouteVariables() RouteVariablesStore {
	return s.DatabaseLayer.RouteVariables()
}

func (s *LayeredStore) InboundQueue() InboundQueueStore {
	return s.DatabaseLayer.InboundQueue()
}

func (s *LayeredStore) Call() CallStore {
	return s.DatabaseLayer.Call()
}

func (s *LayeredStore) CallbackQueue() CallbackQueueStore {
	return s.DatabaseLayer.CallbackQueue()
}

func (s *LayeredStore) Endpoint() EndpointStore {
	return s.DatabaseLayer.Endpoint()
}

func (s *LayeredStore) RoutingInboundCall() RoutingInboundCallStore {
	return s.DatabaseLayer.RoutingInboundCall()
}
func (s *LayeredStore) RoutingOutboundCall() RoutingOutboundCallStore {
	return s.DatabaseLayer.RoutingOutboundCall()
}
