package fs

import (
	"github.com/webitel/acr/src/pkg/fs/esl"
)

type Server interface {
	Listen()
}

type Connection interface {
	GetVar(name string) (string)
	Hangup(cause string) (err error)
	Execute(app, args string) (*esl.Event, error)
}

type HandleFunc func(Connection)

func NewEsl(addr string, onConnect, onDisconnect HandleFunc) Server {
	return esl.New(addr, func(connection *esl.Connection) {
		onConnect(connection)
	}, func(connection *esl.Connection) {
		onDisconnect(connection)
	})
}
