package esl

import (
	esl "github.com/fiorix/go-eventsocket/eventsocket"
	"fmt"
)

type Connection struct {
	esl *esl.Connection
	channelData  *esl.Event
	context string
}

func (c *Connection) Hangup(cause string) (err error)   {
	 var e  *esl.Event
	 e, err = c.esl.Execute("hangup", cause, true)
	 fmt.Println(e)
	 return err
}
func (c *Connection) Execute(app, args string) (err error) {
	 var e  *esl.Event
	 e, err = c.esl.Execute(app, args, true)
	 fmt.Println(e)
	 return err
}

func (c *Connection) GetUuid() string {
	return ""
}

func (c *Connection) GetVariable(string) string {
	return ""
}

func (c *Connection) GetDisconnected() bool {
	return false
}
