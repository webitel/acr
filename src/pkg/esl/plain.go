package esl

import (
	"fmt"
	"io"
	"strconv"
)

// Message content.
type Message struct {
	Header Header
	Body   []byte
}

// Headers that Request.Write handles itself and should be skipped.
var msgWriteExcludeHeader = map[string]bool{
	`Content-Length`: true,
}

func (msg Message) WriteHeader(w io.Writer) (n int, err error) {

	n, err = msg.Header.WriteSubset(w, msgWriteExcludeHeader)

	if err != nil {
		return n, err
	}

	if contentLength := len(msg.Body); contentLength > 0 {
		wn, err := WriteHeader(w, "Content-Length", strconv.Itoa(contentLength))
		if n += int(wn); err != nil {
			return n, err
		}
	}

	if (n) > 0 { // Finally: \n\n
		wn, err := w.Write(lineFeed)
		if n += int(wn); err != nil {
			return n, err
		}
	}

	return // n, err
}

func (msg Message) PlainText(w io.Writer) (n int, err error) {

	n, err = msg.WriteHeader(w)

	if err != nil {
		return
	}

	cLength := len(msg.Body)

	if cLength > 0 {

		wn, err := w.Write(msg.Body)
		if n += int(wn); err != nil {
			return n, err
		}

		for i := (0); i < 2; (i)++ { // Finally: \n\n

			if (cLength > i) && msg.Body[cLength-i-1] != ('\n') {
				wn, err := w.Write(lineFeed)
				if n += int(wn); err != nil {
					return n, err
				}
			}
		}
	}

	return
}

func (msg Message) Write(body []byte) (n int, err error) {
	return
}

// PlainEvent perform event basic identification ...
func PlainEvent(plain Message) (int, string, error) {
	event, err := ParseEvent(plain.Header.Get(`Event-Name`))

	if err != nil {
		return (0), (""), (err)
	}

	if event == EventCustom {
		subclass := plain.Header.Get(`Event-Subclass`)
		if subclass == "" {
			return (0), (""), fmt.Errorf("!Custom(?)")
		}
		return (EventCustom), (subclass), (nil)
	}

	return (event), (""), (nil)
}
