package app

import "github.com/webitel/acr/src/model"

func (app *App) CreateCallbackMember(domain, queue, number, widgetName string) (*model.CallbackMember, error) {
	result := <-app.Store.CallbackQueue().CreateMember(domain, queue, number, widgetName)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.CallbackMember), nil
}
