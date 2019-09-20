package app

import "github.com/webitel/acr/src/model"

func (a *App) GetDistinctDevices(domainId int, request []*model.EndpointsRequest) ([]*model.EndpointsResponse, error) {
	if result := <-a.Store.Endpoint().GetDistinctDevices(domainId, request); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.([]*model.EndpointsResponse), nil
	}
}
