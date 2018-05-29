package esl

import (
	esl "github.com/fiorix/go-eventsocket/eventsocket"
	"fmt"
	"sync"
	"github.com/satori/go.uuid"
)

type Connection struct {
	sync.Mutex
	esl *esl.Connection
	channelData  *esl.Event
	ev *esl.Event
	context string
	cbWrapper map[string]chan *esl.Event
	uuid string
	exit chan bool
}

type Event = esl.Event

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
func (c *Connection) Execute(app, args string) (*Event, error) {
	guid, _ := uuid.NewV4()
	ch := make(chan *Event, 1)
	c.Lock()
	c.cbWrapper[guid.String()] = ch
	c.Unlock()
	_, err := c.esl.SendMsg(esl.MSG{
		"call-command":     "execute",
		"execute-app-name": app,
		"execute-app-arg":  args,
		"event-lock":  "false",
		"Event-UUID":  guid.String(),
	}, "", "")

	if err != nil {
		return c.ev, err
	}
	select {
	case <-c.exit:
		return c.ev, nil
	case out, ok := <-ch:
		if !ok {
			return out, nil
		}
		return out, nil
	}
}

func (c *Connection) OnExecuteComplete()  {
	c.Lock()
	defer c.Unlock()
	if v, ok := c.cbWrapper[c.ev.Get("Application-Uuid")]; ok {
		delete(c.cbWrapper, c.ev.Get("Application-Uuid"))
		v <- c.ev
		close(v)
	} else if v, ok := c.cbWrapper[c.ev.Get("Event-Uuid")]; ok {
		delete(c.cbWrapper, c.ev.Get("Event-Uuid"))
		v <- c.ev
		close(v)
	}
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
