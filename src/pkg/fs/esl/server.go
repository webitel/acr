package esl

import (
	esl "github.com/fiorix/go-eventsocket/eventsocket"
	"fmt"
	"sync"
	"github.com/webitel/acr/src/pkg/logger"
)

type handler func(*Connection)

type Server struct {
	sync.Mutex
	connections map[string]*esl.Connection
	addr string
	onConnect handler
	onDisconnect handler
}

func (s *Server) Listen()  {
	esl.ListenAndServe(s.addr, s.handleConnection)
}

func (s *Server) handleConnection(connection *esl.Connection)  {
	fmt.Println("new client:", connection.RemoteAddr())
	var err error
	con := &Connection{
		esl:connection,
	}

	con.channelData, err = connection.Send("connect")
	if err != nil {
		logger.Error("connect: %s", err.Error())
		return
	}
	_, err = connection.Send("myevents")
	if err != nil {
		logger.Error("subscribe: %s", err.Error())
		return
	}
	_, err = connection.Send("linger 10")
	if err != nil {
		logger.Error("linger: %s", err.Error())
		return
	}

	con.context = con.channelData.Header["Channel-Context"]

	go s.onConnect(con)

	var ev *esl.Event
	for {
		ev, err = connection.ReadEvent()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		if ev.Header["Event-Name"] == "CHANNEL_HANGUP_COMPLETE" {
			//ev.PrettyPrint()
			break
		}
	}
	connection.Send("exit")
	s.onDisconnect(con)
}

func New(addr string, onConnect, onDisconnect handler) *Server {
	return &Server{
		addr:addr,
		onConnect:onConnect,
		onDisconnect:onDisconnect,
	}
}



