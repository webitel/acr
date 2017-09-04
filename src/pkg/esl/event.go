package esl

import (
//	"io"
	"fmt"
	"strings"
	"reflect"
	"errors"
)

// Event types enum.
const (

	EventCustom = iota
	EventClone
	EventChannelCreate
	EventChannelDestroy
	EventChannelState
	EventChannelCallState
	EventChannelAnswer
	EventChannelHangup
	EventChannelHangupComplete
	EventChannelExecute
	EventChannelExecuteComplete
	EventChannelHold
	EventChannelUnhold
	EventChannelBridge
	EventChannelUnbridge
	EventChannelProgress
	EventChannelProgressMedia
	EventChannelOutgoing
	EventChannelPark
	EventChannelUnpark
	EventChannelApplication
	EventChannelOriginate
	EventChannelUUID
	EventAPI
	EventLOG
	EventInboundChan
	EventOutboundChan
	EventStartup
	EventShutdown
	EventPublish
	EventUnpublish
	EventTalk
	EventNoTalk
	EventSessionCrash
	EventModuleLoad
	EventModuleUnload
	EventDTMF
	EventMessage
	EventPresenceIn
	EventNotifyIn
	EventPresenceOut
	EventPresenceProbe
	EventMessageWaiting
	EventMessageQuery
	EventRoster
	EventCodec
	EventBackgroundJob
	EventDetectedSpeech
	EventDetectedTone
	EventPrivateCommand
	EventHeartbeat
	EventTrap
	EventAddSchedule
	EventDelSchedule
	EventExeSchedule
	EventReSchedule
	EventReloadXML
	EventNotify
	EventPhoneFeature
	EventPhoneFeatureSubscribe
	EventSendMessage
	EventRecvMessage
	EventRequestParams
	EventChannelData
	EventGeneral
	EventCommand
	EventSessionHeartbeat
	EventClientDisconnect
	EventServerDisconnect
	EventSendInfo
	EventRecvInfo
	EventRecvMessageRTCP
	EventCallSecure
	EventNAT
	EventRecordStart
	EventRecordStop
	EventPlaybackStart
	EventPlaybackStop
	EventCallUpdate
	EventFailure
	EventSocketData
	EventMediaBugStart
	EventMediaBugStop
	EventConferenceDataQuery
	EventConferenceData
	EventCallSetupReq
	EventCallSetupResult
	EventCallDetail
	EventDeviceState

	// pseudo descriptor.
	EventAll
	// DO NOT ADD ANY BELOW
)

var events = [ ]string {
	`CUSTOM`,
	`CLONE`,
	`CHANNEL_CREATE`,
	`CHANNEL_DESTROY`,
	`CHANNEL_STATE`,
	`CHANNEL_CALLSTATE`,
	`CHANNEL_ANSWER`,
	`CHANNEL_HANGUP`,
	`CHANNEL_HANGUP_COMPLETE`,
	`CHANNEL_EXECUTE`,
	`CHANNEL_EXECUTE_COMPLETE`,
	`CHANNEL_HOLD`,
	`CHANNEL_UNHOLD`,
	`CHANNEL_BRIDGE`,
	`CHANNEL_UNBRIDGE`,
	`CHANNEL_PROGRESS`,
	`CHANNEL_PROGRESS_MEDIA`,
	`CHANNEL_OUTGOING`,
	`CHANNEL_PARK`,
	`CHANNEL_UNPARK`,
	`CHANNEL_APPLICATION`,
	`CHANNEL_ORIGINATE`,
	`CHANNEL_UUID`,
	`API`,
	`LOG`,
	`INBOUND_CHAN`,
	`OUTBOUND_CHAN`,
	`STARTUP`,
	`SHUTDOWN`,
	`PUBLISH`,
	`UNPUBLISH`,
	`TALK`,
	`NOTALK`,
	`SESSION_CRASH`,
	`MODULE_LOAD`,
	`MODULE_UNLOAD`,
	`DTMF`,
	`MESSAGE`,
	`PRESENCE_IN`,
	`NOTIFY_IN`,
	`PRESENCE_OUT`,
	`PRESENCE_PROBE`,
	`MESSAGE_WAITING`,
	`MESSAGE_QUERY`,
	`ROSTER`,
	`CODEC`,
	`BACKGROUND_JOB`,
	`DETECTED_SPEECH`,
	`DETECTED_TONE`,
	`PRIVATE_COMMAND`,
	`HEARTBEAT`,
	`TRAP`,
	`ADD_SCHEDULE`,
	`DEL_SCHEDULE`,
	`EXE_SCHEDULE`,
	`RE_SCHEDULE`,
	`RELOADXML`,
	`NOTIFY`,
	`PHONE_FEATURE`,
	`PHONE_FEATURE_SUBSCRIBE`,
	`SEND_MESSAGE`,
	`RECV_MESSAGE`,
	`REQUEST_PARAMS`,
	`CHANNEL_DATA`,
	`GENERAL`,
	`COMMAND`,
	`SESSION_HEARTBEAT`,
	`CLIENT_DISCONNECTED`,
	`SERVER_DISCONNECTED`,
	`SEND_INFO`,
	`RECV_INFO`,
	`RECV_RTCP_MESSAGE`,
	`CALL_SECURE`,
	`NAT`,
	`RECORD_START`,
	`RECORD_STOP`,
	`PLAYBACK_START`,
	`PLAYBACK_STOP`,
	`CALL_UPDATE`,
	`FAILURE`,
	`SOCKET_DATA`,
	`MEDIA_BUG_START`,
	`MEDIA_BUG_STOP`,
	`CONFERENCE_DATA_QUERY`,
	`CONFERENCE_DATA`,
	`CALL_SETUP_REQ`,
	`CALL_SETUP_RESULT`,
	`CALL_DETAIL`,
	`DEVICE_STATE`,

	`ALL`,
}

// ParseEvent interprets a string s as event name
// and returns either the corresponding integer key
// or error in case of string event name unrecognized.
func ParseEvent( s string ) ( int, error ) {
	if ( s != "" ) {
		for key := 0; ( key < len( events )); ( key )++ {
			if strings.EqualFold( events[ key ], s ) {
				return ( key ), ( nil )
			}
		}
	}
	return ( 0 ), fmt.Errorf( "!Event(%s)", s )
}

// EventName interprets an integer e as event key
// and returns either the corresponding literal
// or empty string in case of event key unknown.
func EventName( e int ) ( string ) {
	if ( 0 <= ( e )) && ( e ) < len( events ) {
		return events[ e ]
	}
	return ( "" )
}

// Event Key.
// type Event struct {
// 	 Key int
// 	 Subclass string
// }

// EventArgs ...
type Event struct {
	 // Event
	 Event    int
	 Subclass string
	 // Data
	 Message
	 // Args
	 SendData interface{ }
	 BindData interface{ }
}

// Handler function delegate.
type Handler func( Event )

// Name returns absolute event-name string.
// Output format: EVENT-NAME[ custom::subclass]
/*func ( e Event ) Name( ) string {
	ename := EventName( e.Key )
	if len( ename ) == 0 { // Error
		return fmt.Sprintf(`!Event[%+d]`, int( e.Key ))
	}
	if ( e.Key == EventCustom ) && len( e.Subclass ) > 0 {
		return ( ename + ( " " ) + e.Subclass )
	}
	return ename
}*/

// Headers that Event.Plain handles itself and should be skipped.
var eventWriteExcludeHeader = map[ string ] bool {
	`Event-Name`:		true,
	`Event-Subclass`:	true,
}

// Plain writes text/plain event message to the underlying io.Writer. 
/*func ( e Event ) Plain( w io.Writer ) ( int, error ) {
	// Event-Name: ...
	// Event-Subclass: ...
	// e.Header.WriteSubset( w, eventWriteExcludeHeader ) 
	return e.Message.Plain( w )
}*/

// -------------------------------------------------------- //
//						  Dispatch							//
// -------------------------------------------------------- //

var ErrEventType = errors.New( "unknown event type" )
var ErrEventCustom = errors.New( "event subclass missing" )
var ErrEventHandler = errors.New( "event handler missing" )
var ErrEventSubclass = errors.New( "event subclass invalid" )

// Binding node.
type Binding struct {
	*handler
	 next *Binding
}

type handler struct {
	 // protected
	 cid string
	 event int
	 subclass string
	 filter Header
	 // private
	 data interface{}
	 callback Handler
}

// Next binding.
func ( node *Binding ) Next( ) *Binding {
	if ( node == nil ) {
		return ( nil )
	}
	return ( node.next )
}

// CID as [C]onsumer [ID]entifier string.
func ( node *Binding ) CID( ) string {
	return ( node.cid )
}

// Filter event like ...
func ( node *Binding ) Filter( ) Header {
	return ( node.filter )
}

// Event (custom::subclass) .
func ( node *Binding ) Event( ) ( int, string ) {
	return ( node.event ), ( node.subclass )
}



// Match reports whether bind falls under event.
func ( node *handler ) match( e Event ) bool {
	
	match := ( false )

	if ( node.event == EventAll ) {
		match = ( true )
		if ( node.subclass == "" ) {
			return ( match )
		}
	}

	if ( match ) || ( node.event == e.Event ) {
		if ( node.subclass == "" ) {
			return ( true )
		}
		if ( e.Subclass == "" ) {
			return ( false )
		}
		
		match = ( false )
		if ( strings.HasPrefix( node.subclass, "file:" )) {
			match = ( node.subclass[ :5 ] == e.Header.Get( "Event-Calling-File" ))
		
		} else if ( strings.HasPrefix( node.subclass, "func:" )) {
			match = ( node.subclass[ :5 ] == e.Header.Get( "Event-Calling-Function" ))
		
		} else /* Match subclass name! */ {
			match = ( node.subclass == e.Subclass )
		}

		if ( match ) && len( node.filter ) > ( 0 ) {
			// https://freeswitch.org/stash/projects/FS/repos/freeswitch/browse/src/mod/event_handlers/mod_event_socket/mod_event_socket.c#308
		}
	}

	return ( match )
}

// Notify safe invokes handler if subscription matche.
func ( node *handler ) dispatch( e Event )  {
	if node.match( e ) {
		// make safe
		defer func( ) {
			if recover( ) != nil {
				// log: handle error
			}
		}( )
		// attach binding data
		e.BindData = node.data
		// invoke arguments
		( node ).callback( e )
	}
}


// Dispatcher is a standalone event demultiplexer.
type Dispatcher map[ int ]( *Binding )

// EventDispatcher allocates demultiplexer structure.
func EventDispatcher( ) Dispatcher {
	return make( map[ int ]( *Binding ), ( EventAll + 1 ))
}

// Node returns binding iterator for given event key..
func ( demux Dispatcher ) Node( event int ) *Binding {
	node, bind := demux[ event ]
	if ( bind ) && ( node != nil ) {
		return ( node )
	}
	return ( nil )
}

// Bind an event handler( callback ) function.
func ( demux Dispatcher ) Bind( cid string, event int, subclass string, filter Header,
								callback Handler, data interface{ } ) ( interface{ }, error ) {
	// <callback> defined ?
	if ( callback == nil ) {
		return nil, ErrEventHandler
	}
	// <event> recognized ?
	if ( event < EventCustom ) || ( EventAll < event ) {
		return nil, ErrEventType
	}
	// custom <subclass> ? 
	if ( event == EventCustom ) && ( subclass == "" ) {
		return nil, ErrEventCustom
	} 
	// <subclass> custom ?
	if ( subclass != "" ) && ( event != EventCustom ) {
		return nil, ErrEventSubclass // use filter instead
	}

	hash := reflect.ValueOf( callback ).Pointer( )

	var node *Binding
	// Lookup for bindings node / tail ...
	for node = demux[ event ]; ( node != nil ) && ( node.next != nil ); node = node.next {
		match := reflect.ValueOf( node.callback ).Pointer( )
		if ( hash ) == ( match ) {
			break ;
		}
	}

	entry := &Binding{ &handler{ cid, event, subclass, filter, data, callback }, nil }

	// insert subsequent ...
	if ( node != nil ) {
		entry.next, node.next = node.next, entry
		return ( entry ), ( nil )
	}

	// insert elementary ...
	demux[ event ] = entry
	
	// status: success
	return ( entry ), ( nil )
}

// Unbind an event handler. ( either: binding -or- EventHandler )
func ( demux Dispatcher ) Unbind( this interface{ } ) *Binding {
	
	if ( this == nil ) {
		return ( nil )
	}

	if callback, ok := this.( Handler ); ( ok ) {
		return demux.UnbindHandler( callback ) 
	}
	
	that, ok := this.( *handler )
	
	if ( !ok ) {
		return ( nil )
	}

	
	entry := demux[ that.event ]
	// Thru bind event[node]
	var node *Binding
	for ( entry != nil ) {
		// That binding entry ?
		if ( entry.handler ) == ( that ) {
			//	Cut !
			if ( node != nil ) {
				node.next = entry.next
			//	Shift !
			} else if ( entry.next != nil ) {
				demux[ that.event ] = entry.next
			//	Delete !
			} else {
			//	Single elimination !
				delete( demux, that.event )
			}
			// single elimination ...
			entry.next = nil
			entry.data = nil
			entry.callback = nil

			return ( entry )
		}
		// Thru binding(s)
		node, entry = entry, entry.next
	}

	return ( nil )
}

// UnbindHandler callback function all node(s).
func ( demux Dispatcher ) UnbindHandler( callback Handler ) *Binding {
	// Function defined ?
	if ( callback == nil ) {
		return ( nil )
	}

	hash := reflect.ValueOf( callback ).Pointer( )

	var node, this, free *Binding
	// Thru binding events
	for ekey, entry := range demux {
		
		node = nil
		// Thru binding entries
		for ( entry != nil ) {
			// Callback function(s) match ?
			match := reflect.ValueOf( entry.callback ).Pointer( )
			if ( hash ) != ( match ) {
				node, entry = entry, entry.next
				continue
			}

			// Cut !
			if ( node != nil ) {
				node.next = entry.next
			// Shift !
			} else if ( entry.next != nil ) {
				demux[ ekey ] = entry.next
			// Destroy !
			} else {
			// Single elimination !
				delete( demux, ekey )
			}

			this, entry = entry, entry.next

			this.next = nil
			this.data = nil
			this.callback = nil

			// Resume freed list
			if ( free != nil ) {
				free.next = this
			} else {
				free = this
			}
		}
	}

	return ( free )
}

// Clear all bindings ...
func ( demux Dispatcher ) Clear( ) {
	for ekey, node := range demux {
		delete( demux, ekey )
		
		var entry *Binding
		for ( node != nil ) {
			entry, node = node, node.next
			entry.next = nil
			entry.data = nil
			entry.callback = nil
		}
	}
}

// Fire an event.
func ( demux Dispatcher ) Fire( e Event ) {
	for ekey := e.Event ;; ekey = EventAll {
		
		for node := demux[ ekey ];
			node != nil; node = node.next {
		//	Notify subscriber.
			node.handler.dispatch( e )
		}

		if ( ekey == EventAll ) {
			break;
		}
	}
}

// -------------------------------------------------------- //
//						   Header							//
// -------------------------------------------------------- //

// Interns common header string(s).
var commonHeader = make( map[ string ] string )
var channelHeader = make( map[ string ] string )

func init( ) {
	
	for _, key := range [ ]string {
		
		// Event base ...
		`Event-Subclass`,
		`Event-Name`,
		`Core-UUID`,
		`FreeSWITCH-Hostname`,
		`FreeSWITCH-Switchname`,
		`FreeSWITCH-IPv4`,
		`FreeSWITCH-IPv6`,
		
		`Event-Date-Local`,
		`Event-Date-GMT`,
		`Event-Date-Timestamp`,
		`Event-Calling-File`,
		`Event-Calling-Function`,
		`Event-Calling-Line-Number`,
		`Event-Sequence`,

		// Socket data ...
		`Content-Length`,
		`Content-Type`,
		`Reply-Text`,
		`Job-UUID`,
		
		`Log-Level`,
		`Text-Channel`,
		`Log-File`,
		`Log-Func`,
		`Log-Line`,
		`User-Data`,
		
		
		`Allowed-Events`,
		`Allowed-LOG`,
		`Allowed-API`,

		`Socket-Mode`,				// [async] | [static]
		`Control`,					// [full] | [single-channel]
		`Controlled-Session-UUID`,	// <unique-id>
		`Content-Disposition`,		// [disconnect] | [linger]
		`Channel-Name`,
		`Linger-Time`,
	
	} {
		commonHeader[ key ] = key
	}
	// Channel event header(s)
	for _, key := range [ ]string {
				
		`Channel-State`,
		`Channel-Call-State`,
		`Channel-State-Number`,
		`Channel-Name`,
		`Unique-ID`,
		`Call-Direction`,
		`Presence-Call-Direction`,
		`Channel-HIT-Dialplan`,

		`Channel-Presence-ID`,
		`Channel-Presence-Data`,
		`Presence-Data-Cols`,

		`Channel-Call-UUID`,
		`Answer-State`,
		`Hangup-Cause`,

		`Channel-Read-Codec-Name`,
		`Channel-Read-Codec-Rate`,
		`Channel-Read-Codec-Bit-Rate`,

		`Channel-Write-Codec-Name`,
		`Channel-Write-Codec-Rate`,
		`Channel-Write-Codec-Bit-Rate`,

		`Other-Type`,

		// `Other-Leg`, // prefix

	}{
		channelHeader[ key ] = key
	}
}