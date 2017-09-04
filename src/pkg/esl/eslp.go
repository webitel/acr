package esl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrConnected = errors.New("connection active")
var ErrDisconnected = errors.New("disconnected")

// EventSocket connection.
type EventSocket struct {
	// Execute duration limit.
	Timeout time.Duration
	// protect following
	rw   sync.Mutex
	conn net.Conn

	cbuf   []byte
	ramq   *message
	rbuf   *Reader
	wbuf   *bufio.Writer
	opened bool

	hostname string
	cmd      []byte
	amq      *message
	// subscription
	log    LogDispatcher
	event  Dispatcher
	custom map[string](int)
}

type message struct {
	Message
	next *message
}

// Close disconnect the connection.
func (esl *EventSocket) Close() error {

	esl.rw.Lock()
	defer esl.rw.Unlock()

	if esl.conn != nil {
		err := esl.conn.Close()
		esl.conn = nil
		return (err)
	}

	return nil
}

// Open authorize the network connection.
func (esl *EventSocket) Open(conn net.Conn, username, password string) error {

	esl.rw.Lock()
	defer esl.rw.Unlock()

	err := esl.connection(conn)
	if err != nil {
		return (err)
	}

	// Waiting [auth/request]
	recv, err := esl.recv(5e9, false)

	if err != nil {
		esl.disconnect(err) // disconnect
		return (err)
	}

	if recv.Header.Get(`Content-Type`) != (`auth/request`) {
		err = errors.New(`esl: connection error`)
		esl.disconnect(err) // disconnect
		return (err)
	}

	if strings.ToLower(username) == `auth` {
		(username) = ""
	}

	//	defaults
	if (username == "") && (password == "") {
		(password) = `ClueCon`
	}

	if username == "" {
		recv, err = esl.call(nil, 5e9, `auth`, password)
	} else {
		recv, err = esl.call(nil, 5e9, `userauth`, username+`:`+password)
	}

	if err != nil {
		err = errors.New(`esl: connection error`)
		esl.disconnect(err) // disconnect
		return (err)
	}

	if recv.Header.Get(`Reply-Text`) != (`+OK accepted`) {
		err = errors.New(`esl: authentication error`)
		esl.disconnect(err) // disconnect
		return (err)
	}
	// debug purpose ...
	recv, err = esl.call(nil, 5e9, `api`, `hostname`)
	if (err == nil) && len(recv.Body) > (0) {
		esl.hostname = string(recv.Body)
	}

	return (nil)
}

// Receive waits for upcomming message from network connection limited time duration.
// Reports whether at least one more package is available -or- protocol error while reading.
func (esl *EventSocket) Receive(timeout time.Duration) (ack bool, err error) {

	esl.rw.Lock() // ------------------ < LOCKED >
	recv, err := esl.recv(timeout, true)
	esl.rw.Unlock() // ---------------- < UNLOCK >

	if err != nil {
		// Timeout ?
		tmp, ok := err.(net.Error)
		if (ok) && tmp.Timeout() {
			// receive time out !
			return (false), nil
		}
		// disconnect
		return (false), (err)
	}

	// Content-Type: text/event-plain
	if plain, event, err := PlainEvent(recv); err == nil {
		esl.event.Fire(Event{Event: plain, Subclass: event, Message: recv, SendData: esl})
		return (true), (nil)
	}

	// Content-Type: log/data
	// if ( recv.Header.Get(`Content-Type`) == `log/data` ) { }

	fmt.Printf("esl/recv: %#v\n", recv)
	panic("Content-Type: unrecognized")
}

// Routine the connection ...
func (esl *EventSocket) Routine(ctx context.Context) error {

	if ctx == nil {
		return errors.New("routine: context <nil>")
	}
	//if ( ctx.Done( ) == nil ) {
	//	return errors.New( "routine: infinite runtime context" )
	//}

	var ack bool
	var err error

	for err == nil {
		select {
		// cancel ?
		case <-ctx.Done():
			// return ctx.Err( )
			return esl.Close()
		// receive !
		default:

			ack, err = esl.Receive(1e8) // 100ms
			for (ack) && (err == nil) {
				ack, err = esl.Receive(1e8)
			}
			if err != nil {
				break
			}
			runtime.Gosched()
		}
	}

	return (err)
}

// LocalAddr returns the local network address.
func (esl *EventSocket) LocalAddr() net.Addr {
	esl.rw.Lock()
	defer esl.rw.Unlock()
	if esl.conn != nil {
		return esl.conn.LocalAddr()
	}
	return (nil)
}

// RemoteAddr returns the remote network address.
func (esl *EventSocket) RemoteAddr() net.Addr {
	esl.rw.Lock()
	defer esl.rw.Unlock()
	if esl.conn != nil {
		return esl.conn.RemoteAddr()
	}
	return (nil)
}

// SetTimeout specifies the maximum amount of time
// the connection will wait for future API calls to complete.
// Negative -or- zero value, a default timeout of 5s is used.
func (esl *EventSocket) SetTimeout(p time.Duration) error {
	return nil
}

// Log binds logger function for future ESL [log/data] notifications.
func (esl *EventSocket) Log(level int, logger LogFunc) error {

	// Function defined ?
	if logger == nil {
		return ErrEventHandler
	}

	esl.rw.Lock()
	defer esl.rw.Unlock()

	_, hardLevel := esl.log.Node()
	err := esl.log.Bind(logger, level)

	if err != nil {
		return (err)
	}
	// need synchronization ?
	if (level) <= (hardLevel) {
		return (nil)
	}

	resp, err := esl.call(nil, 5e9, `log`, strconv.Itoa(level))
	// DEBUG: local error ?
	if err != nil {
		esl.log.Unbind(logger)
		return (err)
	}
	// DEBUG: remote error ?
	reply := resp.Header.Get(`Reply-Text`)
	if !strings.HasPrefix(reply, `+OK log level `) {
		esl.log.Unbind(logger)
		return errors.New(reply)
	}

	return (nil)
}

// Api ( BLOCKING MODE ) execute( send/receive ) API cmd( arg(s)) request ...
func (esl *EventSocket) Api(cmd string, arg ...string) (resp []byte, err error) {

	cmd = strings.TrimSpace(cmd)
	// Trim explicit command name
	if len(cmd) >= 3 && strings.EqualFold(cmd[:3], "api") {
		if (len(cmd) > 3) && (cmd[3] == ' ') {
			cmd = cmd[4:]
		}
	}

	esl.rw.Lock() // ------------------ < LOCKED >
	recv, err := esl.call(nil, esl.Timeout, (`api ` + cmd), arg...)
	esl.rw.Unlock() // ---------------- < UNLOCK >

	if err != nil {
		return nil, err
	}

	if recv.Header.Get("Content-Type") != ("api/response") {
		return nil, errors.New("eslp: protocol error")
	}
	// assert( len( recv.Body ) > 0, "esl: protocol error")
	return recv.Body, nil
}

// Event bind callback function for future event(subclass) notifications.
// Returns binding interface which can be pass thru Unbind method to unsubscribe exactly this subscription.
func (esl *EventSocket) Event(cid string, event int, subclass string, filter Header, callback Handler, data interface{}) (interface{}, error) {

	// <callback> defined ?
	if callback == nil {
		return nil, ErrEventHandler
	}
	// <event> recognized ?
	if (event < EventCustom) || (EventAll < event) {
		return nil, ErrEventType
	}
	// <subclass> event ?
	if subclass != "" {
		// <subclass> custom ?
		if event != EventCustom {
			return nil, ErrEventSubclass // use filter instead
		}
	}
	// custom <subclass> ?
	if event == EventCustom {
		// canonize <subclass> !
		subclass = strings.ToLower(strings.TrimSpace(subclass))
		// <subclass> name ?
		if subclass == "" {
			return nil, ErrEventSubclass
		}
	}

	esl.rw.Lock()
	defer esl.rw.Unlock()
	// ------------------------------------------------------- < LOCKED >

	// Lazy init !
	if esl.event == nil {
		esl.event = EventDispatcher()
	}

	send := (esl.event[event] == nil)
	// Bind internal structure ...
	node, err := esl.event.Bind(cid, event, subclass, filter, callback, data)

	if err != nil {
		return nil, err
	}

	if event == EventCustom {
		// pre-register subclass state
		send = (esl.custom[subclass] == 0)

		if esl.custom == nil {
			esl.custom = make(map[string](int))
		}
		// register subclass !
		esl.custom[subclass] = (esl.custom[subclass] + 1)
	}

	// Perform !
	if (!send) || ((event < EventAll) &&
		(esl.event[EventAll] != nil)) {
		// No need
		return node, nil
	}

	var resp Message
	//  event <event>[ <subclass>] ...
	if event > EventCustom {
		resp, err = esl.call(nil, 5e9, "event", EventName(event))
	} else {
		resp, err = esl.call(nil, 5e9, "event", EventName(event), subclass)
	}

	// Connection error ?
	if err == nil {
		// Binding error ?
		reply := resp.Header.Get("Reply-Text")
		if !strings.HasPrefix(reply, "+OK event listener enabled plain") {
			err = errors.New(reply)
		}
	}

	if err != nil {
		// rollback !
		esl.event.Unbind(node)
		if event == EventCustom {
			// unregister subclass !
			esl.custom[subclass] = (esl.custom[subclass] + 1)
		}
		// failed
		return nil, err
	}
	// success
	return node, nil
	// ------------------------------------------------------- < UNLOCK >
}

// Unbind previously binded interface ( either: Event() subscription, EventHandler, LogFunc )
func (esl *EventSocket) Unbind(this interface{}) int {

	esl.rw.Lock()
	defer esl.rw.Unlock()
	// Unbind LOG Function
	if logger, ok := this.(LogFunc); ok {
		// pre-operational status ...
		_, level := esl.log.Node()
		// bound, level := esl.log.Node( )
		if esl.log.Unbind(logger) {
			// re-sync [/log] level notification(s).
			bind, limit := esl.log.Node()
			if (level) != (limit) {

				var err error
				var resp Message

				if !bind {
					// Unsubscribe from log notification(s)
					resp, err = esl.call(nil, 5e9, `nolog`)
				} else {
					// Sync remote log notification(s) level.
					resp, err = esl.call(nil, 5e9, `log`, strconv.Itoa(limit))
				}
				// DEBUG: local error ?
				if err != nil {
					// connection error
				}
				reply := resp.Header.Get(`Reply-Text`)
				// DEBUG: remote error ?
				if !strings.HasPrefix(reply, `+OK `) {
					// operation failed
				}
			}
			return (1)
		}
		return (0)
	}

	count := 0
	// Unbind EVENT Handler
	for node := esl.event.Unbind(this); node != nil; node = node.Next() {
		// Sum()
		(count)++
		// Binding args ...
		event, subclass := node.Event()
		// Unregister <event>
		if event != EventCustom {
			if esl.event.Node(event) == (nil) {
				esl.call(nil, 5e9, `nixevent`, EventName(event))
			}
			continue
		}
		// Unregister <subclass>
		custom := esl.custom[subclass]
		switch custom {
		case (0):
			break // internal error
		default:
			esl.custom[subclass] = (custom - 1)
		case (1):
			delete(esl.custom, subclass)
			esl.call(nil, 5e9, `nixevent`, `custom`, subclass)
		}
	}

	return (count)
}

func (esl *EventSocket) disconnect(cause error) {

	// esl.err = err
	//esl.rw.Lock()
	//defer esl.rw.Unlock()

	esl.opened = false
	conn := esl.conn
	if conn != nil {

		conn.Close()
		esl.conn = nil
		// stream detach !
		esl.rbuf.Reset(nil)
		esl.wbuf.Reset(nil)

		// linger ?
		var free *message
		recv := esl.amq
		esl.amq = nil
		for recv != nil {
			free, recv = recv, recv.next
			free.next = nil // avoid memory leaks
		}
	}
}

func (esl *EventSocket) connection(conn net.Conn) error {

	if esl.conn != nil {
		return ErrDisconnected
	}

	if esl.rbuf != nil {
		esl.rbuf.Reset(conn)
		esl.wbuf.Reset(conn)
	} else { // lazy init
		esl.rbuf = PlainReader(conn, 2048) // bufio.NewReaderSize( src, 2048 )
		esl.wbuf = bufio.NewWriterSize(conn, 2048)
	}

	esl.hostname = conn.RemoteAddr().String()
	// esl.err = nil
	esl.conn = conn
	esl.opened = true

	return (nil)
}

// Send command request ...
func (esl *EventSocket) send(req *Message, cmd string, arg ...string) error {

	if esl.conn == nil {
		return ErrDisconnected
	}

	esl.cmd = CommandLine(esl.cmd[:0], cmd)
	esl.cmd = CommandLine(esl.cmd, arg...)

	var params Message
	if req != nil {
		params = *(req)
	}

	_, err := commandRequest(esl.wbuf, esl.cmd, params)

	if err == nil {
		//fmt.Printf("\nesl:send@%s /%s [%d]b.\n", esl.hostname, esl.cmd, esl.wbuf.Buffered())
		err = esl.wbuf.Flush()
	}

	if err != nil {
		// disconnect
	}

	return (err)
}

// Recv message package ...
func (esl *EventSocket) recv(timeout time.Duration, mq bool) (recv Message, err error) {
	// Poll acquired ?
	if (mq) && (esl.amq != nil) {
		// Pop acquired !
		recv = esl.amq.Message
		esl.amq = esl.amq.next

	} else {

		if esl.conn == nil {
			return recv, ErrDisconnected
		}
		// Read connection ...
		if (timeout > 0) && (esl.rbuf.R.Buffered() < 5) {
			// Limit poll period !
			err = esl.conn.SetReadDeadline(time.Now().Add(timeout))

			if err == nil {
				// Poll data available !
				_, err = esl.rbuf.R.Peek(5)
				// Release read timeout !
				esl.conn.SetReadDeadline(time.Time{ /*_zero_*/ })
			}

			if err != nil {
				// Timed out ?
				tmp, ok := err.(net.Error)
				if (!ok) || !(tmp).Timeout() {
					// disconnect: connection error
					esl.disconnect(err)
				}
				return recv, err
			}
			// Data Available !
		}
		// Parse package ...
		recv, err = esl.rbuf.ReadMessage()
		// Explicit error ?
		if err != nil {
			// disconnect: protocol error
			esl.disconnect(err)
			return // msg, err
		}
	}
	// Parse message ...
	//ctype := recv.Header.Get("Content-Type")
	//if ( ctype == "text/disconnect-notice" ) {
	//	dispo := recv.Header.Get( "Content-Disposition" )
	//	if ( true  || len( dispo ) == 0 ) || ( dispo != "linger" ) {
	//		esl.disconnect( nil ) // remote disconnect
	//		return recv, ErrDisconnected
	//	}
	//}
	// debug: event dump
	// fmt.Printf("\neslp/recv: ")
	// recv.WriteTo(os.Stdout)

	return recv, err
}

// Exec command request ...
func (esl *EventSocket) call(req *Message, ttl time.Duration, cmd string, arg ...string) (resp Message, err error) {

	if !esl.opened {
		err = errors.New("Socket closed")
		return
	}

	err = esl.send(req, cmd, arg...)

	if err != nil {
		return // nil, err
	}

	if ttl > 0 {
		// TTL: min(time.Second)
		if ttl < time.Second {
			ttl = time.Second
			// TTL: max(time.Second *15)
		} else if ttl > 15e9 {
			ttl = 15e9
		}
	}

	// perform: recv
	for err == nil {

		resp, err = esl.recv(ttl, false)

		if err == nil {

			ctype := resp.Header.Get("Content-Type")
			if (ctype != "api/response") && (ctype != "command/reply") {

				recv := esl.amq // Find receiv[ed] event(s) tail !
				for (recv != nil) && (recv.next != nil) {
					recv = recv.next
				}

				if recv == nil { // Push receive[d]
					esl.amq = &message{resp, nil}
				} else {
					recv.next = &message{resp, nil}
				}

				resp = Message{ /*_undef_*/ }
				continue // receive
			}

			break // receive[d]
		}
	}

	return // resp, err
}
