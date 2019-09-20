package app

import "github.com/webitel/acr/src/model"

func (app *App) GetRoutingFromGateway(domainId, gatewayId int) (*model.Routing, error) {
	return app.Store.RoutingInboundCall().FromGateway(domainId, gatewayId)
}

func (app *App) SearchOutboundRouting(domainId int, destination string) (*model.Routing, error) {
	return app.Store.RoutingOutboundCall().SearchByDestination(domainId, destination)
}
