package logbunny

import (
	"go.uber.org/zap"
)

// Tee is used to combine some logger into one. Tee give the msg a copy to the sublogger to logout
func Tee(cfg *Config, logs ...Logger) Logger {
	switch len(logs) {
	case 0:
		return nil
	case 1:
		return logs[0]
	default:
		log := &multiLogger{loggers: logs}
		for _, l := range logs {
			core, ok := l.(*zapLogger)
			if ok {
				core.lg = core.lg.WithOptions(zap.AddCallerSkip(defaultTeeCallSkip))
			}
		}
		return log
	}
}

// multiLogger is used for Tee to combine Loggers into one
type multiLogger struct {
	loggers    []Logger
	level      LogLevel
	withCaller bool
}

func (log *multiLogger) SetLevel(lv LogLevel) {
	log.level = lv
	for _, lg := range log.loggers {
		lg.SetLevel(lv)
	}
}

func (log *multiLogger) Level() LogLevel {
	return log.level
}

// line below implement the logger
func (log *multiLogger) Log(lv LogLevel, msg string, fields ...*Field) {
	if log.level.enable(lv) {
		for _, lg := range log.loggers {
			lg.Log(lv, msg, fields...)
		}
	}
}
func (log *multiLogger) Debug(msg string, fields ...*Field) {
	log.Log(DebugLevel, msg, fields...)
}
func (log *multiLogger) Info(msg string, fields ...*Field) {
	log.Log(InfoLevel, msg, fields...)
}
func (log *multiLogger) Warn(msg string, fields ...*Field) {
	log.Log(WarnLevel, msg, fields...)
}
func (log *multiLogger) Error(msg string, fields ...*Field) {
	log.Log(ErrorLevel, msg, fields...)
}
func (log *multiLogger) Panic(msg string, fields ...*Field) {
	log.Log(PanicLevel, msg, fields...)
}
func (log *multiLogger) Fatal(msg string, fields ...*Field) {
	log.Log(FatalLevel, msg, fields...)
}
func (log *multiLogger) With(fields ...*Field) Logger {
	if len(fields) == 0 {
		return log
	}
	l := *log
	l.loggers = make([]Logger, len(log.loggers))
	copy(l.loggers, log.loggers)
	for k, sg := range l.loggers {
		l.loggers[k] = sg.With(fields...)
	}
	return &l
}

// Not supported now
func (log *multiLogger) Sugar() *SugaredLogger {
	return nil
}

// FIXME when refactor tee
func (log *multiLogger) Check(lvl LogLevel, msg string) *CheckedEntry {
	return nil
}
