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
	"github.com/webitel/acr/src/pkg/models"
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

func (a *ACR) CreateCall(destinationNumber string, c *esl.SConn, cf *models.CallFlow, context call.ContextId) {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.calls[c] = call.MakeCall(destinationNumber, c, cf, a, context)
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
	if c.SwitchUuid == "" {
		logger.Error("Bad connection 'Core-UUID': ", c.ChannelData)
		return
	}
	if _, ok := a.GlobalVars[c.SwitchUuid]; !ok {
		a.GlobalVars[c.SwitchUuid] = make(map[string]string)
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
				a.GlobalVars[c.SwitchUuid][string(val[0])] = string(val[1])
			}
		}

		logger.Info("Success init global variables: %s", c.SwitchUuid)
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

func (a *ACR) GetDomainVariables(domainName string) (models.DomainVariables, error) {
	return a.DB.GetDomainVariables(domainName)
}

func (a *ACR) SetDomainVariable(domainName, key, value string) error {
	return a.DB.SetDomainVariable(domainName, key, value)
}

func (a *ACR) AddMember(data interface{}) error {
	return a.DB.AddMember(data)
}

func (a *ACR) AddCallbackMember(domainName, queueName, number, widgetName string) (error, int) {
	return a.DB.CreateCallbackMember(domainName, queueName, number, widgetName)
}

func (a *ACR) UpdateMember(id string, data interface{}) error {
	return a.DB.UpdateMember(id, data)
}

func (a *ACR) GetPrivateCallFlow(uuid string, domain string) (models.CallFlow, error) {
	return a.DB.GetPrivateCallFlow(uuid, domain)
}

func (a *ACR) InsertPrivateCallFlow(uuid, domain, timeZone string, deadline int, apps models.ArrayApplications) error {
	return a.DB.InsertPrivateCallFlow(uuid, domain, timeZone, deadline, apps)
}

func (a *ACR) RemovePrivateCallFlow(uuid, domain string) error {
	return a.DB.RemovePrivateCallFlow(uuid, domain)
}

func (a *ACR) ExistsMediaFile(name, typeFile, domainName string) bool {
	return a.DB.ExistsMediaFile(name, typeFile, domainName)
}

func (a *ACR) ExistsDialer(name, domain string) bool {
	return a.DB.ExistsDialer(name, domain)
}

func (a *ACR) ExistsMemberInDialer(name, domain string, data []byte) bool {
	return a.DB.ExistsMemberInDialer(name, domain, data)
}

func (a *ACR) ExistsQueue(name, domain string) bool {
	return a.DB.ExistsQueue(name, domain)
}

func (a *ACR) FindUuidByPresence(presence string) string {
	return a.DB.FindUuidByPresence(presence)
}

func (a *ACR) CountAvailableAgent(queueName string) (count int) {
	return a.DB.CountAvailableAgent(queueName)
}

func (a *ACR) CountAvailableMembers(queueName string) (count int) {
	return a.DB.CountAvailableMembers(queueName)
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

func (a *ACR) FireRPCEventToEngine(rk string, option rpc.PublishingOption) error {
	return a.rpc.Fire("engine", rk, option)
}

func (a *ACR) FireRPCEventToStorage(rk string, option rpc.PublishingOption) error {
	return a.rpc.Fire("Storage.Commands", rk, option)
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
