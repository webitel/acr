package app

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/call_center/discovery"
)

type cluster struct {
	app       *App
	discovery discovery.ServiceDiscovery
}

func NewCluster(app *App) *cluster {
	return &cluster{
		app: app,
	}
}

func (c *cluster) Start() error {
	sd, err := discovery.NewServiceDiscovery(c.app.nodeId, c.app.Config().DiscoverySettings.Url, func() (b bool, appError error) {
		return true, nil
	})
	if err != nil {
		return err
	}
	c.discovery = sd

	err = sd.RegisterService(model.APP_SERVICE_NAME, c.app.CallSrv.Host(), c.app.CallSrv.Ip(),
		model.APP_SERVICE_TTL, model.APP_DEREGESTER_CRITICAL_TTL)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) Stop() {
	c.discovery.Shutdown()
}
