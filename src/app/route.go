package app

import (
	"github.com/webitel/acr/src/model"
)

func (a *App) GetDefaultRoute(domain, destination string) (*model.CallFlow, error) {
	result := <-a.Store.DefaultRoute().Get(domain, destination)

	if result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.CallFlow), nil
	}
}

func (a *App) GetExtensionRoute(domain, extension string) (*model.CallFlow, error) {
	result := <-a.Store.ExtensionRoute().Get(domain, extension)

	if result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.CallFlow), nil
	}
}

func (a *App) GetPublicRoute(destination string) (*model.CallFlow, error) {
	result := <-a.Store.PublicRoute().Get(destination)

	if result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.CallFlow), nil
	}
}

func (a *App) GetPrivateRoute(domain, id string) (*model.CallFlow, error) {
	result := <-a.Store.PrivateRoute().Get(id, domain)

	if result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.CallFlow), nil
	}
}

func (a *App) GetOutboundIVRRoute(domain, dialerId string) (*model.CallFlow, error) {
	result := <-a.Store.OutboundQueue().GetIVRCallFlow(dialerId, domain)

	if result.Err != nil {
		return nil, result.Err
	}

	ivr := result.Data.(*model.OutboundIVRCallFlow)

	return ivr.ToCallFlow(domain), nil
}
