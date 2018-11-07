package logbunny

import "go.uber.org/zap"

// globalLog can used to print the log directly
var globalLog Logger

func init() {
	var err error
	globalLog, err = New()
	if err != nil {
		internalError("error in init the logger", err, nil)
	}
}

func GlobalLogger() Logger {
	return globalLog
}

func SetGlobalLogger(log Logger) {
	if lg, ok := log.(*zapLogger); ok {
		globalLog.(*zapLogger).lg = lg.lg.WithOptions(zap.AddCallerSkip(defaultGlobalCallSkip))
		return
	}
	globalLog = log
}

// Debug can print out the log directly with the globalLog
func Debug(msg string, fields ...*Field) {
	globalLog.Debug(msg, fields...)
}

// Info can print out the log directly with the globalLog
func Info(msg string, fields ...*Field) {
	globalLog.Info(msg, fields...)
}

// Warn can print out the log directly with the globalLog
func Warn(msg string, fields ...*Field) {
	globalLog.Warn(msg, fields...)
}

// Error can print out the log directly with the globalLog
func Error(msg string, fields ...*Field) {
	globalLog.Error(msg, fields...)
}

// Panic can print out the log directly with the globalLog
func Panic(msg string, fields ...*Field) {
	globalLog.Panic(msg, fields...)
}

// Fatal can print out the log directly with the globalLog
func Fatal(msg string, fields ...*Field) {
	globalLog.Fatal(msg, fields...)
}

// Sugared logger
func Sugar() *SugaredLogger {
	return globalLog.Sugar()
}

//Check returns a CheckedEntry if logging a message at the specified level is enabled
func Check(lvl LogLevel, msg string) *CheckedEntry {
	return globalLog.Check(lvl, msg)
}
