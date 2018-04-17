package esl

import (
	esl "github.com/fiorix/go-eventsocket/eventsocket"
	"fmt"
	"sync"
)

type Connection struct {
	sync.Mutex
	esl *esl.Connection
	channelData  *esl.Event
	ev *esl.Event
	context string
	cbWrapper map[string]chan *esl.Event
	uuid string
}

func (c *Connection) GetVar(name string) (string)  {
	c.Lock()
	defer c.Unlock()

	if c.ev != nil {
		return c.ev.Get(name)
	}

	return c.channelData.Get(name)
}

func (c *Connection) Hangup(cause string) (err error)   {
	 var e  *esl.Event
	 e, err = c.esl.Execute("hangup", cause, true)
	 fmt.Println(e)
	 return err
}
func (c *Connection) Execute(app, args string) (err error) {
	c.esl.Execute(app, args, false)

	//var e  *esl.Event
	//guid, _ := uuid.NewV4()
	//e, err = c.esl.SendMsg(esl.MSG{
	//	 "call-command":     "execute",
	//	 "execute-app-name": app,
	//	 "execute-app-arg":  args,
	//	 //"event-lock":  "true",
	//	 //"event-uuid": guid.String(),
	// }, "", "")
	//
	 //fmt.Println(guid, e)
	 //if e != nil {}
	 return err
}

func (c *Connection) GetUuid() string {
	return c.uuid
}

func (c *Connection) GetVariable(string) string {
	return ""
}

func (c *Connection) GetDisconnected() bool {
	return false
}
