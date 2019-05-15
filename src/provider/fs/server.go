package fs

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/provider"
	"github.com/webitel/acr/src/provider/fs/eventsocket"
	"github.com/webitel/wlog"
	"io"
	"net"
	"sync"
)

const (
	EVENT_HANGUP_COMPLETE  = "CHANNEL_HANGUP_COMPLETE"
	EVENT_EXECUTE_COMPLETE = "CHANNEL_EXECUTE_COMPLETE"
	EVENT_ANSWER           = "CHANNEL_ANSWER"
)

type ServerImpl struct {
	address         string
	listener        net.Listener
	consume         chan provider.Connection
	startOnce       sync.Once
	didFinishListen chan struct{}
}

func NewCallServer(settings model.CallServerSettings) provider.CallServer {
	return &ServerImpl{
		address: settings.Host,
		consume: make(chan provider.Connection),
	}
}

func (s *ServerImpl) Start() {
	wlog.Info(fmt.Sprintf("started call server on %s", s.address))

	s.startOnce.Do(func() {
		listener, err := net.Listen("tcp", s.address)
		if err != nil {
			panic(err) //TODO
		}

		s.listener = listener

		go func() {
			defer func() {
				wlog.Info("stopped call server")
			}()
			eventsocket.Listen(listener, s.handler)
		}()
	})
}

func (s *ServerImpl) Stop() {
	wlog.Debug("stopping call server")
	if s.listener != nil {
		s.listener.Close()
	}
	close(s.consume)
}

func (s *ServerImpl) Consume() <-chan provider.Connection {
	return s.consume
}

func (s *ServerImpl) handler(c *eventsocket.Connection) {
	e, err := c.Send("connect")
	if err != nil {
		wlog.Error(fmt.Sprintf("connect to call %v error: %s", c.RemoteAddr(), err.Error()))
		return
	}

	uuid := e.Get(HEADER_ID_NAME)

	_, err = c.Send("linger 30")
	if err != nil {
		wlog.Error(fmt.Sprintf("set linger call %s error: %s", uuid, err.Error()))
		return
	}

	_, err = c.Send("filter unique-id " + uuid)
	if err != nil {
		wlog.Error(fmt.Sprintf("call %s filter events error: %s", uuid, err.Error()))
		return
	}

	_, err = c.Send(fmt.Sprintf("events plain %s %s %s", EVENT_HANGUP_COMPLETE, EVENT_EXECUTE_COMPLETE, EVENT_ANSWER))
	if err != nil {
		wlog.Error(fmt.Sprintf("call %s events error: %s", uuid, err.Error()))
		return
	}

	connection := newConnection(c, e)

	defer func() {

		if connection.Stopped() {
			wlog.Debug(fmt.Sprintf("call %s stopped connect %v", uuid, c.RemoteAddr()))
		} else {
			wlog.Warn(fmt.Sprintf("call %s bad close connection %v", uuid, c.RemoteAddr()))
		}

		connection.Lock()
		if len(connection.callbackMessages) > 0 {
			for k, v := range connection.callbackMessages {
				v <- &eventsocket.Event{}
				close(v)
				delete(connection.callbackMessages, k)
			}
		}
		connection.Unlock()

		if connection.lastEvent.Get(HEADER_EVENT_NAME) != EVENT_HANGUP_COMPLETE {
			wlog.Warn(fmt.Sprintf("call %s no found event hangup", connection.Id()))
		}

		connection.connection.Close()
		close(connection.disconnected)
	}()

	wlog.Debug(fmt.Sprintf("receive new call %s connect %v", uuid, c.RemoteAddr()))
	s.consume <- connection

	for {
		if connection.Stopped() {
			break
		}

		e, err = c.ReadEvent()
		if err == io.EOF {
			return
		} else if err != nil {
			wlog.Error(fmt.Sprintf("call %s socket error: %s", uuid, err.Error()))
			continue
		}

		connection.setEvent(e)
	}
}
