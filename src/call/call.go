package call

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/provider"
	"github.com/webitel/acr/src/router"
	"github.com/webitel/wlog"
	"sync"
	"time"
)

type Call struct {
	provider.Connection
	router                *CallRouterImpl
	callRouting           *model.Routing
	breakCall             bool
	debugLog              bool
	regExp                CallRegExp
	debugMap              map[string]interface{}
	disconnectedVariables map[string]string
	currentIterator       *router.Iterator
	CurrentQueue          string
	onlineDebug           bool
	log                   *wlog.Logger
	sync.RWMutex
}

func NewCall(router *CallRouterImpl, connection provider.Connection) *Call {
	call := &Call{
		Connection:            connection,
		router:                router,
		debugMap:              make(map[string]interface{}),
		disconnectedVariables: make(map[string]string),
	}

	call.log = router.app.Log.With(
		wlog.String("uuid", connection.Id()),
	)

	return call
}

func (call *Call) Domain() string {
	return call.callRouting.DomainName
}

//func (call *Call) DomainId() int64 {
//	return call.callRouting.DomainId
//}

func (call *Call) Timezone() string {
	return call.callRouting.TimezoneName
}

func (call *Call) ValidateApp(name string) bool {
	return call.router.ExistsApplication(name)
}

func (call *Call) GetDate() (now time.Time) {
	if call.Timezone() != "" {
		if loc, err := time.LoadLocation(call.Timezone()); err == nil {
			now = time.Now().In(loc)
			return
		}
	}

	now = time.Now()
	return
}

func (c *Call) SetBreak() {
	if c.GetBreak() {
		return
	}
	c.setBreak(true)
}

func (c *Call) UnSetBreak() {
	c.setBreak(false)
}

func (c *Call) IsDebugLog() bool {
	return c.callRouting.Debug
}

func (c *Call) OnlineDebug() bool {
	return c.onlineDebug
}

func (c *Call) RouteId() int {
	panic("FIXME")
	//return c.callFlow.Id
	return 0
}

func (c *Call) setBreak(val bool) {
	c.Lock()
	defer c.Unlock()
	c.breakCall = val
	c.LogDebug("break", fmt.Sprintf("%v", val), "success")
}

func (c *Call) GetBreak() bool {
	c.RLock()
	defer c.RUnlock()

	return c.breakCall
}

func (call *Call) Route() {
	defer call.Reporting()
	defer call.router.app.RemoveRPCCommands(call.Id())

	if call.Timezone() != "" {
		if err := call.Set(model.CALL_VARIABLE_TIMEZONE_NAME, call.Timezone()); err != nil {
			wlog.Error(fmt.Sprintf("call %s set timezone error: %s", call.Id(), err.Error()))
			return
		}
	}

	if call.GetVariable(model.CALL_VARIABLE_DEFAULT_LANGUAGE_NAME) == model.CALL_LANGUAGE_RU {
		call.Set(model.CALL_VARIABLE_SOUND_PREF_NAME, model.CALL_LANGUAGE_RU_DIRECTORY)
	} else {
		call.Set(model.CALL_VARIABLE_SOUND_PREF_NAME, model.CALL_LANGUAGE_DEFAULT_DIRECTORY)
	}

	call.regExp = setupNumber(call.callRouting.SourceData, call.Destination())

	if call.GetVariable("presence_data") == "" {
		SetVar(call, "presence_data="+call.Domain())
	}

	//FIXME
	//if call.callRouting.Id > 0 {
	//	SetVar(call, []string{
	//		fmt.Sprintf("%s=%d", model.CALL_VARIABLE_SHEMA_ID, call.callFlow.Id),
	//		fmt.Sprintf("%s=%s", model.CALL_VARIABLE_SHEMA_NAME, call.callFlow.Name),
	//	})
	//}

	if call.GetVariable(model.CALL_VARIABLE_DEBUG_NAME) == "true" {
		call.onlineDebug = true
	}

	if call.IsDebugLog() {
		call.debugMap["action"] = "execute"
		call.debugMap["uuid"] = call.Id()
		call.debugMap["domain"] = call.Domain()
	}

	call.initDomainVariables()

	call.currentIterator = router.NewIterator("call", call.callRouting.Scheme, call)
	call.iterateCallApplication(call.currentIterator)
	call.WaitForDisconnect()

	//FIXME add application trigger on disconnect
	//if call.callFlow.OnDisconnect != nil && len(*call.callFlow.OnDisconnect) > 0 {
	//	call.setBreak(false)
	//	call.currentIterator = router.NewIterator("disconnected", *call.callFlow.OnDisconnect, call)
	//	call.iterateDisconnectedCallApplication()
	//}
}
