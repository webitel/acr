package esl

import (
	"io"
	"fmt"
	"sort"
	"sync"
)

// A Header respresents name:[]value storage.
type Header map[ string ][ ]string

// GetValue safe gets the value associated with the given key and index i.
// If there is no value associated with the key, GetValue returns "", false.
// To access multiple values of a key, access the map directly.
func ( h Header ) GetValue( key string, i int ) ( value string, ok bool ) {
	varr := h[ key ]
	// return textproto.MIMEHeader(h).Get(key)
	if len( varr ) > 0 {
		// [I]ndicated ?
		if ( i > -1 ) {
			if ( i < len( varr )) {
				return varr[ i ], true
			}
			return "", false
		}
		// Single value ?
		if len( varr ) == 1 {
			return varr[ 0 ], true
		}
		// Multi value !
		var n int
		for i = 0; i < len( varr ); i++ {
			n += len( varr[ i ]) + 2 // separate
		}

		bval := make([ ]byte, n + 7 )
		
		n = copy( bval[0:], "ARRAY::" )
		n += copy( bval[n:], varr[ 0 ])
		
		for i = 1; i < len( varr ); i++ {
			n += copy( bval[n:], "|:" )
			n += copy( bval[n:], varr[i] )
		}

		return string( bval[:n] ), true
	}
	return // "", false
}

func ( h Header ) DelValue( key string, i int ) ( value string, ok bool ) {
	varr := h[ key ]
	// return textproto.MIMEHeader(h).Get(key)
	if len( varr ) > 0 {
		// [I]ndicated ?
		if ( i > -1 ) {
			if ( i < len( varr )) {
				return varr[ i ], true
			}
			return "", false
		}
		// Single value ?
		if len( varr ) == 1 {
			return varr[ 0 ], true
		}
		// Multi value !
		var n int
		for i = 0; i < len( varr ); i++ {
			n += len( varr[ i ]) + 2 // separate
		}

		bval := make([ ]byte, n + 7 )
		
		n = copy( bval[0:], "ARRAY::" )
		n += copy( bval[n:], varr[ 0 ])
		
		for i = 1; i < len( varr ); i++ {
			n += copy( bval[n:], "|:" )
			n += copy( bval[n:], varr[i] )
		}

		return string( bval[:n] ), true
	}
	return // "", false
}


// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
func ( h Header ) Add( name string, value ...string ) {
	// textproto.MIMEHeader(h).Add(key, value)
	h[ name ] = append( h[ name ], value... )
}

// Set sets the header entries associated with key to
// the single element value. It replaces any existing
// values associated with key.
func ( h Header ) Set( name string, value ...string ) {
	// textproto.MIMEHeader(h).Set(key, value)
	if ( value == nil ) {
		delete( h, name ) // h.Del( name )
		return
	}
	
	h[ name ] = ( value )
}

// Exists reports whether h store key name.
func ( h Header ) Exists( name string ) bool {
	_, exists := h[ name ]
	return exists
}

// Get gets value string associated with the given key.
// If there are no values associated with the key, Get returns "".
// To access multiple values of a key, access the map directly
// with CanonicalHeaderKey.
func ( h Header ) Get( key string ) string {
	// return textproto.MIMEHeader(h).Get(key)
	v, _ := h.GetValue( key, -1 )
	return v
}

// Del deletes the values associated with key.
func ( h Header ) Del( key string ) {
	// textproto.MIMEHeader(h).Del(key)
	delete( h, key )
}

// Clone returns copy of Header h.
func ( h Header ) Clone( ) Header {
	// Avoid lots of small slice allocations later by allocating one
	// large one ahead of time which we'll cut up into smaller
	// slices. If this isn't big enough later, we allocate small ones.	
	var hint = len( h )
	var stmp []string
	
	if ( hint > 0 ) {
		stmp = make( []string, hint )
	}

	hn := make( Header, hint )
	
	for key, value := range h {
		var array []string
		
		if n := len( value ); n > 0 {
			if len( stmp ) >= n {
				// More than likely this will be a single-element key.
				// Most headers aren't multi-valued.
				// Set the capacity on strs[0] to 1, so any future append
				// won't extend the slice into the other strings.
				array, stmp = stmp[:n:n], stmp[n:]
			} else {
				array = make([ ]string, n )
			}
			copy( array, value )
		}
		hn[ key ] = array
	}

	return hn
}








// var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")
/*
type writeStringer interface {
	WriteString(string) (int, error)
}

// stringWriter implements WriteString on a io.Writer.
type stringWriter struct {
	wr io.Writer
}

func (w stringWriter) WriteString(s string) (n int, err error) {
	return io.WriteString(w.wr, s)
}*/

type header struct {
	 name	string
	 value  []string
}

// A headerSorter implements sort.Interface by sorting a []Attribute
// by key. It's used as a pointer, so it can fit in a sort.Interface
// interface value without allocation.
type headerSorter struct {
	 arr []header
}

func ( h *headerSorter ) Len( ) int				{ return len( h.arr ) }
func ( h *headerSorter ) Swap( i, j int )		{ h.arr[ i ], h.arr[ j ] = h.arr[ j ], h.arr[ i ] }
func ( h *headerSorter ) Less( i, j int ) bool	{ return h.arr[ i ].name < h.arr[ j ].name }

var headerPool = sync.Pool {
	New: func( ) interface{ } {
		 return new( headerSorter )
	},
}

// sortedKeyValues returns h's keys sorted in the returned kvs
// slice. The headerSorter used to sort is also returned, for possible
// return to headerSorterCache.
func ( h Header ) sort( exclude map[ string ] bool ) ( harr [ ]header, hsrt *headerSorter ) {
	hsrt = headerPool.Get( ).( *headerSorter )
	if cap( hsrt.arr ) < len( h ) {
		hsrt.arr = make( []header, 0, len( h ))
	}
	harr = hsrt.arr[:0]
	for name, varr := range h {
		if !exclude[ name ] {
			harr = append( harr, header{ name, varr })
		}
	}
	hsrt.arr = harr
	sort.Sort( hsrt )
	return harr, hsrt
}

// WriteSubset writes a header in wire format.
// If exclude is not nil, keys where exclude[key] == true are not written.
func ( h Header ) WriteSubset( w io.Writer, exclude map[ string] bool ) ( n int, err error ) {
	// ws, ok := w.(writeStringer); if !ok {
	// 	ws = stringWriter{ w }
	// }

	header, sorter := h.sort( exclude )
	defer headerPool.Put( sorter )
	
	for _, kvh := range header {
		
		c, err := WriteHeader( w, kvh.name, kvh.value...)
		
		( n ) += int( c )
		if ( err != nil ) {
			return n, err
		}
	}

	return // n, err
}

// func WriteHeader( w io.Writer, h Header, contentLength int ) ( n int, err error ) {

// }

func WriteHeader( w io.Writer, name string, value ...string ) ( n int, err error ) {
	if len( name ) > 0 {
		if len( value ) > 0 {
			if len( value ) == ( 1 ) && len( value[ 0 ]) > ( 0 ) {
				n, err = fmt.Fprintf( w, "%s: %s", name, value[ 0 ])
			} else {
				wn, err := fmt.Fprintf( w, "%s: ARRAY::%s", name, value[ 0 ])
				if n += wn; err != nil {
					return n, err
				}
				
				for i := 1; i < len( value ); i++ {
					wn, err := fmt.Fprintf( w, "|:%s", value[ i ])
					if n += wn; err != nil {
						return n, err
					}
				}
			}
			
			wn, err := w.Write( lineFeed )
			if n += wn; err != nil {
				return n, err
			}
		}
	}
	return
}