package dump

import (
	"bufio"
	"fmt"
	"path/filepath"
	"time"

	"github.com/distributedio/titan/conf"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
)

const (
	timeFormat      = "2006-01-02 15:04:05"
	defaultFileSize = 500 // megabytes
)

//Record redis command information
type Record struct {
	StartTime  time.Time
	Cost       float64
	TraceID    string
	Name       string
	Args       []string
	NameSpace  string
	RemoteAddr string
}

var (
	config    *conf.Server
	nsFileMap map[string]*bufio.Writer
	//RecordSeq all of redis command  transfer chan
	RecordSeq chan *Record
	//CloseDumpChan close dump chan
	CloseDumpChan chan bool
)

//InitDumpCommand init dump command
func InitDumpCommand(c *conf.Server) error {
	config = c
	RecordSeq = make(chan *Record, 1000000)
	nsFileMap = make(map[string]*bufio.Writer)
	CloseDumpChan = make(chan bool)
	go loop()
	zap.L().Info(fmt.Sprintf("init dump command successful."))
	return nil
}

func newFile(namespace string) *bufio.Writer {
	if config.DumpRedisCommandRotateSize == 0 {
		config.DumpRedisCommandRotateSize = defaultFileSize
	}
	return bufio.NewWriter(&lumberjack.Logger{
		Filename:   filepath.Join(config.DumpRedisCommand, namespace),
		MaxSize:    config.DumpRedisCommandRotateSize, // megabytes
		MaxBackups: 3000,
		MaxAge:     3000, //days
		Compress:   true, // disabled by default
	})
}

func loop() {
	for {
		select {
		case m := <-RecordSeq:
			writeCommand(m)
		case <-CloseDumpChan:
			stop()
			return
		}
	}
}

func writeCommand(m *Record) error {
	w, ok := nsFileMap[m.NameSpace]
	if !ok {
		w = newFile(m.NameSpace)
		nsFileMap[m.NameSpace] = w
	}
	_, err := w.WriteString(fmt.Sprintf("%s||%f||%s||%s||%s||%v\n",
		m.StartTime.Local().Format(timeFormat), m.Cost, m.RemoteAddr, m.TraceID, m.Name, m.Args))
	if err != nil {
		zap.L().Info(fmt.Sprintf("write dump log error %v", err))
	}
	return nil
}

func stop() error {
	close(RecordSeq)
	for _, f := range nsFileMap {
		f.Flush()
	}
	return nil
}

func init() {
}
