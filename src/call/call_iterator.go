package call

import (
	"github.com/webitel/acr/src/router"
)

func (call *Call) Iterator() *router.Iterator {
	return call.currentIterator
}

func (call *Call) SetIterator(i *router.Iterator) {
	call.currentIterator = i
}

func (call *Call) iterateCallApplication(iterator *router.Iterator) {

	defer func() {
		if call.IsDebugLog() {
			call.FireDebugApplication(router.NewBaseApp("disconnect", "end"))
		}
	}()

	for {
		if iterator.IsCancel() {
			call.LogDebug("iterator", iterator.Name(), "cancel")
			break
		}

		if call.Stopped() {
			call.LogDebug("iterator", iterator.Name(), "stopped")
			break
		}

		if call.GetBreak() {
			call.LogDebug("iterator", iterator.Name(), "break")
			break
		}

		appConfig := iterator.NextApp()
		if appConfig == nil {
			call.LogDebug("iterator", iterator.Name(), "EOF")
			break
		}

		if appConfig.IteratorApp() {
			call.FireDebugApplication(appConfig)
			appConfig.Execute(iterator)
		} else if app, ok := call.router.GetApplication(appConfig.GetName()); ok {
			if !app.allowNoConnect && call.Stopped() {
				call.LogWarn(appConfig.GetName(), appConfig.GetArgs(), "not allowed execute when disconnected")
				continue
			}

			call.FireDebugApplication(appConfig)
			if err := app.Execute(iterator, call, appConfig.GetArgs()); err != nil {
				call.LogError("error", iterator.Name(), err.Error())
				call.LogDebug("iterator", iterator.Name(), "stopped")
				break
			}

			if appConfig.IsBreak() {
				call.SetBreak()
				call.LogDebug(appConfig.GetName(), appConfig.GetArgs(), "break")
				break
			}
		} else {
			call.LogError(appConfig.GetName(), appConfig.GetArgs(), "app not found")
		}
	}
}

//TODO
func (call *Call) iterateDisconnectedCallApplication() {
	for {
		if call.GetBreak() {
			break
		}

		appConfig := call.Iterator().NextApp()
		if appConfig == nil {
			call.LogDebug("iterator", call.Iterator().Name(), "EOF")
			break
		}

		if appConfig.IteratorApp() {
			appConfig.Execute(call.Iterator())
		} else if app, ok := call.router.GetApplication(appConfig.GetName()); ok {
			if !app.allowNoConnect && call.Stopped() {
				call.LogWarn(appConfig.GetName(), appConfig.GetArgs(), "not allowed execute when disconnected")
				continue
			}

			if err := app.Execute(call.RootScope(), call, appConfig.GetArgs()); err != nil {
				//TODO
			}

			if appConfig.IsBreak() {
				call.SetBreak()
				break
			}
		} else {
			call.LogError(appConfig.GetName(), appConfig.GetArgs(), "app not found")
		}
	}
}
