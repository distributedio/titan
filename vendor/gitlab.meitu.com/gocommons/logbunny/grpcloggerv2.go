package logbunny

import "go.uber.org/zap"

type GrpcLoggerV2 struct {
	l *SugaredLogger
}

func NewGrpcLoggerV2(l Logger) *GrpcLoggerV2 {
	if l == nil {
		return nil
	}
	core, ok := l.(*zapLogger)
	if !ok {
		return nil
	}
	core.lg = core.lg.WithOptions(zap.AddCallerSkip(defaultGrpcCallSkip))
	return &GrpcLoggerV2{core.Sugar()}
}

func (l *GrpcLoggerV2) Info(args ...interface{}) {
	l.l.Info(args...)
}
func (l *GrpcLoggerV2) Infof(format string, args ...interface{}) {
	l.l.Infof(format, args...)
}
func (l *GrpcLoggerV2) Infoln(args ...interface{}) {
	l.l.Info(args...)
}

func (l *GrpcLoggerV2) Warning(args ...interface{}) {
	l.l.Warn(args...)
}
func (l *GrpcLoggerV2) Warningf(format string, args ...interface{}) {
	l.l.Warnf(format, args...)
}
func (l *GrpcLoggerV2) Warningln(args ...interface{}) {
	l.l.Warn(args...)
}

func (l *GrpcLoggerV2) Error(args ...interface{}) {
	l.l.Error(args...)
}
func (l *GrpcLoggerV2) Errorf(format string, args ...interface{}) {
	l.l.Errorf(format, args...)
}
func (l *GrpcLoggerV2) Errorln(args ...interface{}) {
	l.l.Error(args...)
}

func (l *GrpcLoggerV2) Fatal(args ...interface{}) {
	l.l.Fatal(args...)
}
func (l *GrpcLoggerV2) Fatalf(format string, args ...interface{}) {
	l.l.Fatalf(format, args...)
}
func (l *GrpcLoggerV2) Fatalln(args ...interface{}) {
	l.l.Fatal(args...)
}

func (l *GrpcLoggerV2) V(level int) bool {
	// level of zap:      debug(-1), info(0), warn(1), error(2), dpanic(3), panic(4), fatal(5)
	// level of grpclog:  info(0), warn(1), error(2), fatal(3)
	// level of logbunny: debug(0), info(1), warn(2), error(3), panic(4), fatal(5)
	// map from grpclog level to logbunny level here
	levelmap := []int{1, 2, 3, 5}

	level = levelmap[level]
	return level >= int(l.l.base.Level())
}
