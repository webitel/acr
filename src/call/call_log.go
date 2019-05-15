package call

import (
	"encoding/json"
	"fmt"
	"github.com/webitel/acr/src/router"
	"github.com/webitel/acr/src/rpc"
	"github.com/webitel/wlog"
)

func (call *Call) LogInfo(app string, data interface{}, result string) {
	wlog.Info(fmt.Sprintf("call %s %s[%v] %s", call.Id(), app, data, result))
	//call.log.Info("", wlog.String("app", app), wlog.Any("data", data), wlog.String("result", result))
}

func (call *Call) LogDebug(app string, data interface{}, result string) {
	wlog.Debug(fmt.Sprintf("call %s %s[%v] %s", call.Id(), app, data, result))
	//call.log.Debug("", wlog.String("app", app), wlog.Any("data", data), wlog.String("result", result))
}

func (call *Call) LogWarn(app string, data interface{}, result string) {
	wlog.Warn(fmt.Sprintf("call %s %s[%v] %s", call.Id(), app, data, result))
	//call.log.Warn("", wlog.String("app", app), wlog.Any("data", data), wlog.String("result", result))
}

func (call *Call) LogError(app string, data interface{}, result string) {
	wlog.Error(fmt.Sprintf("call %s %s[%v] %s", call.Id(), app, data, result))
	//call.log.Error("", wlog.String("app", app), wlog.Any("data", data), wlog.String("result", result))
}

func (call *Call) FireDebugApplication(app router.App) {
	if call.OnlineDebug() && app.GetId() != "" {
		call.debugMap["app-id"] = app.GetId()
		call.debugMap["app-name"] = app.GetName()

		if body, err := json.Marshal(call.debugMap); err == nil {
			call.router.app.FireRPCEventToEngine(fmt.Sprintf("*.broadcast.message.%d", call.RouteId()), rpc.PublishingOption{
				Body: body,
			})
		}
	}
}
