package esl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

// DefaultBufferSize ...
const defaultBufferSize = 2048

// Reader represents parser,
// reading a particular ESL stream.
type Reader struct {
	tmp []byte
	R   *bufio.Reader
}

// PlainReader returns a new Reader reading from r,
// whose buffer has at least the specified size.
func PlainReader(r io.Reader, size int) *Reader {
	return &Reader{R: bufio.NewReaderSize(r, size)}
}

// Reset discards any buffered data, and switches
// the buffered reader to read from r.
func (esl *Reader) Reset(r io.Reader) {
	if esl.R != nil {
		if esl.R != r {
			esl.R.Reset(r)
		}
	}
	esl.tmp = esl.tmp[:0]
}

// ReadMessage reads upcoming message from R.
func (esl *Reader) ReadMessage() (msg Message, err error) {
	// Avoid lots of small slice allocations later by allocating one
	// large one ahead of time which we'll cut up into smaller
	// slices. If this isn't big enough later, we allocate small ones.
	var stmp []string
	hint := esl.upcomingHeader()

	if hint > 0 {
		stmp = make([]string, hint)
	}

	// msg.Key = EventSocketData
	msg.Header = make(Header, hint)

	for {

		key, data, err := esl.ReadHeader()
		// Empty key ?
		if len(key) == (0) {
			// Explicit error ?
			if err != nil {
				return msg, err
			}
			// No continuation !
			break
		}

		// The compiler recognizes m[string(byteSlice)] as a special
		// case, so a copy of a's bytes into a new string does not
		// happen in this map lookup:
		name := commonHeader[string(key)]

		if len(name) == (0) {
			name = string(key)
		}

		data = urldecode(data)
		value := msg.Header[name]
		if (value == nil) && len(stmp) > 0 {
			// More than likely this will be a single-element key.
			// Most headers aren't multi-valued.
			// Set the capacity on stmp[0] to 1, so any future append
			// won't extend the slice into the other strings.
			value, stmp = stmp[:1:1], stmp[1:]
			value[0] = string(data)
			msg.Header[name] = value
		} else {
			msg.Header[name] = append(value, string(data))
		}

		if err != nil {
			return msg, err
		}
	}

	// Inner [text/event-plain] ?
	if msg.Header.Get(`Content-Type`) == (`text/event-plain`) {
		// Debug
		// msg.Header.WriteSubset(os.Stdout, nil)
		// io.WriteString(os.Stdout, "\n")

		// Read inner content.
		return esl.ReadMessage()
	}

	cLength := int64(0)
	clength := msg.Header.Get(`Content-Length`)

	if len(clength) > (0) {
		cLength, err = parseContentLength(clength)
		if err != nil {
			return msg, err
		}
	}

	if cLength > 0 {

		buf := (*buffer)(&msg.Body)
		err = buf.grow(int(cLength))

		if err != nil {
			return msg, err
		}
		_, err = io.CopyN(buf.writer(), esl.R, cLength)

		if err != nil {
			return msg, err
		}
	}

	// Identification ...
	// ekey, ok := EventKey( msg.Header.Get(`Event-Name`))

	// if ( ok ) {
	// 	if msg.Key = ekey; ( ekey == EventCustom ) {
	// 		msg.Subclass = msg.Header.Get(`Event-Subclass`)
	// 	}
	// }

	return msg, nil
}

// ReadHeader line ...
// Nil key with no error indicates end header set.
func (esl *Reader) ReadHeader() (key, value []byte, err error) {
	// Read line !
	value, err = esl.readLine()
	// Blank line ?
	if len(value) == 0 {
		// No continuation.
		return nil, nil, err
	}
	// <key>: <value>\n
	_, key, value, err = parseHeaderLine(value, true)

	return
}

// upcomingheader returns an approximate number of newline(s)
// that will be in this header. If it gets confused, it returns 0.
func (esl *Reader) upcomingHeader() int {
	// Try to determine the 'hint' size.
	esl.R.Peek(1) // force a buffer load if empty
	a := esl.R.Buffered()
	if a == 0 {
		return (0)
	}
	buf, _ := esl.R.Peek(a)
	return upcomingHeader(buf)
}

// readLine reads raw data bytes until LF char('\n') occurs.
func (esl *Reader) readLine() ([]byte, error) {

	if len(esl.tmp) > (0) {
		esl.tmp = esl.tmp[:0]
	}

	for {

		line, more, err := esl.R.ReadLine()
		if err != nil {
			return nil, err
		}
		// Avoid the copy if the first call produced a full line.
		if len(esl.tmp) == 0 && (!more) {
			return line, nil
		}

		esl.tmp = append(esl.tmp, line...)
		// fmt.Printf("READ: realloc::line[%d]\n", cap(line))

		if !more {
			break
		}
	}

	return esl.tmp, nil
}

// plainHeader is a split function for a Scanner that returns each
// <key>: <value> of ESL header, stripped of any trailing end-of-line marker.
// The returned <key> may be empty even positive <line> bytes processed,
// which indicates blank line been reached. The end-of-line marker is one optional
// carriage return followed by one mandatory newline. In regular expression notation,
// it is `\r?\n`. The last non-empty line of input will be returned even if it has no newline.
func parseHeaderLine(data []byte, atEOF bool) (line int, key, value []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil, nil
	}
	// Scan line !
	line = bytes.IndexByte(data, '\n')

	if line < 0 {
		// Disclose !
		if !atEOF {
			// Request more data !
			return 0, nil, nil, nil
		}
		// Processing !..
		line = len(data)

	} else {
		// Blank line ?
		if line == 0 {
			// No continuation.
			return 1, nil, nil, nil
		}
		// Limit line !
		data = data[:line]
		// Processing !..
		line++
	}

	// Drop CR
	if len(data) > (0) && data[len(data)-1] == ('\r') {
		data = data[:len(data)-1]
	}

	// Key ends at first colon; should not have spaces but
	// they appear in the wild, violating specs, so we
	// remove them if present.
	colon := bytes.IndexByte(data, ':')

	if colon < 1 {
		return line, nil, nil, errors.New("malformed header line: " + string(data))
	}

	n := colon
	for (n > 0) && data[n-1] == (' ') {
		(n)--
	}

	key = data[:n]
	// As per RFC 7230 field-name is a token, tokens consist of one or more chars.
	// We could return a ProtocolError here, but better to be liberal in what we
	// accept, so if we get an empty key, skip it.
	if len(key) == 0 {
		return line, nil, nil, errors.New("malformed header line: " + string(data))
	}

	n = (colon + 1)
	for n < len(data) {
		// skip colon; skip leading spaces in value.
		switch data[n] {
		case ' ', '\t':
			(n)++
			continue
		default:
			value = data[n:]
		}
		break
	}

	return // line, key, value, nil
}

// ParseContentLength trims whitespace from s and returns -1 if no value
// is set, or the value if it's >= 0.
func parseContentLength(clen string) (int64, error) {

	if clen == "" {
		return (0), nil
	}

	n, err := strconv.ParseInt(clen, 10, 64)

	if (err != nil) || (n < 0) {
		return -1, errors.New("Content-Length decode error: " + err.Error())
	}
	return (n), nil
}

// UpcomingHeader returns an approximate number of newline(s)
// that will be in this header. If it gets confused, it returns 0.
func upcomingHeader(data []byte) (n int) {
	for len(data) > 0 {
		i := bytes.IndexByte(data, '\n')
		if i < 3 {
			// Not present (-1) or found within the next few bytes,
			// implying we're at the end ("\r\n\r\n" or "\n\n")
			return
		}
		(n)++
		data = data[i+1:]
	}
	return
}

// UrlDecode returns slice of b without escaped hexpair(s) format [%][hex][hex]
func urldecode(s []byte) []byte {
	var w int
	for r := 0; (r) < len(s); (r)++ {
		if s[r] == '%' && len(s[r:]) > (2) {
			if hex0, ok := hexbyte(s[r+1]); ok {
				if hex1, ok := hexbyte(s[r+2]); ok {
					s[w] = (hex0 << 4) | hex1
					r += 2
					w++
					continue
				}
			}
		}
		// Move...
		if w < r {
			s[w] = s[r]
		}
		(w)++
	}
	return s[:w]
}

// fromHexChar converts a hex character into its value and a success flag.
func hexbyte(hex byte) (byte, bool) {

	switch {

	case ('0' <= hex) && (hex <= '9'):
		return (hex - '0'), true
	case ('a' <= hex) && (hex <= 'f'):
		return (hex - 'a' + 10), true
	case ('A' <= hex) && (hex <= 'F'):
		return (hex - 'A' + 10), true
	}

	return (0), (false)
}
