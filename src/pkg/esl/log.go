package esl

import (
	"fmt"
	"errors"
	"strings"
	"strconv"
	"reflect"
)

// Log level enum.
const (

	LogConsole = iota
	LogAlert
	LogCritical
	LogError
	LogWarning
	LogNotice
	LogInfo
	LogDebug
)

// Log level name.
var levels = [ ]string {
	`CONSOLE`,
	`ALERT`,
	`CRIT`,
	`ERR`,
	`WARNING`,
	`NOTICE`,
	`INFO`,
	`DEBUG`,
}

// LogArgs ...
type LogArgs struct {
	 Level int
	 Message
	 SendData interface{}
	 BindData interface{}
}

// LogFunc function.
type LogFunc func( LogArgs )

// ParseLogLevel interprets a string s in the log level integer key. 
func ParseLogLevel( s string ) ( int, error ) {

	level, err := strconv.Atoi( s )

	if ( err == nil ) {
		if ( level > LogDebug ) {
			return LogDebug, nil
		} else if ( level < 0 ) {
			return 0, nil
		}
		return level, nil
	}

	for x := 0; x < len( levels ); x++ {
		if strings.EqualFold( levels[x], s ) {
			return x, nil
		}
	}

	return 0, fmt.Errorf( "log[%s] level unknown", s )
}

// LogDispatcher ...
type LogDispatcher struct {
	 level int
	 node *bindingLog
}

type bindingLog struct {
	 level int
	 logF LogFunc
	 next *bindingLog
}

// Node returns current binding(s) level.
func ( demux *LogDispatcher ) Node( ) ( bind bool, level int ) {
	return ( demux.node != nil ), ( demux.level )
}

// Bind logger function handler.
func ( demux *LogDispatcher ) Bind( logger LogFunc, level int ) error {
	
	if ( logger == nil ) {
		return errors.New( "log func missing" )
	}

	if ( level > LogDebug ) {
		level = ( LogDebug )
	} else if ( level < 0 ) {
		level = ( 0 )
	}

	hash := reflect.ValueOf( logger ).Pointer( )

	var node, entry *bindingLog
	// Lookup for binding tail ...
	for node = demux.node; ( node != nil ) &&
		( node.next != nil ); node = node.next {
		// Binding logger entry ?
		match := reflect.ValueOf( node.logF ).Pointer( )
		
		if ( hash ) == ( match ) {
			( entry ) = ( node )
			break // Already binded
		}
	}

	if ( entry ) != ( nil ) {
		entry.level = level
	} else {
		entry = &bindingLog{ level, logger, nil }
	}

	// insert ...
	if ( node != nil ) {
		entry.next, node.next = node.next, entry
	} else {
		demux.node = entry
	}
	// [/log] updating ...
	if ( demux.level < level ) {
		demux.level = level
	}

	return ( nil )
}

// Unbind logger function handler. Indicates success and limit of binding level.
func ( demux *LogDispatcher ) Unbind( logger LogFunc ) ( ok bool ) {
	
	if ( logger == nil ) {
		return
	}

	var level int
	var node, free *bindingLog
	
	hash := reflect.ValueOf( logger ).Pointer( )
	
	entry := demux.node
	for ( entry != nil ) {
		// Lookup for binding entry ...
		match := reflect.ValueOf( entry.logF ).Pointer( )
		
		if ( hash ) == ( match ) {
			
			if ( node != nil ) {
				node.next = entry.next
			} else {
				demux.node = entry.next
			}
			free, entry = entry, entry.next
			
			free.logF = nil
			ok = ( true )
			continue
		}
		
		if ( level < node.level ) {
			level = node.level
		}
		// iterate next ...
		node, entry = entry, entry.next
	}
	// [/log] updating ...
	if ( demux.level != level ) {
		demux.level = level
	}

	return
}

// Dispatch message to binding node(s).
func ( demux *LogDispatcher ) Dispatch( msg Message, level int ) {
	for node := demux.node; ( node != nil ); node = node.next {
		if ( node.level >= level ) {
			// TODO: invoke safe !
			node.logF( LogArgs{ level, msg, nil, nil })
		}
	}
}