package logbunny

// LogLevel is used to unique the different log level
type LogLevel int32

// Zap & logrus level is not the same so we defined our level
const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

// enable is used to check the level if is allowed to log out
func (l LogLevel) enable(lv LogLevel) bool {
	return lv >= l
}
