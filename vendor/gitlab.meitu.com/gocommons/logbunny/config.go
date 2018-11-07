package logbunny

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// LogType defined the type about the logger
type LogType int

const (
	// ZapLogger defined the type about the zap
	ZapLogger LogType = iota
)

// EncoderType Defined the encoder type for logger
type EncoderType int

const (
	// JSONEncoder Defined the JSON encoder type for logger
	JSONEncoder EncoderType = iota
	// ConsoleEncoder Defined the Console\Text encoder type for logger
	ConsoleEncoder
)

// LogbunnyConfig used to construct the logbunny from config file
type LogbunnyConfig struct {
	Type       string `json:"type"`         //Set the logger type
	Level      string `json:"level"`        //Default logger level
	Encoder    string `json:"encoder"`      //JSON or text encoder
	WithCaller bool   `json:"with_caller"`  //Print the filename & line number within the log if set true, default is false
	WithNoLock bool   `json:"with_no_lock"` //If use the zap with no lock if set true, default is false
}

func NewConfigFromFile(path string, out io.Writer) (*Config, error) {
	v, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cf := &LogbunnyConfig{}
	if err := json.Unmarshal(v, cf); err != nil {
		return nil, err
	}
	var ops []Option

	if cf.WithNoLock {
		ops = append(ops, WithNoLock())
	}

	switch cf.Encoder {
	case "console":
		ops = append(ops, WithConsoleEncoder())
	case "json", "JSON":
		ops = append(ops, WithJSONEncoder())
	default:
		ops = append(ops, WithJSONEncoder())
	}

	switch strings.ToLower(cf.Level) {
	case "debug":
		ops = append(ops, WithDebugLevel())
	case "info":
		ops = append(ops, WithInfoLevel())
	case "warn":
		ops = append(ops, WithWarnLevel())
	case "error":
		ops = append(ops, WithErrorLevel())
	case "fatal":
		ops = append(ops, WithFatalLevel())
	case "panic":
		ops = append(ops, WithPanicLevel())
	default:
		ops = append(ops, WithInfoLevel())
	}

	switch strings.ToLower(cf.Type) {
	case "zap":
		ops = append(ops, WithZapLogger())
	default:
		ops = append(ops, WithZapLogger())
	}

	if out != nil {
		ops = append(ops, WithOutput(out))
	}

	return NewConfig(ops...), nil
}

// Config include the core configuration to build a logbunny logger
type Config struct {
	Field          []*Field
	Type           LogType     //defined the logger type
	Level          LogLevel    //default logger level
	Encoder        EncoderType //json or text encoder
	WithCaller     bool        //print the filename & line number within the log
	Out            io.Writer   //defined the out put io writer TBD
	WithNoLock     bool        //wether add the lock for the zap writer, default is false
	TimePattern    string      //Define the time pattern for use
	Name           string      //Named adds a new path segment to the logger's name
	WithMetrics    bool        //enable metrics hook or not
	WithCallerSkip int         //callers skipped, default is 0
}

// NewConfig will return a config struct
func NewConfig(opts ...Option) *Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Option used to set the option to Config
type Option func(c *Config)

// WithOutput will get a Option set for the config Out
func WithOutput(o io.Writer) Option {
	return func(c *Config) {
		if o == nil {
			internalError("error in withoutput option", ErrArgumentIllegal, nil)
			o = os.Stderr
		}
		c.Out = o
	}
}

// WithNoLock will get a Option set for the config withlock
func WithNoLock() Option { return func(c *Config) { c.WithNoLock = true } }

// WithCaller will get a Option set for the config withcaller
func WithCaller() Option { return func(c *Config) { c.WithCaller = true } }

// WithZapLogger will get a Option set for the config logger type
func WithZapLogger() Option { return func(c *Config) { c.Type = ZapLogger } }

// WithJSONEncoder will get a Option set for the config encoder type
func WithJSONEncoder() Option { return func(c *Config) { c.Encoder = JSONEncoder } }

// WithConsoleEncoder will get a Option set for the config encoder type
func WithConsoleEncoder() Option { return func(c *Config) { c.Encoder = ConsoleEncoder } }

// WithDebugLevel will get a option set for the config level
func WithDebugLevel() Option { return func(c *Config) { c.Level = DebugLevel } }

// WithInfoLevel will get a option set for the config level
func WithInfoLevel() Option { return func(c *Config) { c.Level = InfoLevel } }

// WithWarnLevel will get a option set for the config level
func WithWarnLevel() Option { return func(c *Config) { c.Level = WarnLevel } }

// WithErrorLevel will get a option set for the config level
func WithErrorLevel() Option { return func(c *Config) { c.Level = ErrorLevel } }

// WithPanicLevel will get a option set for the config level
func WithPanicLevel() Option { return func(c *Config) { c.Level = PanicLevel } }

// WithFatalLevel will get a option set for the config level
func WithFatalLevel() Option { return func(c *Config) { c.Level = FatalLevel } }

// WithTimePattern will set the time pattern to the logger
func WithTimePattern(pattern string) Option { return func(c *Config) { c.TimePattern = pattern } }

// Named adds a new path segment to the logger's name.  By default, Loggers are unnamed.
func WithName(name string) Option { return func(c *Config) { c.Name = name } }

// WithFields will add fields as internal part of logger
func WithFields(fields ...*Field) Option { return func(c *Config) { c.Field = fields } }

// WithoutMetrics disable the log statistics
func WithMetrics() Option { return func(c *Config) { c.WithMetrics = true } }

// WithCallerSkip increases the number of callers skipped by caller annotation(as enabled by the WithCaller option).
func WithCallerSkip(skip int) Option { return func(c *Config) { c.WithCallerSkip = skip } }
