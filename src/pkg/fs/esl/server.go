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
	count int64
}

func (s *Server) Listen()  {
	err := esl.ListenAndServe(s.addr, s.handleConnection)
	logger.Error("Stop server: %s", err.Error())
}

func (s *Server) handleConnection(connection *esl.Connection)  {
	fmt.Println("new client:", connection.RemoteAddr())
	var err error
	con := &Connection{
		esl:connection,
		cbWrapper: make(map[string]chan *Event),
		exit: make(chan bool, 1),
	}

	con.channelData, err = connection.Send("connect")
	if err != nil {
		logger.Error("connect: %s", err.Error())
		return
	}
	_, err = connection.Send("myevents")
	if err != nil {
		logger.Error("myevents: %s", err.Error())
		return
	}

	_, err = connection.Send("events CHANNEL_HANGUP_COMPLETE CHANNEL_EXECUTE_COMPLETE")
	if err != nil {
		logger.Error("subscribe: %s", err.Error())
		return
	}

	_, err = connection.Send("linger 30")
	if err != nil {
		logger.Error("linger: %s", err.Error())
		return
	}

	con.context = con.channelData.Get("Channel-Context")
	con.uuid = con.channelData.Get("Unique-Id")

	go s.onConnect(con)

	var loop  = true
	var e *esl.Event

	for loop {
		e, err = connection.ReadEvent()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		if e != nil {
			con.ev = e
		}

		switch con.ev.Header["Event-Name"] {
		case "CHANNEL_EXECUTE_COMPLETE":
			con.OnExecuteComplete()
		case "CHANNEL_HANGUP_COMPLETE":
			loop = false
		}
	}
	connection.Send("exit")
	con.exit <- true
	s.onDisconnect(con)
}

func New(addr string, onConnect, onDisconnect handler) *Server {
	return &Server{
		addr:addr,
		onConnect:onConnect,
		onDisconnect:onDisconnect,
	}
}



