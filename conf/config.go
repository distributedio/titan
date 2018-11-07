package conf

import "time"

type Thanos struct {
	Server      Server     `cfg:"server"`
	Status      Status     `cfg:"status"`
	TikvLog     TikvLogger `cfg:"tikv-logger"`
	Logger      Logger     `cfg:"logger"`
	PIDFileName string     `cfg:"pid-filename; thanos.pid; ; the file name to record connd PID"`
}

// Config is the config of titan server
type Server struct {
	Tikv          Tikv   `cfg:"tikv"`
	Auth          string `cfg:"auth;"";;client connetion auth"`
	Listen        string `cfg:"listen; 0.0.0.0:7369; netaddr; address to listen"`
	MaxConnection int64  `cfg:"max-connection;1000;numeric;client connection count"`
}

type Tikv struct {
	PdAddrs string `cfg:"pd-addrs;required; nonempty;pd address in tidb"`
	ZT      ZT     `cfg:"zt"`
}

type ZT struct {
	Wrokers    int           `cfg:"workers;5;numeric;parallel workers count"`
	BatchCount int           `cfg:"batch;10;numeric;object transfer limitation per-transection"`
	QueueDepth int           `cfg:"depth;100;numeric;ZT Worker queue depth"`
	Interval   time.Duration `cfg:"interval;1000ms; ;Queue fill interval in milsecond"`
}

type Logger struct {
	Name       string `cfg:"name; thanos; ; the default logger name"`
	Path       string `cfg:"path; logs/thanos.log; ; the default log path"`
	Level      string `cfg:"level; info; ; log level(debug, info, warn, error, panic, fatal)"`
	Compress   bool   `cfg:"compress; false; boolean; true for enabling log compress"`
	TimeRotate string `cfg:"time-rotate; 0 0 0 * * *; ; log time rotate pattern(s m h D M W)"`
}

type TikvLogger struct {
	Path       string `cfg:"path; logs/tikv.log;nonempty ; the default log path"`
	Level      string `cfg:"level; info; ; log level(debug, info, warn, error, panic, fatal)"`
	Compress   bool   `cfg:"compress; false; boolean; true for enabling log compress"`
	TimeRotate string `cfg:"time-rotate; 0 0 0 * * *; ; log time rotate pattern(s m h D M W)"`
}

type Status struct {
	Listen string `cfg:"listen;0.0.0.0:7345;nonempty; listen address of http server"`
}
