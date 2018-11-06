package bunnystub

import (
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/robfig/cron"
)

type RotatePolicy int

const (
	WithoutRotate = iota
	TimeRotate
	VolumeRotate
)

var (
	// Precision defined the precision about how many SECONDS will be waitted before
	// the reopen operation check the condition
	Precision = 1
)

type Manager interface {
	Enable() chan string
	WriteN(int)
	ResetFileSize()
	LockFree() bool
	NameParts() []string
	FileMode() os.FileMode
	Compress() bool
	IgnoreOK() bool
	Concurency() int
	Spining() bool
}
type Option func(*IOManager)

func WithConcurrency(concurency int, spining bool) Option {
	return func(p *IOManager) {
		p.concurency = concurency
		p.spining = spining
	}
}

func WithTimePattern(pattern string) Option {
	return func(p *IOManager) {
		p.RollingTimePattern = pattern
	}
}
func WithVolumeSize(size string) Option {
	return func(p *IOManager) {
		p.RollingVolumeSize = size
	}
}
func WithPath(path string) Option {
	return func(p *IOManager) {
		paths := strings.Split(path, "/")
		switch len(paths) {
		case 0:
		case 1:
			ss := strings.Split(paths[0], ".")
			switch len(ss) {
			case 0:
			case 1:
				p.Prefix = ss[0]
			default:
				p.Suffix = "." + ss[len(ss)-1]
				p.Prefix = strings.Join(ss[:len(ss)-1], ".")
			}
		default:
			p.FilePath = strings.Join(paths[:len(paths)-1], "/") + "/"
			ss := strings.Split(paths[len(paths)-1], ".")
			switch len(ss) {
			case 0:
			case 1:
				p.Prefix = ss[0]
			default:
				p.Suffix = "." + ss[len(ss)-1]
				p.Prefix = strings.Join(ss[:len(ss)-1], ".")
			}
		}
	}
}
func WithDir(dir string) Option {
	return func(p *IOManager) {
		dir = strings.TrimSuffix(dir, "/")
		dir = dir + "/"
		p.FilePath = dir
	}
}
func WithPrefix(prefix string) Option {
	return func(p *IOManager) {
		p.Prefix = prefix
	}
}
func WithSuffix(suffix string) Option {
	return func(p *IOManager) {
		p.Suffix = suffix
	}
}
func WithFileMode(mode uint32) Option {
	return func(p *IOManager) {
		p.FileMod = os.FileMode(mode)
	}
}
func WithIgnoreOK() Option {
	return func(p *IOManager) {
		p.IsIgnoreOK = true
	}
}
func WithLockFree() Option {
	return func(p *IOManager) {
		p.IsLockFree = true
	}
}
func WithCompress() Option {
	return func(p *IOManager) {
		p.IsCompress = true
	}
}

type IOManager struct {
	FilePath string
	// file name js like this style: prefix-timestamp-suffix.log
	// compressed log file is named like this: prefix-timestamp-suffix.tar
	Prefix     string
	Suffix     string
	FileMod    os.FileMode
	IsIgnoreOK bool
	IsCompress bool
	IsLockFree bool

	concurency int
	spining    bool

	// pattern is just like the crontable style without year, second minute hour day mounth weekday
	RollingTimePattern string
	// VolumeSize can be give a size with K M G
	RollingVolumeSize string
	// TimeFormatPattern defined the rolling time pattern which will used to name the .log file
	TimeFormatPattern string

	// return the file name for rename
	enable chan string

	// check the condition
	trigger func()

	//rolling family var takes the rolling checking condition
	rollingPoint  time.Time
	rollingVolume int64

	cr       *cron.Cron
	fileSize int64
	// use to stop the writer
	terminate chan int
}

func newManager(ops ...Option) *IOManager {
	m := &IOManager{
		FileMod:            os.FileMode(0644),
		RollingTimePattern: "0 0 * * * *",
		RollingVolumeSize:  "1G",
		TimeFormatPattern:  "20060102",
		enable:             make(chan string),
		IsLockFree:         false,
		IsIgnoreOK:         false,
		cr:                 cron.New(),
	}
	for _, o := range ops {
		o(m)
	}
	m.rollingPoint = time.Now()
	return m
}

func (m *IOManager) Enable() chan string   { return m.enable }
func (m *IOManager) WriteN(n int)          { atomic.AddInt64(&m.fileSize, int64(n)) }
func (m *IOManager) ResetFileSize()        { atomic.SwapInt64(&m.fileSize, 0) }
func (m *IOManager) LockFree() bool        { return m.IsLockFree }
func (m *IOManager) IgnoreOK() bool        { return m.IsIgnoreOK }
func (m *IOManager) Compress() bool        { return m.IsCompress }
func (m *IOManager) NameParts() []string   { return []string{m.FilePath, m.Prefix, m.Suffix} }
func (m *IOManager) FileMode() os.FileMode { return m.FileMod }
func (m *IOManager) Concurency() int       { return m.concurency }
func (m *IOManager) Spining() bool         { return m.spining }

func (m *IOManager) parseVolume() {
	s := []byte(strings.ToUpper(m.RollingVolumeSize))

	var p int64 = 0
	var unit int64 = 1
	var unitstr string
	if s[len(s)-1] == 'B' {
		tp, _ := strconv.Atoi(string(s[:len(s)-2]))
		p = int64(tp)
		unitstr = string(s[len(s)-2:])
	} else {
		tp, _ := strconv.Atoi(string(s[:len(s)-1]))
		p = int64(tp)
		unitstr = string(s[len(s)-1])
	}

	switch unitstr {
	case "G", "GB":
		unit *= 1024
		fallthrough
	case "M", "MB":
		unit *= 1024
		fallthrough
	case "K", "KB":
		unit *= 1024
	default:
		m.rollingVolume = 1024 * 1024 * 1024
		return
	}

	m.rollingVolume = p * unit
}

func NewTimeRotateManager(ops ...Option) Manager {
	m := newManager(ops...)
	m.trigger = func() {}
	m.cr.AddFunc(m.RollingTimePattern, func() {
		m.enable <- m.FilePath + m.Prefix + m.Suffix + "." + m.rollingPoint.Format(m.TimeFormatPattern)
		m.rollingPoint = time.Now()
	})
	m.cr.Start()
	return m
}

func NewVolumeRotateManager(ops ...Option) Manager {
	m := newManager(ops...)
	m.parseVolume()
	m.trigger = func() {
		if atomic.LoadInt64(&m.fileSize) > m.rollingVolume {
			m.enable <- m.FilePath + m.Prefix + m.Suffix + "." + m.rollingPoint.Format(m.TimeFormatPattern)
			m.rollingPoint = time.Now()
		}
	}
	return m
}
