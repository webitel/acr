/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package acr

import (
	"bytes"
	"github.com/webitel/acr/src/pkg/call"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/db"
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/router"
	"github.com/webitel/acr/src/pkg/rpc"
	"sync"
	"sync/atomic"
)

type ACR struct {
	DB         *db.DB
	GlobalVars map[string]map[string]string
	Count      int32
	calls      map[*esl.SConn]*call.Call
	mx         *sync.Mutex
	rpc        *rpc.RPC
}

var acr *ACR

func (a *ACR) CreateCall(destinationNumber string, c *esl.SConn, cf *router.CallFlow) {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.calls[c] = call.MakeCall(destinationNumber, c, cf, a)
}

func (a *ACR) GetGlobalVar(call *call.Call, varName string) (val string, ok bool) {
	val, ok = a.GlobalVars[call.SwitchId][varName]
	return
}
func (a *ACR) GetGlobalVarBySwitchId(switchId, varName string) (val string, ok bool) {
	val, ok = a.GlobalVars[switchId][varName]
	return
}

func (a *ACR) initGlobalVar(c *esl.SConn) {
	uuid := c.ChannelData.Header.Get("Core-UUID")
	if uuid == "" {
		logger.Error("Bad connection 'Core-UUID': ", c.ChannelData)
		return
	}
	if _, ok := a.GlobalVars[uuid]; !ok {
		a.GlobalVars[uuid] = make(map[string]string)
		data, err := c.Api("global_getvar")
		if err != nil {
			logger.Error("Bad api global_getvar: ", err)
			return
		}

		if len(data) < 1 {
			logger.Error("Bad response global_getvar")
			return
		}

		rows := bytes.Split(data, []byte("\n"))
		var val [][]byte
		for i := 0; i < len(rows); i++ {
			val = bytes.SplitN(rows[i], []byte("="), 2)
			if len(val) == 2 {
				a.GlobalVars[uuid][string(val[0])] = string(val[1])
			}
		}

		logger.Info("Success init global variables: %s", uuid)
	}
}

//TODO
func (a *ACR) CheckBlackList(domainName, name, number string) (error, int) {
	return a.DB.CheckBlackList(domainName, name, number)
}

func (a *ACR) GetEmailConfig(domainName string, dataStructure interface{}) error {
	return a.DB.GetEmailConfig(domainName, dataStructure)
}

func (a *ACR) GetCalendar(name, domainName string, dataStructure interface{}) error {
	return a.DB.GetCalendar(name, domainName, dataStructure)
}

func (a *ACR) FindLocation(sysLength int, numbers []string, dataStructure interface{}) error {
	return a.DB.FindLocation(sysLength, numbers, dataStructure)
}

func (a *ACR) GetDomainVariables(domainName string, dataStructure interface{}) error {
	return a.DB.GetDomainVariables(domainName, dataStructure)
}

func (a *ACR) SetDomainVariable(domainName, key, value string) error {
	return a.DB.SetDomainVariable(domainName, key, value)
}

func (a *ACR) addConnection(uuid string) {
	atomic.AddInt32(&a.Count, 1)
	logger.Debug("New connection %s, all %d", uuid, a.Count)
}

func (a *ACR) CloseConnection(uuid string) {
	atomic.AddInt32(&a.Count, -1)
	logger.Debug("Close connection %s, all %d", uuid, a.Count)
}

func (a *ACR) GetRPCCommandsQueueName() string {
	return a.rpc.GetCommandsQueueName()
}

func (a *ACR) AddRPCCommands(uuid string) rpc.ApiT {
	return a.rpc.AddCommands(uuid)
}

func (a *ACR) FireRPCEvent(body []byte, rk string) error {
	return a.rpc.Fire(body, rk)
}

func onConnect(c *esl.SConn) {
	acr.addConnection(c.Uuid)
	acr.initGlobalVar(c)
	acr.routeContext(c)
}

func onDisconnect(con *esl.SConn) {
	acr.rpc.RemoveCommands(con.Uuid, rpc.ApiT{})
	acr.mx.Lock()
	defer acr.mx.Unlock()
	if c, ok := acr.calls[con]; ok {
		delete(acr.calls, con)
		if c.OnDisconnectIterator != nil {
			logger.Debug("Call %s switch to OnDisconnect router", c.Uuid)
			go func(a *ACR, _call *call.Call) {
				_call.OnDisconnectTrigger()
				a.CloseConnection(_call.Uuid)
			}(acr, c)
		} else {
			acr.CloseConnection(c.Uuid)
		}

	} else {
		acr.CloseConnection(con.Uuid)
	}
}

func New() {

	acr = &ACR{
		DB:    db.NewDB(config.Conf.Get("mongodb:uri")),
		calls: make(map[*esl.SConn]*call.Call),
		rpc:   rpc.New(),
	}
	acr.mx = &sync.Mutex{}
	acr.GlobalVars = make(map[string]map[string]string)

	s := esl.NewServer(config.Conf.Get("server:host")+":"+config.Conf.Get("server:ports"), onConnect, onDisconnect)
	logger.Info("Start server %s:%s", config.Conf.Get("server:host"), config.Conf.Get("server:ports"))
	err := s.Listen()
	if err != nil {
		logger.Error("Stop server: %v", err.Error())
	}
}
