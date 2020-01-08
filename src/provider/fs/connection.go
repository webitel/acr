package fs

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/webitel/acr/src/provider/fs/eventsocket"
	"github.com/webitel/wlog"
	"strconv"
	"sync"
)

const (
	HEADER_DOMAIN_ID  = "variable_sip_h_X-Webitel-Domain-Id"
	HEADER_USER_ID    = "variable_sip_h_X-Webitel-User-Id"
	HEADER_GATEWAY_ID = "variable_sip_h_X-Webitel-Gateway-Id"

	HEADER_CONTEXT_NAME              = "Channel-Context"
	HEADER_ID_NAME                   = "Unique-ID"
	HEADER_DIRECTION_NAME            = "variable_sip_h_X-Webitel-Direction"
	HEADER_EVENT_NAME                = "Event-Name"
	HEADER_EVENT_ID_NAME             = "Event-UUID"
	HEADER_CORE_ID_NAME              = "Core-UUID"
	HEADER_CORE_NAME                 = "FreeSWITCH-Switchname"
	HEADER_APPLICATION_ID_NAME       = "Application-UUID"
	HEADER_APPLICATION_NAME          = "Application"
	HEADER_APPLICATION_DATA_NAME     = "Application-Data"
	HEADER_APPLICATION_RESPONSE_NAME = "Application-Response"
	HEADER_HANGUP_CAUSE_NAME         = "variable_hangup_cause"
	HEADER_CONTENT_TYPE_NAME         = "Content-Type"
	HEADER_CONTENT_DISPOSITION_NAME  = "Content-Disposition"

	HEADER_CHANNEL_DESTINATION_NAME  = "Channel-Destination-Number"
	HEADER_CALLER_DESTINATION_NAME   = "Caller-Destination-Number"
	HEADER_VARIABLE_DESTINATION_NAME = "variable_destination_number"
)

var errExecuteAfterHangup = errors.New("not allow after hangup")

type ConnectionImpl struct {
	uuid             string
	nodeId           string
	nodeName         string
	context          string
	destination      string
	stopped          bool
	direction        string
	gatewayId        int
	domainId         int
	userId           int
	disconnected     chan struct{}
	lastEvent        *eventsocket.Event
	connection       *eventsocket.Connection
	callbackMessages map[string]chan *eventsocket.Event
	variables        map[string]string
	hangupCause      string
	sync.RWMutex
}

func newConnection(baseConnection *eventsocket.Connection, dump *eventsocket.Event) *ConnectionImpl {
	connection := &ConnectionImpl{
		uuid:             dump.Get(HEADER_ID_NAME),
		nodeId:           dump.Get(HEADER_CORE_ID_NAME),
		nodeName:         dump.Get(HEADER_CORE_NAME),
		context:          dump.Get(HEADER_CONTEXT_NAME),
		direction:        dump.Get(HEADER_DIRECTION_NAME),
		gatewayId:        getIntFromStr(dump.Get(HEADER_GATEWAY_ID)),
		domainId:         getIntFromStr(dump.Get(HEADER_DOMAIN_ID)),
		userId:           getIntFromStr(dump.Get(HEADER_USER_ID)),
		connection:       baseConnection,
		lastEvent:        dump,
		callbackMessages: make(map[string]chan *eventsocket.Event),
		disconnected:     make(chan struct{}),
		variables:        make(map[string]string),
	}
	connection.initDestination(dump)
	connection.updateVariablesFromEvent(dump)
	return connection
}

func getIntFromStr(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

func (c *ConnectionImpl) Id() string {
	return c.uuid
}

func (c *ConnectionImpl) DomainId() int {
	return c.domainId
}

func (c *ConnectionImpl) UserId() int {
	return c.userId
}

func (c *ConnectionImpl) Context() string {
	return c.context
}

func (c *ConnectionImpl) InboundGatewayId() int {
	return c.gatewayId
}

func (c *ConnectionImpl) Direction() string {
	return c.direction
}

func (c *ConnectionImpl) PrintLastEvent() {
	if c.lastEvent != nil {
		c.lastEvent.PrettyPrint()
	}
}

func (c *ConnectionImpl) SetDirection(direction string) error {
	if c.direction == "" {
		if err := c.Execute("set", "webitel_direction="+direction); err != nil {
			return err
		}
		c.direction = direction
	}
	return nil
}

func (c *ConnectionImpl) Set(key, value string) error {
	return c.Execute("set", fmt.Sprintf("%s=%s", key, value))
}

func (c *ConnectionImpl) Export(data string) error {
	return c.Execute("export", fmt.Sprintf("nolocal:%s", data))
}

func (c *ConnectionImpl) initDestination(dump *eventsocket.Event) {
	c.destination = dump.Get(HEADER_CHANNEL_DESTINATION_NAME)
	if c.destination != "" {
		return
	}

	c.destination = dump.Get(HEADER_CALLER_DESTINATION_NAME)
	if c.destination != "" {
		return
	}

	c.destination = dump.Get(HEADER_VARIABLE_DESTINATION_NAME)
	if c.destination != "" {
		return
	}

}

func (c *ConnectionImpl) Destination() string {
	return c.destination
}

func (c *ConnectionImpl) NodeId() string {
	return c.nodeId
}

func (c *ConnectionImpl) Node() string {
	return c.nodeName
}

func (c *ConnectionImpl) setEvent(event *eventsocket.Event) {
	c.Lock()
	defer c.Unlock()
	if event.Get(HEADER_EVENT_NAME) != "" {
		c.lastEvent = event
		c.updateVariablesFromEvent(event)

		switch event.Get(HEADER_EVENT_NAME) {
		case EVENT_EXECUTE_COMPLETE:
			if s, ok := c.callbackMessages[event.Get(HEADER_APPLICATION_ID_NAME)]; ok {
				delete(c.callbackMessages, event.Get(HEADER_APPLICATION_ID_NAME))
				s <- event
				close(s)
			} else if s, ok := c.callbackMessages[event.Get(HEADER_EVENT_ID_NAME)]; ok {
				delete(c.callbackMessages, event.Get(HEADER_EVENT_ID_NAME))
				s <- event
				close(s)
			}
			wlog.Debug(fmt.Sprintf("call %s executed app: %s %s %s", c.Id(), event.Get(HEADER_APPLICATION_NAME),
				event.Get(HEADER_APPLICATION_DATA_NAME), event.Get(HEADER_APPLICATION_RESPONSE_NAME)))
		case EVENT_HANGUP_COMPLETE:
			c.hangupCause = event.Get(HEADER_HANGUP_CAUSE_NAME)
			wlog.Debug(fmt.Sprintf("call %s hangup %s", c.Id(), c.hangupCause))
			//TODO SET DISCONNECT ROUTE
			c.connection.Send("exit")
			c.stopped = true
		default:
			wlog.Debug(fmt.Sprintf("call %s receive event %s", c.Id(), event.Get(HEADER_EVENT_NAME)))
		}
	} else if event.Get(HEADER_CONTENT_TYPE_NAME) == "text/disconnect-notice" && event.Get(HEADER_CONTENT_DISPOSITION_NAME) == "Disconnected" {

	}
}

func (c *ConnectionImpl) Stopped() bool {
	c.RLock()
	defer c.RUnlock()
	return c.stopped
}

func (c *ConnectionImpl) Api(cmd string) ([]byte, error) {
	res, err := c.connection.Send(fmt.Sprintf("api %s", cmd))
	if err != nil {
		return []byte(""), err
	}

	return []byte(res.Body), nil
}

func (c *ConnectionImpl) HangupCause() string {
	c.RLock()
	defer c.RUnlock()
	return c.hangupCause
}

func (c *ConnectionImpl) Execute(app, args string) error {
	if c.Stopped() {
		return errExecuteAfterHangup
	}

	wlog.Debug(fmt.Sprintf("call %s try execute %s %s", c.uuid, app, args))

	guid, err := uuid.NewV4()
	if err != nil {
		return err
	}
	e := make(chan *eventsocket.Event, 1)

	c.Lock()
	c.callbackMessages[guid.String()] = e
	c.Unlock()

	_, err = c.connection.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": app,
		"execute-app-arg":  args,
		"event-lock":       "false",
		"Event-UUID":       guid.String(),
	}, "", "")

	if err != nil {
		return err
	}

	if c.Stopped() {
		return errExecuteAfterHangup
	}

	<-e
	return nil
}

func (c *ConnectionImpl) Hangup(cause string) error {
	return c.Execute("hangup", cause)
}

func (c *ConnectionImpl) updateVariablesFromEvent(event *eventsocket.Event) {
	for k, _ := range event.Header {
		c.variables[k] = event.Get(k)
	}
}

func (c *ConnectionImpl) GetVariable(name string) (value string) {

	//if v, ok := c.variables[name]; ok {
	//	return v
	//}
	//return ""

	if c.lastEvent != nil {
		value = c.lastEvent.Get(name)
	}

	//TODO bug
	//if e, err := c.connection.Send("getvar aaa"); err == nil {
	//	fmt.Println(e)
	//} else {
	//	fmt.Println(err.Error())
	//}

	return
}

func (c *ConnectionImpl) GetGlobalVariables() (map[string]string, error) {
	variables := make(map[string]string)
	data, err := c.Api("global_getvar")
	if err != nil {
		return variables, err
	}

	rows := bytes.Split(data, []byte("\n"))
	var val [][]byte
	for i := 0; i < len(rows); i++ {
		val = bytes.SplitN(rows[i], []byte("="), 2)
		if len(val) == 2 {
			variables[string(val[0])] = string(val[1])
		}
	}
	return variables, nil
}

func (c *ConnectionImpl) WaitForDisconnect() {
	<-c.disconnected
}

func (c *ConnectionImpl) SendEvent(m map[string]string, name string) error {
	return c.connection.SendEvent(m, name)
}

func (c *ConnectionImpl) DumpVariables() map[string]string {
	return c.variables
}
