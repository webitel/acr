package app

import "github.com/webitel/acr/src/rpc"

func (app *App) FireRPCEventToEngine(rk string, option rpc.PublishingOption) error {
	return app.rpc.Fire("engine", rk, option)
}
func (app *App) FireRPCEventToHook(msg []byte) error {
	return app.rpc.Hook(msg)
}

func (app *App) FireRPCEventToStorage(rk string, option rpc.PublishingOption) error {
	return app.rpc.Fire("Storage.Commands", rk, option)
}

func (app *App) GetRPCCommandsQueueName() string {
	return app.rpc.GetCommandsQueueName()
}

func (app *App) AddRPCCommands(uuid string) rpc.ApiT {
	return app.rpc.AddCommands(uuid)
}

func (app *App) RemoveRPCCommands(uuid string) {
	app.rpc.RemoveCommands(uuid, rpc.ApiT{})
}
