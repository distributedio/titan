package rollingwriter

import (
	"errors"
	"io"
	"os"
	"strings"
)

// RollingPolicies giveout 3 policy for rolling.
const (
	WithoutRolling = iota
	TimeRolling
	VolumeRolling
)

var (
	// BufferSize defined the buffer size, by default 1M buffer will be allocate
	BufferSize = 0x100000
	// QueueSize defined the queue size for asynchronize write
	QueueSize = 1024
	// Precision defined the precision about the reopen operation condition
	// check duration within second
	Precision = 1
	// DefaultFileMode set the default open mode
	DefaultFileMode = os.FileMode(0644)
	// DefaultFileFlag set the default file flag
	DefaultFileFlag = os.O_RDWR | os.O_CREATE | os.O_APPEND

	// ErrInternal defined the internal error
	ErrInternal = errors.New("error internal")
	// ErrClosed defined write while ctx close
	ErrClosed = errors.New("error write on close")
	// ErrInvalidArgument defined the invalid argument
	ErrInvalidArgument = errors.New("error argument invalid")
)

// Manager used to trigger rolling event.
type Manager interface {
	// Fire will return a string channel
	// while the rolling event occoured, new file name will generate
	Fire() chan string
	// Close the Manager
	Close()
}

// RollingWriter implement the io writer
type RollingWriter interface {
	io.Writer
	Close() error
}

// Config give out the config for manager
type Config struct {
	// LogPath defined the full path of log file directory.
	// there comes out 2 different log file:
	//
	// 1. the current log
	//	log file path is located here:
	//	[LogPath]/[FileName].log
	//
	// 2. the tuncated log file
	//	the tuncated log file is backup here:
	//	[LogPath]/[FileName].log.[TimeTag]
	//  if compressed true
	//	[LogPath]/[FileName].log.gz.[TimeTag]
	//
	// NOTICE: blank field will be ignored
	// By default I use '-' as separator, you can set it yourself
	TimeTagFormat string `json:"time_tag_format"`
	LogPath       string `json:"log_path"`
	FileName      string `json:"file_name"`
	// MaxRemain will auto clear the roling file list, set 0 will disable auto clean
	MaxRemain int `json:"max_remain"`

	// RollingPolicy give out the rolling policy
	// We got 3 policies(actually, 2):
	//
	//	1. WithoutRolling: no rolling will happen
	//	2. TimeRolling: rolling by time
	//	3. VolumeRolling: rolling by file size
	RollingPolicy      int    `json:"rolling_ploicy"`
	RollingTimePattern string `json:"rolling_time_pattern"`
	RollingVolumeSize  string `json:"rolling_volume_size"`

	// Compress will compress log file with gzip
	Compress bool `json:"compress"`
	// Asynchronous enable the asynchronous write
	// by default the writer will be synchronous
	Asynchronous bool `json:"asynchronous"`
	// Lock enable the lock for writer, the writer will guarantee parallel safity
	// NOTICE: this will take effect only when writer is synchronous
	Lock bool `json:"lock"`
}

// NewDefaultConfig return the default config
func NewDefaultConfig() Config {
	return Config{
		LogPath:            "./log",
		TimeTagFormat:      "200601021504",
		FileName:           "log",
		MaxRemain:          -1,            // disable auto delete
		RollingPolicy:      1,             // TimeRotate by default
		RollingTimePattern: "0 0 0 * * *", // Rolling at 00:00 AM everyday
		RollingVolumeSize:  "1G",
		Compress:           false,
		Asynchronous:       false,
		Lock:               false,
	}
}

// LogFilePath return the absolute path on log file
func LogFilePath(c *Config) string {
	var filepath string
	filepath = strings.TrimRight(c.LogPath, "/") + "/" + c.FileName + ".log"
	return filepath
}

// Option defined config option
type Option func(*Config)

// WithTimeTagFormat set the TimeTag format string
func WithTimeTagFormat(format string) Option {
	return func(p *Config) {
		p.TimeTagFormat = format
	}
}

// WithLogPath set the log dir and auto create dir tree
// if the dir/path is not exist
func WithLogPath(path string) Option {
	return func(p *Config) {
		p.LogPath = path
	}
}

// WithFileName set the log file name
func WithFileName(name string) Option {
	return func(p *Config) {
		p.FileName = name
	}
}

// WithAsynchronous enable the asynchronous write for writer
func WithAsynchronous() Option {
	return func(p *Config) {
		p.Asynchronous = true
	}
}

// WithLock will enable the lock in writer
// Writer will call write with the Lock to guarantee the parallel safe
func WithLock() Option {
	return func(p *Config) {
		p.Lock = true
	}
}

// WithCompress will auto compress the tuncated log file with gzip
func WithCompress() Option {
	return func(p *Config) {
		p.Compress = true
	}
}

// WithMaxRemain enable the auto deletion for old file when exceed the given max value
// Bydefault -1 will disable the auto deletion
func WithMaxRemain(max int) Option {
	return func(p *Config) {
		p.MaxRemain = max
	}
}

// WithRollingPolicy set the rolling policy
func WithRollingPolicy(policy int) Option {
	return func(p *Config) {
		p.RollingPolicy = policy
	}
}

// WithRollingTimePattern set the time rolling policy time pattern obey the Corn table style
// visit http://crontab.org/ for details
func WithRollingTimePattern(pattern string) Option {
	return func(p *Config) {
		p.RollingTimePattern = pattern
	}
}

// WithRollingVolumeSize set the rolling file truncation threshold size
func WithRollingVolumeSize(size string) Option {
	return func(p *Config) {
		p.RollingVolumeSize = size
	}
}
