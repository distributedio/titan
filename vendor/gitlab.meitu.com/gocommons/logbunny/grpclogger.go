package logbunny

import "go.uber.org/zap"

type GrpcLogger struct {
	l *SugaredLogger
}

func NewGrpcLogger(l Logger) *GrpcLogger {
	if l == nil {
		return nil
	}

	core, ok := l.(*zapLogger)
	if !ok {
		return nil
	}
	core.lg = core.lg.WithOptions(zap.AddCallerSkip(defaultGrpcCallSkip))
	return &GrpcLogger{core.Sugar()}
}

func (l *GrpcLogger) Print(args ...interface{}) {
	l.l.Info(args...)
}
func (l *GrpcLogger) Printf(format string, args ...interface{}) {
	l.l.Infof(format, args...)

}
func (l *GrpcLogger) Println(args ...interface{}) {
	l.l.Info(args...)
}

func (l *GrpcLogger) Fatal(args ...interface{}) {
	l.l.Fatal(args...)
}
func (l *GrpcLogger) Fatalf(format string, args ...interface{}) {
	l.l.Fatalf(format, args...)
}
func (l *GrpcLogger) Fatalln(args ...interface{}) {
	l.l.Fatal(args...)
}
