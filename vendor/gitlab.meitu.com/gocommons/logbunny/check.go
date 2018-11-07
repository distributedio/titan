package logbunny

import (
	"go.uber.org/zap/zapcore"
)

type CheckedEntry struct {
	ce *zapcore.CheckedEntry
}

func (ce *CheckedEntry) Write(fields ...*Field) {
	//add another allocation ?
	zapFields := make([]zapcore.Field, len(fields))
	for i := range fields {
		zapFields[i] = zapcore.Field(*fields[i])
	}
	ce.ce.Write(zapFields...)
}
