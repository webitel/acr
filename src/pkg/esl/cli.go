package esl

import (
	"io"
	"fmt"
	"errors"
	"strings"
	"strconv"
	"unicode/utf8"
)

var lineFeed = [ ]byte{ '\n' }

// AppendLine data ...
func AppendLine( line []byte, data string ) []byte {
	return line
}

// CommandLine appends Go string literal(s),
// representing [cmd( 'arg')...], to the line and returns the extended buffer.
// https://wiki.freeswitch.org/wiki/Event_Socket_Library#Quoting_and_Escaping
func CommandLine( line []byte, cmd ...string ) [ ]byte {
	for i := 0; i < len( cmd ); i++ {
		// Separate argument(s) !
		if len( line ) > 0 && line[ len( line )-1 ] != ( ' ' ) {
			line = append( line, ' ' )
		}
		line = commandLine( line, cmd[ i ], len( line ) > 0 )
	}
	return line
}

// CommandData [ 'arg']...
func CommandData( line []byte, arg ...string ) [ ]byte {
	for i := 0; i < len( arg ); i++ {
		// Separate argument(s) !
		if len( line ) > 0 && line[ len( line )-1 ] != ' ' {
			line = append( line, ' ' )
		}
		line = commandLine( line, arg[ i ], true )
	}
	return ( line )
}

// commandLine ...
// regex test1234|\d                  <== Returns "true"
// regex m:/test1234/\d               <== Returns "true"
// regex m:~test1234~\d               <== Returns "true"
// regex test|\d                      <== Returns "false"
// regex test1234|(\d+)|$1            <== Returns "1234"
// regex sip:foo@bar.baz|^sip:(.*)|$1 <== Returns "foo@bar.baz"
// regex testingonetwo|(\d+)|$1       <== Returns "testingonetwo" (no match)
// regex m:~30~/^(10|20|40)$/~$1      <== Returns "30" (no match)
// regex m:~30~/^(10|20|40)$/~$1~n    <== Returns "" (no match)
// regex m:~30~/^(10|20|40)$/~$1~b    <== Returns "false" (no match)
func commandLine( line []byte, data string, quote bool ) []byte {
	// Empty data ?
	if len( data ) == ( 0 ) {
		if !( quote ) { // Nothing to do
			return line
		}
	}
	// Argument ?
	if ( !quote ) {
		data = strings.TrimSpace( data )
	} else if len( data ) > ( 0 ) {
		quote = ( -1 < strings.IndexByte( data, ' ' ))
	}

	if ( quote ) { // Begin argument !
		line = append( line, '\'' )
	}

	var r rune
	for off := 0; len( data ) > 0; data = data[off:] {
		r, off = rune( data[0] ), 1
		if r >= utf8.RuneSelf {
			r, off = utf8.DecodeRuneInString(data)
		}
		if ( off == 1 ) && ( r == utf8.RuneError ) {
			line = append( line, `\x`...)
			line = append( line, lowerhex[ data[0]>>4 ])
			line = append( line, lowerhex[ data[0]&0xF ])
			continue
		}
		line = commandRune( line, r, quote )
	}

	if ( quote ) { // End argument !
		line = append( line, '\'' )
	}
	
	return line
}

const lowerhex = "0123456789abcdef"

func commandRune( line []byte, r rune, escape bool ) []byte {
	var cbyte [ utf8.UTFMax ]byte
	
	if ( escape ) && ( r == '\\' || r == '\'' ) { // always backslashed
		line = append( line, '\\' )
		line = append( line, byte( r ))
		return line
	}
	if ( r < utf8.RuneSelf ) && strconv.IsPrint( r ) {
		// if escape && r == ' ' {
		// 	cmd = append( cmd, `\s`...)
		// 	return cmd
		// }
		line = append( line, byte( r ))
		return line
	
	} else if strconv.IsPrint( r ) {
		size := utf8.EncodeRune( cbyte[:], r )
		line = append( line, cbyte[ :size ]...)
		return line
	}
	switch ( r ) {
	
	case '\a': 	line = append( line, `\a`...)
	case '\b': 	line = append( line, `\b`...)
	case '\f': 	line = append( line, `\f`...)
	case '\n': 	line = append( line, `\n`...)
	case '\r': 	line = append( line, `\r`...)
	case '\t': 	line = append( line, `\t`...)
	case '\v': 	line = append( line, `\v`...)
	
	default: 	switch {
				case ( r < ' ' ):
					line = append( line, `\x`...)
					line = append( line, lowerhex[ byte( r )>>4 ])
					line = append( line, lowerhex[ byte( r )&0xF ])
				case ( r > utf8.MaxRune ):
					r = 0xFFFD
					fallthrough
				case ( r < 0x10000 ):
					line = append( line, `\u`...)
					for s := 12; s >= 0; s -= 4 {
						line = append( line, lowerhex[ r>>uint( s )&0xF ])
					}
				default:
					line = append( line, `\U`...)
					for s := 28; s >= 0; s -= 4 {
						line = append( line, lowerhex[ r>>uint( s )&0xF ])
					}
				}
	}
	return line
}

// Headers that Request.Write handles itself and should be skipped.
var reqWriteExcludeHeader = map[string]bool {
	"Command":				true,
	"Content-Length":		true,
}

func commandRequest( wr io.Writer, cmd []byte, req Message ) ( n int64, err error ) {
	
	if len( cmd ) == ( 0 ) {
		// Try to determine command from request.
		cmd = CommandLine( cmd, req.Header[`Command`]...)
	}

	if len( cmd ) == ( 0 ) {
		return 0, errors.New(`esl: request command is missing`)
	}

	var w int
	// command(+arg(s))\n
	w, err = fmt.Fprintf( wr, "%s\n", cmd )
	( n ) += int64( w )
	if ( err != nil ) {
		return
	}

	// [request: header\n]
	w, err = req.Header.WriteSubset( wr, reqWriteExcludeHeader )
	( n ) += int64( w )
	if ( err != nil ) {
		return
	}

	// [Content-Length]
	cLength := len( req.Body )
	if ( cLength ) > ( 0 ) {
		
		w, err = fmt.Fprintf( wr, "Content-Length: %s\n",
					strconv.FormatInt( int64( cLength ), 10 ))
		
		( n ) += int64( w )
		if ( err != nil ) {
			return
		}
	}

	// \n
	w, err = wr.Write( lineFeed )
	( n ) += int64( w )
	if ( err != nil ) {
		return
	}

	// [request:body]
	if ( cLength ) > ( 0 ) {
		
		w, err = wr.Write( req.Body )
		( n ) += int64( w )
		if ( err != nil ) {
			return
		}
		// (lineFeed) \n\n
		for i := ( 0 ); ( i < 2 ); ( i )++ {
			
			if ( cLength > i ) && req.Body[ cLength -i -1 ] == ( '\n' ) {
				continue
			}

			w, err = wr.Write( lineFeed )
			( n ) += int64( w )
			if ( err != nil ) {
				return
			}
		}
	}

	return
}