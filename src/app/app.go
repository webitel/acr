package app

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/provider"
	"github.com/webitel/acr/src/provider/fs"
	"github.com/webitel/acr/src/rpc"
	"github.com/webitel/acr/src/store"
	"github.com/webitel/acr/src/store/mongo_store"
	"github.com/webitel/acr/src/store/pg_store"
	"github.com/webitel/acr/src/utils"
	"github.com/webitel/wlog"
)

type App struct {
	Log     *wlog.Logger
	CallSrv provider.CallServer
	Store   store.Store
	config  *model.Config
	rpc     *rpc.RPC
}

func New(options ...string) *App {
	conf, err := utils.LoadConfig()
	if err != nil {
		panic(err.Error())
	}

	app := &App{
		config: conf,
	}

	app.Log = wlog.NewLogger(&wlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleLevel:  app.config.LogLevel,
	})

	wlog.RedirectStdLog(app.Log)
	wlog.InitGlobalLogger(app.Log)

	app.Store = store.NewLayeredStore(
		pg_store.NewSqlSupplier(app.config.SqlSettings),
		mongo_store.NewNoSqlSupplier(app.config.NoSqlSettings))

	app.CallSrv = fs.NewCallServer(app.config.CallServerSettings)

	app.CallSrv.Start()

	app.rpc = rpc.New()

	return app
}

func (a *App) Shutdown() {
	wlog.Info("stopping app...")
	defer wlog.Info("stopped app.")

	if a.CallSrv != nil {
		a.CallSrv.Stop()
	}
}
