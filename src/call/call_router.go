package call

import (
	"fmt"
	"github.com/webitel/acr/src/app"
	"github.com/webitel/acr/src/config"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/provider"
	"github.com/webitel/acr/src/utils"
	"github.com/webitel/wlog"
	"runtime/debug"
	"sync/atomic"
)

type CallRouter interface {
	Stop()
}

const (
	CALL_CACHE_SIZE = 10000
)

type CallRouterImpl struct {
	app                      *app.App
	callCount                int32
	globalVarsStore          utils.GlobalVariableStore
	applications             Applications
	didFinishListen          chan struct{}
	stopped                  chan struct{}
	protectedVariables       [][]byte
	domainCache              *utils.Cache
	defaultPublicRouteNumber *string
}

func InitCallRouter(app *app.App) CallRouter {
	router := &CallRouterImpl{
		app:             app,
		globalVarsStore: utils.NewGlobalVariablesMemoryStore(),
		didFinishListen: make(chan struct{}),
		stopped:         make(chan struct{}),
		domainCache:     utils.NewLru(utils.ParseIntValueFromString(config.Conf.Get("application:maxCallCacheSize"), CALL_CACHE_SIZE)),
	}

	defPublicNum := config.Conf.Get("defaultPublicRout")
	if defPublicNum != "" && defPublicNum != "<nil>" {
		wlog.Info(fmt.Sprintf("setup default public number %s", defPublicNum))
		router.defaultPublicRouteNumber = &defPublicNum
	}

	router.initProtectedVariables()
	router.initApplications()

	go router.listenCalls()
	return router
}

func (router *CallRouterImpl) listenCalls() {
	defer func() {
		wlog.Info("stopped route call")
	}()
	for {
		select {
		case <-router.didFinishListen:
			wlog.Info("receive stop route call")
			close(router.stopped)
			return
		case connection, ok := <-router.app.CallSrv.Consume():
			if !ok {
				return
			}
			router.addCallConnection(connection)
			go func() {
				router.initGlobalsVars(connection)
				router.handleCallConnection(connection)
				router.removeCallConnection(connection)
			}()
		}
	}
}

func (router *CallRouterImpl) Stop() {
	close(router.didFinishListen)
	<-router.stopped
}

func (router *CallRouterImpl) addCallConnection(callConnection provider.Connection) {
	atomic.AddInt32(&router.callCount, 1)

	wlog.Debug(fmt.Sprintf("add call %s [%s], all cals %d", callConnection.Id(), callConnection.Node(), router.callCount))
}

func (router *CallRouterImpl) removeCallConnection(callConnection provider.Connection) {
	atomic.AddInt32(&router.callCount, -1)
	wlog.Debug(fmt.Sprintf("remove call %s [%s], all cals %d", callConnection.Id(), callConnection.Node(), router.callCount))
}

func (router *CallRouterImpl) initGlobalsVars(callConn provider.Connection) {
	if !router.globalVarsStore.ExistsNode(callConn.NodeId()) {
		variables, err := callConn.GetGlobalVariables()
		if err != nil {
			wlog.Error(fmt.Sprintf("init global variable node %s error: %s", callConn.Node(), err.Error()))
			return
		}
		wlog.Info(fmt.Sprintf("init global variable node %s successful", callConn.Node()))
		router.globalVarsStore.SetNodeVariables(callConn.NodeId(), variables)
	}
}

func (router *CallRouterImpl) handleCallConnection(callConn provider.Connection) {
	defer func() {
		if r := recover(); r != nil {
			wlog.Critical(fmt.Sprintf("critical error %v", r))
			debug.PrintStack()
		}
	}()

	fmt.Println(callConn.Context())

	wlog.Debug(fmt.Sprintf("call %s [domain_id=%d direction=%s user_id=%d] - %s", callConn.Id(), callConn.DomainId(),
		callConn.Direction(), callConn.UserId(), callConn.Destination()))

	call := NewCall(router, callConn)

	call.Execute("set", "fire_talk_events=true")
	call.Execute("set", "fire_not_talk_events=true")

	switch call.Direction() {
	case model.CALL_DIRECTION_INBOUND:
		router.handleInboundCall(call)
	case model.CALL_DIRECTION_OUTBOUND, model.CALL_DIRECTION_INTERNAL:
		router.handleOutboundCall(call)
		//case model.CONTEXT_DIALER:
		//	router.handleDialerContext(call)
		//case model.CONTEXT_PRIVATE:
		//	router.handlePrivateContext(call)
		//	break
	}

	if call.callRouting == nil && call.GetVariable("sip_h_X-Webitel-ParentId") != "" {
		call.Execute("answer", "")
		call.Api("uuid_park " + call.Id())
		call.Api("uuid_bridge " + call.GetVariable("sip_h_X-Webitel-ParentId") + " " + call.Id())

		//fmt.Println(fmt.Sprintf("uuid_bridge %s %s", call.Id(), call.GetVariable("sip_h_X-Webitel-ParentId")))
		//call.Api(fmt.Sprintf("uuid_bridge %s %s", call.Id(), call.GetVariable("sip_h_X-Webitel-ParentId")))

		call.SetBreak()
		return
	}

	if call.callRouting == nil {
		call.Hangup(model.HANGUP_NO_ROUTE_DESTINATION)
		return
	}

	call.Route()
}

func (router *CallRouterImpl) handleDialerContext(call *Call) {
	var err error
	dialerId := call.GetVariable(model.CALL_VARIABLE_DIALER_ID)
	domain := call.GetVariable(model.CALL_VARIABLE_DOMAIN_NAME)

	if domain == "" || dialerId == "" {
		wlog.Warn(fmt.Sprintf("call %s not found %s or %s", call.Id(), model.CALL_VARIABLE_DOMAIN_NAME, model.CALL_VARIABLE_DIALER_ID))
		return
	}

	//call.callRouting, err = router.app.GetOutboundIVRRoute(domain, dialerId)

	if err != nil {
		wlog.Error(fmt.Sprintf("call %s error: %s", call.Id(), err.Error()))
		return
	}

}

func (router *CallRouterImpl) handleInboundCall(call *Call) {
	var err error

	wlog.Debug(fmt.Sprintf("call %s inbound from domain_id=%d && gateway_id=%d", call.Id(), call.DomainId(), call.InboundGatewayId()))

	if call.callRouting, err = router.app.GetRoutingFromGateway(call.DomainId(), call.InboundGatewayId()); err != nil {
		wlog.Error(fmt.Sprintf("call %s fetch routing error %s", call.Id(), err.Error()))
		return
	}

	if err = call.SetDirection(model.CALL_DIRECTION_INBOUND); err != nil {
		wlog.Error(fmt.Sprintf("call %s set direction error %s", call.Id(), err.Error()))
		return
	}

	if err = call.Set(model.CALL_VARIABLE_FORCE_TRANSFER_CONTEXT, model.CONTEXT_DEFAULT); err != nil {
		wlog.Error(fmt.Sprintf("call %s set %s error: %s", call.Id(), model.CALL_VARIABLE_FORCE_TRANSFER_CONTEXT, err.Error()))
		return
	}
}

func (router *CallRouterImpl) handleOutboundCall(call *Call) {
	var err error

	if call.callRouting, err = router.app.SearchOutboundRouting(call.DomainId(), call.Destination()); err != nil {
		wlog.Error(fmt.Sprintf("call %s fetch routing error %s", call.Id(), err.Error()))
		return
	}
	if err = UnSet(call.RootScope(), call, "sip_h_call-info"); err != nil {
		wlog.Error(fmt.Sprintf("call %s un set sip_h_call-info error: %s", call.Id(), err.Error()))
		return
	}

	call.SetDirection(model.CALL_DIRECTION_OUTBOUND)
}

func (route *CallRouterImpl) handlePrivateContext(call *Call) {
	panic("FIXME")
	//var err error
	//domain := call.GetVariable(model.CALL_VARIABLE_DOMAIN_NAME)

	//call.callFlow, err = route.app.GetPrivateRoute(domain, call.Destination())
	//if err != nil && err != sql.ErrNoRows {
	//	wlog.Error(fmt.Sprintf("call %s GetPrivateRoute error %s", call.Id(), err.Error()))
	//	return
	//}
}

func (router *CallRouterImpl) AddToDomainCache(call *Call, key, value string, expireSec int64) {
	router.domainCache.AddWithExpiresInSecs(makeDomainKey(call.Domain(), key), value, expireSec)
}

func (router *CallRouterImpl) RemoveFromDomainCache(call *Call, key string) {
	router.domainCache.Remove(makeDomainKey(call.Domain(), key))
}

func (router *CallRouterImpl) GetFromDomainCache(call *Call, key string) (string, bool) {
	v, ok := router.domainCache.Get(makeDomainKey(call.Domain(), key))
	return fmt.Sprintf("%v", v), ok
}

func makeDomainKey(domain, key string) string {
	return fmt.Sprintf("%s-%v", domain, key)
}
