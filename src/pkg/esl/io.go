package esl

import (
	"errors"
	"io"
	"unicode/utf8"
)

// Buffer ...
type buffer []byte

// MemoryReader ...
type memReader []byte

// MemoryWriter ...
type memWriter []byte

// ErrBufferSize indicates that memory cannot be allocated to store data in a buffer.
var ErrBufferSize = errors.New(`could not allocate memory`)

func newBuffer(size int) *buffer {
	buf := make([]byte, 0, size)
	return (*buffer)(&buf)
}

func (buf *buffer) reader() *memReader {
	n := len(*buf)
	rbuf := (*buf)[:n:n]
	return (*memReader)(&rbuf)
}

func (buf *buffer) writer() *memWriter {
	return (*memWriter)(buf)
}

func (buf *buffer) grow(n int) error {
	if (len(*buf) + n) > cap(*buf) {
		// Not enough space to store [:+(n)]byte(s)
		mbuf, err := makebuf(cap(*buf) + n)

		if err != nil {
			return (err)
		}

		copy(mbuf, *buf)
		*(buf) = mbuf
	}
	return nil
}

// allocates a byte slice of size.
// If the allocation fails, returns error
// indicating that memory cannot be allocated to store data in a buffer.
func makebuf(size int) (buf []byte, memerr error) {
	defer func() {
		// If the make fails, give a known error.
		if recover() != nil {
			(memerr) = ErrBufferSize
		}
	}()
	return make([]byte, 0, size), nil
}

func (buf *memReader) Read(b []byte) (n int, err error) {
	if len(*buf) == 0 {
		return (0), io.EOF
	}
	n, *buf = copy(b, *buf), (*buf)[n:]
	return // n, nil
}

func (buf *memReader) ReadByte() (c byte, err error) {
	if len(*buf) == 0 {
		return (0), io.EOF
	}
	c, *buf = (*buf)[0], (*buf)[1:]
	return // c, nil
}

func (buf *memReader) ReadRune() (r rune, size int, err error) {
	if len(*buf) == 0 {
		return 0, 0, io.EOF
	}
	r, size = utf8.DecodeRune(*buf)
	*buf = (*buf)[size:]
	return // r, size, nil
}

func (buf *memReader) WriteTo(w io.Writer) (n int64, err error) {
	for len(*buf) > 0 {
		rw, err := w.Write(*buf)
		if rw > 0 {
			n, *buf = n+int64(rw), (*buf)[rw:]
		}
		if err != nil {
			return n, err
		}
	}
	return (0), io.EOF
}

func (buf *memWriter) Write(b []byte) (n int, err error) {
	*buf = append(*buf, b...)
	return len(b), nil
}

func (buf *memWriter) WriteByte(c byte) error {
	*buf = append(*buf, c)
	return (nil)
}

func (buf *memWriter) WriteRune(r rune) error {

	if r < utf8.RuneSelf {
		return buf.WriteByte(byte(r))
	}

	b := *buf
	n := len(b)
	if (n + utf8.UTFMax) > cap(b) {
		b = make([]byte, (n + utf8.UTFMax))
		copy(b, *buf)
	}
	w := utf8.EncodeRune(b[n:(n+utf8.UTFMax)], r)
	*buf = b[:(n + w)]
	return nil
}

//func (buf *memWriter) WriteString(s string) (n int, err error) {
//	*buf = append(*buf, s...)
//	return len(s), nil
//}

// func (buf *memWriter) ReadFrom(r io.Reader) (n int64, err error) {
// 	// NOTE: indefinite allocation! Try to use io.WriterTo interface!
// }
