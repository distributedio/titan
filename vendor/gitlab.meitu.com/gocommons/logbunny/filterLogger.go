package logbunny

import (
	"io"
)

// filterLogger implement the Logger and contains sublogger which
// ONLY logout the level filter log into given io.Writer
type filterLogger struct {
	loggers map[LogLevel][]Logger
	level   LogLevel
}

// FilterLogger will generate the log with given outs & config
// Split the different log into sublogger with io.Writer
func FilterLogger(c *Config, outs map[LogLevel][]io.Writer) (Logger, error) {
	subloggers := make(map[LogLevel][]Logger)
	for k, v := range outs {
		for _, out := range v {
			cfg := *c
			cfg.Out = out
			cfg.Level = k
			cfg.WithCallerSkip += defaultFilterCallSkip
			log, err := NewWithConfig(&cfg)
			if err != nil {
				return nil, err
			}
			subloggers[k] = append(subloggers[k], log)
		}
	}
	return &filterLogger{loggers: subloggers, level: c.Level}, nil
}

// SetLevel set the filter logger level with given level
// SetLevel will not and should not apply the level to the subloggers
func (log *filterLogger) SetLevel(lv LogLevel) {
	log.level = lv
}

// Level return the filter logger level , this level is a global level filter
// which control the output wether call the sub logger to log the message
func (log *filterLogger) Level() LogLevel {
	return log.level
}

// Line below implement the Logger, only different is here got a enabler to check the level
func (log *filterLogger) Log(lv LogLevel, msg string, fields ...*Field) {
	if log.level.enable(lv) {
		for _, lg := range log.loggers[lv] {
			lg.Log(lv, msg, fields...)
		}
	}
}

func (log *filterLogger) Debug(msg string, fields ...*Field) {
	log.Log(DebugLevel, msg, fields...)
}

func (log *filterLogger) Info(msg string, fields ...*Field) {
	log.Log(InfoLevel, msg, fields...)
}

func (log *filterLogger) Warn(msg string, fields ...*Field) {
	log.Log(WarnLevel, msg, fields...)
}

func (log *filterLogger) Error(msg string, fields ...*Field) {
	log.Log(ErrorLevel, msg, fields...)
}

func (log *filterLogger) Panic(msg string, fields ...*Field) {
	log.Log(PanicLevel, msg, fields...)
}

func (log *filterLogger) Fatal(msg string, fields ...*Field) {
	log.Log(FatalLevel, msg, fields...)
}
func (log *filterLogger) With(fields ...*Field) Logger {
	if len(fields) == 0 {
		return log
	}
	l := *log
	l.loggers = make(map[LogLevel][]Logger)
	for k, v := range log.loggers {
		l.loggers[k] = make([]Logger, len(v))
		copy(l.loggers[k], v)

	}
	for lv, logs := range l.loggers {
		for k, sg := range logs {
			l.loggers[lv][k] = sg.With(fields...)
		}
	}
	return &l
}

// Not supported now
func (log *filterLogger) Sugar() *SugaredLogger {
	return nil
}

// FIXME when refactor filterlogger
func (log *filterLogger) Check(lvl LogLevel, msg string) *CheckedEntry {
	return nil
}
