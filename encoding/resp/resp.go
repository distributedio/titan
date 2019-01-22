package resp

import (
	"errors"
	"io"
	"strconv"
)

var (
	//ErrInvalidProtocol indicates a wrong protocol format
	ErrInvalidProtocol = errors.New("invalid protocol")
)

// ReplyError replies an error
func ReplyError(w io.Writer, msg string) error {
	return NewEncoder(w).Error(msg)
}

// ReplySimpleString replies a simplestring
func ReplySimpleString(w io.Writer, msg string) error {
	return NewEncoder(w).SimpleString(msg)
}

// ReplyBulkString replies a bulkstring
func ReplyBulkString(w io.Writer, msg string) error {
	return NewEncoder(w).BulkString(msg)
}

// ReplyNullBulkString replies a null bulkstring
func ReplyNullBulkString(w io.Writer) error {
	return NewEncoder(w).NullBulkString()
}

// ReplyInteger replies an integer
func ReplyInteger(w io.Writer, val int64) error {
	return NewEncoder(w).Integer(val)
}

// ReplyArray replies an array
func ReplyArray(w io.Writer, size int) (*Encoder, error) {
	r := NewEncoder(w)
	if err := r.Array(size); err != nil {
		return nil, err
	}
	return r, nil
}

// ReadError reads an error
func ReadError(r io.Reader) (string, error) {
	return NewDecoder(r).Error()
}

// ReadSimpleString reads a simplestring
func ReadSimpleString(r io.Reader) (string, error) {
	return NewDecoder(r).SimpleString()
}

// ReadBulkString reads a bulkstring
func ReadBulkString(r io.Reader) (string, error) {
	return NewDecoder(r).BulkString()
}

// ReadInteger reads a integer
func ReadInteger(r io.Reader) (int64, error) {
	return NewDecoder(r).Integer()
}

// ReadArray reads an array
func ReadArray(r io.Reader) (int, error) {
	return NewDecoder(r).Array()
}

// Encoder implements the Encoder interface
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a RESP encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

//Error builds a RESP error
func (r *Encoder) Error(s string) error {
	_, err := r.w.Write([]byte("-" + s + "\r\n"))
	return err
}

//SimpleString builds a RESP simplestring
func (r *Encoder) SimpleString(s string) error {
	_, err := r.w.Write([]byte("+" + s + "\r\n"))
	return err
}

//BulkString builds a RESP bulkstring
func (r *Encoder) BulkString(s string) error {
	length := strconv.Itoa(len(s))
	_, err := r.w.Write([]byte("$" + length + "\r\n" + s + "\r\n"))
	return err
}

// NullBulkString builds a RESP null bulkstring
func (r *Encoder) NullBulkString() error {
	_, err := r.w.Write([]byte("$-1\r\n"))
	return err
}

// Integer builds a RESP integer
func (r *Encoder) Integer(v int64) error {
	s := strconv.FormatInt(v, 10)
	_, err := r.w.Write([]byte(":" + s + "\r\n"))
	return err
}

// Array builds a RESP array
func (r *Encoder) Array(size int) error {
	s := strconv.Itoa(size)
	_, err := r.w.Write([]byte("*" + s + "\r\n"))
	return err
}

// Decoder implements the decoder interface
type Decoder struct {
	r *Reader
}

// NewDecoder creates a RESP decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{&Reader{r}}
}

//Reader implements a reader which supports reading to a delimer
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

//Read bytes into p
func (r *Reader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

//Error parses a RESP error
func (r *Decoder) Error() (string, error) {
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

//SimpleString parses a RESP simplestring
func (r *Decoder) SimpleString() (string, error) {
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

//BulkString parses a RESP bulkstring
func (r *Decoder) BulkString() (string, error) {
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

	if hdr[0] == '-' && len(hdr) == 2 && hdr[1] == '1' {
		// handle $-1 and $-1 null replies.
		return "", nil
	}

	remain, err := strconv.Atoi(string(hdr[1 : l-2]))
	if err != nil || remain < 0 {
		return "", ErrInvalidProtocol
	}

	body := make([]byte, remain+2) //end with \r\n
	_, err = io.ReadFull(r.r, body)
	if err != nil {
		return "", ErrInvalidProtocol
	}
	return string(body[:len(body)-2]), nil
}

//Array parses a RESP array
func (r *Decoder) Array() (int, error) {
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
	if err != nil || remain < 0 {
		return -1, ErrInvalidProtocol
	}
	return remain, nil
}

//Integer parses a RESP integer
func (r *Decoder) Integer() (int64, error) {
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
