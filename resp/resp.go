package resp

import (
	"errors"
	"io"
	"strconv"
)

var (
	ErrInvalidProtocol = errors.New("invalid protocol")
)

func ReplyError(w io.Writer, msg string) error {
	return NewRESPEncoder(w).Error(msg)
}

func ReplySimpleString(w io.Writer, msg string) error {
	return NewRESPEncoder(w).SimpleString(msg)
}

func ReplyBulkString(w io.Writer, msg string) error {
	return NewRESPEncoder(w).BulkString(msg)
}

func ReplyNullBulkString(w io.Writer) error {
	return NewRESPEncoder(w).NullBulkString()
}

func ReplyInteger(w io.Writer, val int64) error {
	return NewRESPEncoder(w).Integer(val)
}

func ReplyArray(w io.Writer, size int) (Encoder, error) {
	r := NewRESPEncoder(w)
	if err := r.Array(size); err != nil {
		return nil, err
	}
	return r, nil
}

func ReadError(r io.Reader) (string, error) {
	return NewRESPDecoder(r).Error()
}

func ReadSimpleString(r io.Reader) (string, error) {
	return NewRESPDecoder(r).SimpleString()
}

func ReadBulkString(r io.Reader) (string, error) {
	return NewRESPDecoder(r).BulkString()
}

func ReadInteger(r io.Reader) (int64, error) {
	return NewRESPDecoder(r).Integer()
}

func ReadArray(r io.Reader) (int, error) {
	return NewRESPDecoder(r).Array()
}

// RESP is a RESP encoder/decoder
type RESPEncoder struct {
	w io.Writer
}

func NewRESPEncoder(w io.Writer) *RESPEncoder {
	return &RESPEncoder{w}
}

func (r *RESPEncoder) Error(s string) error {
	_, err := r.w.Write([]byte("-" + s + "\r\n"))
	return err
}

func (r *RESPEncoder) SimpleString(s string) error {
	_, err := r.w.Write([]byte("+" + s + "\r\n"))
	return err
}

func (r *RESPEncoder) BulkString(s string) error {
	length := strconv.Itoa(len(s))
	_, err := r.w.Write([]byte("$" + length + "\r\n" + s + "\r\n"))
	return err
}
func (r *RESPEncoder) NullBulkString() error {
	_, err := r.w.Write([]byte("$-1\r\n"))
	return err
}

func (r *RESPEncoder) Integer(v int64) error {
	s := strconv.FormatInt(v, 10)
	_, err := r.w.Write([]byte(":" + s + "\r\n"))
	return err
}

func (r *RESPEncoder) Array(size int) error {
	s := strconv.Itoa(size)
	_, err := r.w.Write([]byte("*" + s + "\r\n"))
	return err
}

type RESPDecoder struct {
	r *Reader
}

func NewRESPDecoder(r io.Reader) *RESPDecoder {
	return &RESPDecoder{&Reader{r}}
}

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
func (r *Reader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}
func (r *RESPDecoder) Error() (string, error) {
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

func (r *RESPDecoder) SimpleString() (string, error) {
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
func (r *RESPDecoder) BulkString() (string, error) {
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

func (r *RESPDecoder) Array() (int, error) {
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

func (r *RESPDecoder) Integer() (int64, error) {
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
