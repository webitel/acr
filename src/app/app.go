package app

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/provider"
	"github.com/webitel/acr/src/provider/fs"
	"github.com/webitel/acr/src/rpc"
	"github.com/webitel/acr/src/store"
	"github.com/webitel/acr/src/store/pg_store"
	"github.com/webitel/acr/src/utils"
	"github.com/webitel/wlog"
)

type App struct {
	Log     *wlog.Logger
	nodeId  string
	CallSrv provider.CallServer
	Store   store.Store
	config  *model.Config
	cluster *cluster
	rpc     *rpc.RPC
}

func New(options ...string) *App {
	conf, err := utils.LoadConfig()
	if err != nil {
		panic(err.Error())
	}

	app := &App{
		nodeId: fmt.Sprintf("%s-%s", model.APP_SERVICE_NAME, model.NewId()),
		config: conf,
	}

	app.Log = wlog.NewLogger(&wlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleLevel:  app.config.LogLevel,
	})

	wlog.RedirectStdLog(app.Log)
	wlog.InitGlobalLogger(app.Log)

	app.Store = store.NewLayeredStore(pg_store.NewSqlSupplier(app.config.SqlSettings))

	app.CallSrv = fs.NewCallServer(app.config.CallServerSettings)

	app.CallSrv.Start()

	app.cluster = NewCluster(app)
	if err = app.cluster.Start(); err != nil {
		panic(err.Error())
	}

	app.rpc = rpc.New()

	return app
}

func (a *App) Shutdown() {
	wlog.Info("stopping app...")
	defer wlog.Info("stopped app.")

	if a.CallSrv != nil {
		a.CallSrv.Stop()
	}

	if a.cluster != nil {
		a.cluster.Stop()
	}
}

func (a *App) Config() *model.Config {
	return a.config
}
