package logbunny

import (
	"go.uber.org/zap"
)

type SugaredLogger struct {
	*zap.SugaredLogger
	base Logger
}

func (s *SugaredLogger) Desugar() Logger {
	return s.base
}
