package resp

import (
	"errors"
	"io"
	"strconv"
)

var (
	//ErrInvalidProtocol err invalid protocol
	ErrInvalidProtocol = errors.New("invalid protocol")
)

// ReplyError reply client error
func ReplyError(w io.Writer, msg string) error {
	return NewEncoderRESP(w).Error(msg)
}

// ReplySimpleString reply client string
func ReplySimpleString(w io.Writer, msg string) error {
	return NewEncoderRESP(w).SimpleString(msg)
}

// ReplyBulkString reply client slice string
func ReplyBulkString(w io.Writer, msg string) error {
	return NewEncoderRESP(w).BulkString(msg)
}

// ReplyNullBulkString reply client null string
func ReplyNullBulkString(w io.Writer) error {
	return NewEncoderRESP(w).NullBulkString()
}

// ReplyInteger reply client integer
func ReplyInteger(w io.Writer, val int64) error {
	return NewEncoderRESP(w).Integer(val)
}

// ReplyArray reply client array ,there is integer, string in rrray
func ReplyArray(w io.Writer, size int) (Encoder, error) {
	r := NewEncoderRESP(w)
	if err := r.Array(size); err != nil {
		return nil, err
	}
	return r, nil
}

// ReadError read the msg that client send err type
func ReadError(r io.Reader) (string, error) {
	return NewDecoderRESP(r).Error()
}

// ReadSimpleString read the msg that client send string
func ReadSimpleString(r io.Reader) (string, error) {
	return NewDecoderRESP(r).SimpleString()
}

// ReadBulkString read the msg that client send slice string
func ReadBulkString(r io.Reader) (string, error) {
	return NewDecoderRESP(r).BulkString()
}

// ReadInteger read the msg that client send integer
func ReadInteger(r io.Reader) (int64, error) {
	return NewDecoderRESP(r).Integer()
}

// ReadArray read the msg that client send array
func ReadArray(r io.Reader) (int, error) {
	return NewDecoderRESP(r).Array()
}

// EncoderRESP RESP is a RESP encoder/decoder
type EncoderRESP struct {
	w io.Writer
}

// NewEncoderRESP new resp encode object
func NewEncoderRESP(w io.Writer) *EncoderRESP {
	return &EncoderRESP{w}
}

//Error the type err in RESP
func (r *EncoderRESP) Error(s string) error {
	_, err := r.w.Write([]byte("-" + s + "\r\n"))
	return err
}

//SimpleString the type simplestring in RESP
func (r *EncoderRESP) SimpleString(s string) error {
	_, err := r.w.Write([]byte("+" + s + "\r\n"))
	return err
}

//BulkString the type bulkstring in RESP
func (r *EncoderRESP) BulkString(s string) error {
	length := strconv.Itoa(len(s))
	_, err := r.w.Write([]byte("$" + length + "\r\n" + s + "\r\n"))
	return err
}

// NullBulkString the type nullstring in RESP
func (r *EncoderRESP) NullBulkString() error {
	_, err := r.w.Write([]byte("$-1\r\n"))
	return err
}

//Integer the type integer in RESP
func (r *EncoderRESP) Integer(v int64) error {
	s := strconv.FormatInt(v, 10)
	_, err := r.w.Write([]byte(":" + s + "\r\n"))
	return err
}

//Array the type array in RESP
func (r *EncoderRESP) Array(size int) error {
	s := strconv.Itoa(size)
	_, err := r.w.Write([]byte("*" + s + "\r\n"))
	return err
}

//DecoderRESP decode in RESP
type DecoderRESP struct {
	r *Reader
}

// NewDecoderRESP new decoder object
func NewDecoderRESP(r io.Reader) *DecoderRESP {
	return &DecoderRESP{&Reader{r}}
}

//Reader read buffer
type Reader struct {
	r io.Reader
}

// ReadBytes read bytes until delim, it read byte by byte and
// do not buffer anything, you should use a buffer reader as the
// backend to achieve performance
func (r *Reader) ReadBytes(delim byte) ([]byte, error) {
	b := make([]byte, 1)
	buf := make([]byte, 0, 64)
	// it is necessary to read byte by byte thought it seems silly,
	// the reader here just take bytes it needed and should not exceed,
	// or it will make the outside reader's offset unexpected
	for {
		n, err := r.r.Read(b)
		if n == 1 {
			buf = append(buf, b[0])
		}
		if err != nil {
			if err == io.EOF {
				return buf, err
			}
			return nil, err
		}

		if n == 0 {
			continue
		}

		if b[0] == delim {
			return buf, nil
		}
	}

}

//Read read from io to p return read len
func (r *Reader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

//Error the type err in RESP
func (r *DecoderRESP) Error() (string, error) {
	buf, err := r.r.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	l := len(buf)
	if l < len("-\r\n") {
		return "", ErrInvalidProtocol
	}
	if buf[l-2] != '\r' {
		return "", ErrInvalidProtocol
	}
	if buf[0] != '-' {
		return "", ErrInvalidProtocol
	}
	return string(buf[1 : l-2]), nil
}

//SimpleString the type simplestring in RESP
func (r *DecoderRESP) SimpleString() (string, error) {
	buf, err := r.r.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	l := len(buf)
	if l < len("+\r\n") {
		return "", ErrInvalidProtocol
	}
	if buf[l-2] != '\r' {
		return "", ErrInvalidProtocol
	}
	if buf[0] != '+' {
		return "", ErrInvalidProtocol
	}
	return string(buf[1 : l-2]), nil
}

//BulkString the type bulkstring in RESP
func (r *DecoderRESP) BulkString() (string, error) {
	hdr, err := r.r.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	l := len(hdr)
	if l < len("$\r\n") {
		return "", ErrInvalidProtocol
	}
	if hdr[l-2] != '\r' {
		return "", ErrInvalidProtocol
	}
	if hdr[0] != '$' {
		return "", ErrInvalidProtocol
	}

	remain, err := strconv.Atoi(string(hdr[1 : l-2]))
	if err != nil {
		return "", ErrInvalidProtocol
	}

	body := make([]byte, remain+2) //end with \r\n
	_, err = io.ReadFull(r.r, body)
	if err != nil {
		return "", ErrInvalidProtocol
	}
	return string(body[:len(body)-2]), nil
}

//Array the type array in RESP
func (r *DecoderRESP) Array() (int, error) {
	hdr, err := r.r.ReadBytes('\n')
	if err != nil {
		return -1, err
	}
	l := len(hdr)
	if l < len("*\r\n") {
		return -1, ErrInvalidProtocol
	}
	if hdr[l-2] != '\r' {
		return -1, ErrInvalidProtocol
	}
	if hdr[0] != '*' {
		return -1, ErrInvalidProtocol
	}
	remain, err := strconv.Atoi(string(hdr[1 : l-2]))
	if err != nil {
		return -1, ErrInvalidProtocol
	}
	return remain, nil
}

//Integer the type integer in RESP
func (r *DecoderRESP) Integer() (int64, error) {
	val, err := r.r.ReadBytes('\n')
	if err != nil {
		return -1, err
	}
	l := len(val)
	if l < len(":\r\n") {
		return -1, ErrInvalidProtocol
	}
	if val[l-2] != '\r' {
		return -1, ErrInvalidProtocol
	}
	if val[0] != ':' {
		return -1, ErrInvalidProtocol
	}

	v, err := strconv.ParseInt(string(val[1:l-2]), 10, 64)
	if err != nil {
		return -1, ErrInvalidProtocol
	}
	return v, nil

}
