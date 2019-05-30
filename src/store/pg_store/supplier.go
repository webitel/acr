package pg_store

import (
	dbsql "database/sql"
	sqltrace "log"

	"context"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
	"github.com/webitel/wlog"
	"os"
	"sync/atomic"
	"time"
)

const (
	DB_PING_ATTEMPTS     = 18
	DB_PING_TIMEOUT_SECS = 10
)

const (
	EXIT_CREATE_TABLE = 100
	EXIT_DB_OPEN      = 101
	EXIT_PING         = 102
	EXIT_NO_DRIVER    = 103
)

type SqlSupplierOldStores struct {
	defaultRoute   store.DefaultRouteStore
	extensionRoute store.ExtensionRouteStore
	publicRoute    store.PublicRouteStore
	privateRoute   store.PrivateRouteStore
	routeVariables store.RouteVariablesStore
	inboundQueue   store.InboundQueueStore
	call           store.CallStore
	callbackQueue  store.CallbackQueueStore
	endpoint       store.EndpointStore
}

type SqlSupplier struct {
	rrCounter      int64
	srCounter      int64
	next           store.LayeredStoreSupplier
	master         *gorp.DbMap
	replicas       []*gorp.DbMap
	searchReplicas []*gorp.DbMap
	oldStores      SqlSupplierOldStores
	settings       *model.SqlSettings
	lockedToMaster bool
}

func NewSqlSupplier(settings model.SqlSettings) *SqlSupplier {
	supplier := &SqlSupplier{
		rrCounter: 0,
		srCounter: 0,
		settings:  &settings,
	}
	supplier.initConnection()

	supplier.oldStores.defaultRoute = NewSqlDefaultRouteStore(supplier)
	supplier.oldStores.extensionRoute = NewSqlExtensionRouteStore(supplier)
	supplier.oldStores.publicRoute = NewSqlPublicRouteStore(supplier)
	supplier.oldStores.privateRoute = NewSqlPrivateRouteStore(supplier)
	supplier.oldStores.routeVariables = NewSqlRouteVariablesStore(supplier)
	supplier.oldStores.inboundQueue = NewSqlInboundQueueStore(supplier)
	supplier.oldStores.call = NewSqlCallStore(supplier)
	supplier.oldStores.callbackQueue = NewSqlCallbackQueueStore(supplier)
	supplier.oldStores.endpoint = NewSqlEndpointStore(supplier)

	return supplier
}

func (ss *SqlSupplier) GetAllConns() []*gorp.DbMap {
	all := make([]*gorp.DbMap, len(ss.replicas)+1)
	copy(all, ss.replicas)
	all[len(ss.replicas)] = ss.master
	return all
}

func setupConnection(con_type string, dataSource string, settings *model.SqlSettings) *gorp.DbMap {
	db, err := dbsql.Open(*settings.DriverName, dataSource)
	if err != nil {
		wlog.Critical(fmt.Sprintf("failed to open SQL connection to err:%v", err.Error()))
		time.Sleep(time.Second)
		os.Exit(EXIT_DB_OPEN)
	}

	for i := 0; i < DB_PING_ATTEMPTS; i++ {
		wlog.Info(fmt.Sprintf("pinging SQL %v database", con_type))
		ctx, cancel := context.WithTimeout(context.Background(), DB_PING_TIMEOUT_SECS*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err == nil {
			break
		} else {
			if i == DB_PING_ATTEMPTS-1 {
				wlog.Critical(fmt.Sprintf("failed to ping DB, server will exit err=%v", err))
				time.Sleep(time.Second)
				os.Exit(EXIT_PING)
			} else {
				wlog.Error(fmt.Sprintf("failed to ping DB retrying in %v seconds err=%v", DB_PING_TIMEOUT_SECS, err))
				time.Sleep(DB_PING_TIMEOUT_SECS * time.Second)
			}
		}
	}

	db.SetMaxIdleConns(*settings.MaxIdleConns)
	db.SetMaxOpenConns(*settings.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(*settings.ConnMaxLifetimeMilliseconds) * time.Millisecond)

	var dbmap *gorp.DbMap

	if *settings.DriverName == model.DATABASE_DRIVER_POSTGRES {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	} else {
		wlog.Critical("failed to create dialect specific driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_NO_DRIVER)
	}

	if settings.Trace {
		dbmap.TraceOn("", sqltrace.New(os.Stdout, "sql-trace:", sqltrace.Lmicroseconds))
	}

	return dbmap
}

func (s *SqlSupplier) initConnection() {
	s.master = setupConnection("master", *s.settings.DataSource, s.settings)

	if len(s.settings.DataSourceReplicas) > 0 {
		s.replicas = make([]*gorp.DbMap, len(s.settings.DataSourceReplicas))
		for i, replica := range s.settings.DataSourceReplicas {
			s.replicas[i] = setupConnection(fmt.Sprintf("replica-%v", i), replica, s.settings)
		}
	}

	if len(s.settings.DataSourceSearchReplicas) > 0 {
		s.searchReplicas = make([]*gorp.DbMap, len(s.settings.DataSourceSearchReplicas))
		for i, replica := range s.settings.DataSourceSearchReplicas {
			s.searchReplicas[i] = setupConnection(fmt.Sprintf("search-replica-%v", i), replica, s.settings)
		}
	}
}

func (ss *SqlSupplier) GetMaster() *gorp.DbMap {
	return ss.master
}

func (ss *SqlSupplier) GetReplica() *gorp.DbMap {
	if len(ss.settings.DataSourceReplicas) == 0 || ss.lockedToMaster {
		return ss.GetMaster()
	}

	rrNum := atomic.AddInt64(&ss.rrCounter, 1) % int64(len(ss.replicas))
	return ss.replicas[rrNum]
}

func (ss *SqlSupplier) DriverName() string {
	return *ss.settings.DriverName
}

func (ss *SqlSupplier) DefaultRoute() store.DefaultRouteStore {
	return ss.oldStores.defaultRoute
}

func (ss *SqlSupplier) ExtensionRoute() store.ExtensionRouteStore {
	return ss.oldStores.extensionRoute
}

func (ss *SqlSupplier) PublicRoute() store.PublicRouteStore {
	return ss.oldStores.publicRoute
}

func (ss *SqlSupplier) PrivateRoute() store.PrivateRouteStore {
	return ss.oldStores.privateRoute
}

func (ss *SqlSupplier) RouteVariables() store.RouteVariablesStore {
	return ss.oldStores.routeVariables
}

func (ss *SqlSupplier) InboundQueue() store.InboundQueueStore {
	return ss.oldStores.inboundQueue
}

func (ss *SqlSupplier) Call() store.CallStore {
	return ss.oldStores.call
}

func (ss *SqlSupplier) CallbackQueue() store.CallbackQueueStore {
	return ss.oldStores.callbackQueue
}

func (ss *SqlSupplier) Endpoint() store.EndpointStore {
	return ss.oldStores.endpoint
}
