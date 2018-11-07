package command

import (
	"errors"
	"fmt"
)

// RedisError defines redis protocol error
type RedisError error

const (
	// UnKnownCommandStr is the command not find
	UnKnownCommandStr = "unknown command '%s'"
	// WrongArgs is for wrong number of arguments error
	WrongArgs = "ERR wrong number of arguments for '%s' command"
)

var (
	// OK is the simple string "OK" return to client
	OK = "OK"

	// Empty is the simple string "EMPTY" return to client
	Empty = "Empty"

	// Queued is the simple string "QUEUED" return to client
	Queued = "QUEUED"

	// ErrProtocol invalid request
	ErrProtocol = errors.New("ERR invalid request")

	// ErrUnAuth Authentication required
	ErrUnAuth = errors.New("ERR Authentication required")

	// ErrAuthInvalid invalid password
	ErrAuthInvalid = errors.New("ERR invalid password")

	// ErrAuthUnSet Client sent AUTH, but no password is set
	ErrAuthUnSet = errors.New("ERR Client sent AUTH, but no password is set")

	// ErrUnImplement this command is un implement
	ErrUnImplement = errors.New("ERR this command is un implement")

	// ErrInvalidDB invalid DB index
	ErrInvalidDB = errors.New("ERR invalid DB index")

	// ErrTikv TIKV ERROR
	ErrTikv = errors.New("TIKV ERROR")

	// ErrInteger value is not an integer or out of range
	ErrInteger = errors.New("ERR value is not an integer or out of range")

	// ErrBitInteger bit is not an integer or out of range
	ErrBitInteger = errors.New("ERR bit is not an integer or out of range")

	// ErrBitOffset bit offset is not an integer or out of range
	ErrBitOffset = errors.New("ERR bit offset is not an integer or out of range")

	// ErrOffset offset is out of range
	ErrOffset = errors.New("ERR offset is out of range")

	// ErrIndex offset is out of range
	ErrIndex = errors.New("ERR index out of range")

	// ErrSyntax syntax error
	ErrSyntax = errors.New("ERR syntax error")

	// ErrValue value is not an integer or out of range
	ErrValue = errors.New("ERR value is not an integer or out of range")

	// ErrType Operation against a key holding the wrong kind of value
	ErrType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

	// ErrMSet wrong number of arguments for MSET
	ErrMSet = errors.New("ERR wrong number of arguments for MSET")

	// ErrNoSuchKey reteurn on lset for key which no exist
	ErrNoSuchKey = errors.New("ERR no such key")

	// ErrReturnType return data type error
	ErrReturnType = errors.New("ERR return data type error")

	//ErrMaximum allows the maximum size of a string
	ErrMaximum = errors.New("ERR string exceeds maximum allowed size")

	// ErrMultiNested indicates a nested multi command which is not allowed
	ErrMultiNested = errors.New("ERR MULTI calls can not be nested")
)

//ErrUnKnownCommand return RedisError of the cmd
func ErrUnKnownCommand(cmd string) error {
	return fmt.Errorf(UnKnownCommandStr, cmd)
}

// ErrWrongArgs return RedisError of the cmd
func ErrWrongArgs(cmd string) error {
	return fmt.Errorf(WrongArgs, cmd)
}
