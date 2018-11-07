package logbunny

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger implement the logbunny logger
type zapLogger struct {
	lg                  *zap.Logger
	dynamicLevelHandler zap.AtomicLevel
}

// this two func used for trans the level between zap & logbunny
func toZapLevel(lv LogLevel) zapcore.Level {
	switch lv {
	case DebugLevel:
		return zap.DebugLevel
	case InfoLevel:
		return zap.InfoLevel
	case WarnLevel:
		return zap.WarnLevel
	case ErrorLevel:
		return zap.ErrorLevel
	case PanicLevel:
		return zap.PanicLevel
	case FatalLevel:
		return zap.FatalLevel
	default:
		return 0x7f
	}
}

func fromZapLevel(lv zapcore.Level) LogLevel {
	switch lv {
	case zap.DebugLevel:
		return DebugLevel
	case zap.InfoLevel:
		return InfoLevel
	case zap.WarnLevel:
		return WarnLevel
	case zap.ErrorLevel:
		return ErrorLevel
	case zap.PanicLevel:
		return PanicLevel
	case zap.FatalLevel:
		return FatalLevel
	default:
		return 0x7f
	}
}

// Level will also return a level handler for you, you can totally ignore it
func (log *zapLogger) Level() LogLevel {
	return fromZapLevel(log.dynamicLevelHandler.Level())
}

// SetLevel will also return a level handler for you, you can totally ignore it
func (log *zapLogger) SetLevel(lv LogLevel) {
	log.dynamicLevelHandler.SetLevel(toZapLevel(lv))
}

func (log *zapLogger) Log(lv LogLevel, msg string, fields ...*Field) {
	var msgfields []zapcore.Field
	msgfields = zapFields(fields...)

	switch lv {
	case DebugLevel:
		log.lg.Debug(msg, msgfields...)
	case InfoLevel:
		log.lg.Info(msg, msgfields...)
	case WarnLevel:
		log.lg.Warn(msg, msgfields...)
	case ErrorLevel:
		log.lg.Error(msg, msgfields...)
	case PanicLevel:
		log.lg.Panic(msg, msgfields...)
	case FatalLevel:
		log.lg.Fatal(msg, msgfields...)
	default:
		internalError("error in Log switch level", ErrTypeMissmatch, lv)
	}
}

func (log *zapLogger) Debug(msg string, fields ...*Field) {
	log.Log(DebugLevel, msg, fields...)
}

func (log *zapLogger) Info(msg string, fields ...*Field) {
	log.Log(InfoLevel, msg, fields...)
}

func (log *zapLogger) Warn(msg string, fields ...*Field) {
	log.Log(WarnLevel, msg, fields...)
}

func (log *zapLogger) Error(msg string, fields ...*Field) {
	log.Log(ErrorLevel, msg, fields...)
}

func (log *zapLogger) Panic(msg string, fields ...*Field) {
	log.Log(PanicLevel, msg, fields...)
}

func (log *zapLogger) Fatal(msg string, fields ...*Field) {
	log.Log(FatalLevel, msg, fields...)
}

// With creates a child logger and adds structured context to it. Fields added to the child don't affect the parent, and vice versa.
func (log *zapLogger) With(fields ...*Field) Logger {
	if len(fields) == 0 {
		return log
	}
	l := *log
	l.lg = l.lg.With(zapFields(fields...)...)
	return &l
}

func (log *zapLogger) Sugar() *SugaredLogger {
	lg := log.lg.WithOptions(zap.AddCallerSkip(defaultSugCallSkip))
	return &SugaredLogger{SugaredLogger: lg.Sugar(), base: log}
}

// zapFields make zap field slice with given field
func zapFields(fields ...*Field) []zapcore.Field {
	msgfields := make([]zapcore.Field, 0, len(fields))
	for _, f := range fields {
		msgfields = append(msgfields, (zapcore.Field)(*f))
	}
	return msgfields
}

func (log *zapLogger) Check(lvl LogLevel, msg string) *CheckedEntry {
	zaplvl := toZapLevel(lvl)
	lg := log.lg.WithOptions(zap.AddCallerSkip(defaultCheckCallSkip))
	return &CheckedEntry{lg.Check(zaplvl, msg)}
}
