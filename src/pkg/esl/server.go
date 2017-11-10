package esl

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/webitel/acr/src/pkg/logger"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Server struct {
	addr         string
	onConnect    onConnect
	onDisconnect onDisconnect
}

type onConnect func(c *SConn)
type onDisconnect func(c *SConn)

type SConn struct {
	EventSocket
	contextName  string
	ChannelData  Event
	Uuid         string
	SwitchUuid   string
	Disconnected bool
	cbWrapper    map[string]chan Event
	mx           *sync.Mutex
	exit         chan bool
}

func (c *SConn) SendLock(req *Message, ttl time.Duration, cmd string, args ...string) (Message, error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	return c.call(req, ttl, cmd, args...)
}

func (c *SConn) SndMsg(app string, arg string, look bool, dump bool) (Event, error) {

	if c.GetDisconnected() || !c.opened {
		return Event{}, errors.New("Socket close")
	}

	if dump {
		return c.SndRecMsg(app, arg, look)
	}

	c.rw.Lock()
	defer c.rw.Unlock()
	var lockString string
	if look {
		lockString = "true"
	} else {
		lockString = "false"
	}

	//fmt.Println(c.Uuid, "->", app)
	msg, err := c.call(&Message{
		Header: Header{
			"call-command":     []string{"execute"},
			"execute-app-name": []string{app},
			"execute-app-arg":  []string{arg},
			"event-lock":       []string{lockString},
		},
	}, 1e8, "sendmsg")

	if err != nil {
		return Event{}, err
	}

	return Event{
		Message: msg,
	}, nil
}

func (conn *SConn) GetCbCount() int {
	return len(conn.cbWrapper)
}

func (esl *SConn) BgApi(cmd string, arg ...string) (resp []byte, err error) {

	cmd = strings.TrimSpace(cmd)

	esl.rw.Lock() // ------------------ < LOCKED >
	recv, err := esl.call(nil, esl.Timeout, (`bgapi ` + cmd), arg...)
	esl.rw.Unlock() // ---------------- < UNLOCK >

	if err != nil {
		return nil, err
	}

	if recv.Header.Get("Content-Type") != ("api/response") {
		return nil, errors.New("eslp: protocol error")
	}

	return recv.Body, nil
}

func (conn *SConn) FireEvent(eventName string, m *Message) (Message, error) {
	conn.rw.Lock() // ------------------ < LOCKED >
	res, err := conn.call(m, 1e6, "sendevent "+eventName)
	conn.rw.Unlock() // ---------------- < UNLOCK >
	return res, err
}

func (conn *SConn) SndRecMsg(app string, arg string, look bool) (Event, error) {
	u := uuid.NewV4().String()
	c := make(chan Event, 1)

	conn.rw.Lock() // ------------------ < LOCKED >
	if _, err := conn.call(nil, 0, "filter Application-UUID "+u); err != nil {
		conn.rw.Unlock() // ---------------- < UNLOCK >
		return Event{}, err
	}

	conn.cbWrapper[u] = c
	_, err := conn.call(&Message{
		Header: Header{
			"call-command":     []string{"execute"},
			"execute-app-name": []string{app},
			"execute-app-arg":  []string{arg},
			"event-lock":       []string{"false"},
			"Event-UUID":       []string{u},
		},
	}, 1e6, "sendmsg")
	conn.rw.Unlock() // ---------------- < UNLOCK >
	if err != nil {
		return Event{}, err
	}

	select {
	case <-conn.exit:
		//delete(conn.cbWrapper, u)
		return Event{}, nil
	case out, ok := <-c:

		if !ok {
			return out, nil
		}
		return out, nil
	}
}

func (c *SConn) OnAnswer() bool {
	//_, err := c.call(nil, 0, "filter Event-Name CHANNEL_ANSWER")
	//if err != nil {
	//	logger.Error("Call %s OnAnswer error %s", c.Uuid, err.Error())
	//	return false
	//}

	ch := make(chan bool, 1)
	c.Event("dialer", EventChannelAnswer, "", nil, func(event Event) {
		ch <- true
	}, nil)

	select {
	case <-c.exit:
		return false
	case <-ch:
		return true
	}
}

func (c *SConn) Hangup(cause string) {
	c.SndMsg("hangup", cause, true, false)
}

func (c *SConn) GetDisconnected() bool {
	return c.Disconnected
}

func (c *SConn) GetContextName() string {
	return c.contextName
}

func handle(conn *SConn, s *Server) {
	defer func() {
		conn.disconnect(nil)
	}()

	conn.mx = &sync.Mutex{}
	conn.exit = make(chan bool, 1)
	conn.cbWrapper = make(map[string]chan Event)

	data, err := conn.call(nil, 0, "connect")
	if err != nil {
		logger.Error("Server: connect connection error: %s", err.Error())
		return
	}

	conn.contextName = data.Header.Get("Channel-Context")

	conn.ChannelData = Event{
		Message: data,
	}
	conn.Uuid = data.Header.Get("Unique-ID")
	conn.SwitchUuid = data.Header.Get("Core-UUID")
	//conn.call(nil, 0, "divert_events")
	_, err = conn.call(nil, 0, "linger")
	if err != nil {
		logger.Error("Server: connection %s set linger error: %s", conn.Uuid, err.Error())
		return
	}

	_, err = conn.call(nil, 0, "myevents")
	if err != nil {
		logger.Error("Server: connection %s set myevents error: %s", conn.Uuid, err.Error())
		return
	}

	conn.call(nil, 0, "filter Event-Name CHANNEL_HANGUP_COMPLETE")
	conn.call(nil, 0, "filter Event-Name CHANNEL_ANSWER")

	if conn.event == nil {
		conn.event = EventDispatcher()
	}

	defer func() {
		go s.onDisconnect(conn)
	}()

	go s.onConnect(conn)
	//go conn.SndMsg("hangup", "", false, false)

	var recv Message
	for err == nil {
		conn.rw.Lock() // ------------------ < LOCKED >
		recv, err = conn.recv(1e7, true)
		conn.rw.Unlock() // ---------------- < UNLOCK >

		//fmt.Println(err, recv.Header.Get("Event-Name"), recv.Header.Get("Content-Type"))

		if err != nil {
			// Timeout ?
			tmp, ok := err.(net.Error)
			if (ok) && tmp.Timeout() {
				err = nil
			}
		}
		conn.mx.Lock()

		if recv.Header.Get("Content-Type") == "text/disconnect-notice" {
			//			conn.Disconnected = true
			//if len(conn.cbWrapper) == 0 {
			//	fmt.Println(string(recv.Body))
			//}
		}

		if plain, event, err := PlainEvent(recv); err == nil {
			e := Event{Event: plain, Subclass: event, Message: recv, SendData: conn}

			//fmt.Println(e.Header.Get("Event-Name"), e.Header.Get("Application-UUID"), e.Header.Get("Event-UUID"))

			if e.Header.Get("Event-Name") == "CHANNEL_EXECUTE_COMPLETE" {
				conn.ChannelData = e
				if v, ok := conn.cbWrapper[e.Header.Get("Application-UUID")]; ok {
					delete(conn.cbWrapper, e.Header.Get("Application-UUID"))
					v <- e
					close(v)

				} else if v, ok := conn.cbWrapper[e.Header.Get("Event-UUID")]; ok {
					delete(conn.cbWrapper, e.Header.Get("Event-UUID"))
					v <- e
					close(v)
				}
			} else if e.Header.Get("Event-Name") == "CHANNEL_HANGUP_COMPLETE" {
				conn.Disconnected = true // TODO RACE!!!
				conn.ChannelData = e
				for _, v := range conn.cbWrapper {
					close(v)
					//delete(conn.cbWrapper, ch)
				}
				conn.mx.Unlock()
				runtime.Gosched()
				break
			} else {
				conn.event.Fire(Event{Event: plain, Subclass: event, Message: recv, SendData: nil})
			}
		}
		conn.mx.Unlock()
		runtime.Gosched()
	}
	//fmt.Println(conn.Uuid, "->ENDFOR")

	conn.exit <- true

}

func NewServer(addr string, onConnect onConnect, onDisconnect onDisconnect) *Server {
	s := &Server{
		addr:         addr,
		onConnect:    onConnect,
		onDisconnect: onDisconnect,
	}
	return s
}

func (s *Server) Listen() error {
	srv, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	for {
		c, err := srv.Accept()
		if err != nil {
			return err
		}
		var es SConn
		if e := es.connection(c); e != nil {
			fmt.Sprint(e)
		}
		go handle(&es, s)
	}
}
