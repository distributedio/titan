package logbunny

import (
	"errors"
	"log"

	"go.uber.org/zap"
)

var ErrStdInvalid = errors.New("log cannot reflect zapLogger")

// NewStdLog returns a *log.Logger which writes to the supplied zap Logger at
// InfoLevel. To redirect the standard library's package-global logging
// functions, use RedirectStdLog instead.
func NewStdLog(l Logger) *log.Logger {
	if log, ok := l.(*zapLogger); ok {
		lg := log.lg.WithOptions(zap.AddCallerSkip(defaultStdCallSkip))
		return zap.NewStdLog(lg)
	}
	return nil
}

// NewStdLogAt returns *log.Logger which writes to supplied zap logger at
// required level.
func NewStdLogAt(l Logger, level LogLevel) (*log.Logger, error) {
	if log, ok := l.(*zapLogger); ok {
		lv := toZapLevel(level)
		lg := log.lg.WithOptions(zap.AddCallerSkip(defaultStdCallSkip))
		return zap.NewStdLogAt(lg, lv)
	}
	return nil, ErrStdInvalid
}

// RedirectStdLog redirects output from the standard library's package-global
// logger to the supplied logger at InfoLevel. Since zap already handles caller
// annotations, timestamps, etc., it automatically disables the standard
// library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stderr.
func RedirectStdLog(l Logger) func() {
	if log, ok := l.(*zapLogger); ok {
		lg := log.lg.WithOptions(zap.AddCallerSkip(defaultStdCallSkip))
		return zap.RedirectStdLog(lg)
	}
	return nil
}

// RedirectStdLogAt redirects output from the standard library's package-global
// logger to the supplied logger at the specified level. Since zap already
// handles caller annotations, timestamps, etc., it automatically disables the
// standard library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stderr.
func RedirectStdLogAt(l Logger, level LogLevel) (func(), error) {
	if log, ok := l.(*zapLogger); ok {
		lv := toZapLevel(level)
		lg := log.lg.WithOptions(zap.AddCallerSkip(defaultStdCallSkip))
		return zap.RedirectStdLogAt(lg, lv)
	}
	return nil, ErrStdInvalid
}
