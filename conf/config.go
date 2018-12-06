package conf

import "time"

//Titan configuration center
type Titan struct {
	Server      Server     `cfg:"server"`
	Status      Status     `cfg:"status"`
	TikvLog     TikvLogger `cfg:"tikv-logger"`
	Logger      Logger     `cfg:"logger"`
	PIDFileName string     `cfg:"pid-filename; titan.pid; ; the file name to record connd PID"`
}

//Server config is the config of titan server
type Server struct {
	Tikv             Tikv   `cfg:"tikv"`
	Auth             string `cfg:"auth;;;client connetion auth"`
	Listen           string `cfg:"listen; 0.0.0.0:7369; netaddr; address to listen"`
	MaxConnection    int64  `cfg:"max-connection;1000;numeric;client connection count"`
	ListZipThreshold int64  `cfg:"list-zip-threshold;100;numeric;the max limit length of elements in list"`
}

//Tikv config is the config of tikv sdk
type Tikv struct {
	PdAddrs string `cfg:"pd-addrs;required; ;pd address in tidb"`
	ZT      ZT     `cfg:"zt"`
}

//ZT config is the config of zlist
type ZT struct {
	Wrokers    int           `cfg:"workers;5;numeric;parallel workers count"`
	BatchCount int           `cfg:"batch;10;numeric;object transfer limitation per-transection"`
	QueueDepth int           `cfg:"depth;100;numeric;ZT Worker queue depth"`
	Interval   time.Duration `cfg:"interval;1000ms; ;Queue fill interval in milsecond"`
}

//Logger config is the config of default zap log
type Logger struct {
	Name       string `cfg:"name; titan; ; the default logger name"`
	Path       string `cfg:"path; logs/titan; ; the default log path"`
	Level      string `cfg:"level; info; ; log level(debug, info, warn, error, panic, fatal)"`
	Compress   bool   `cfg:"compress; false; boolean; true for enabling log compress"`
	TimeRotate string `cfg:"time-rotate; 0 0 0 * * *; ; log time rotate pattern(s m h D M W)"`
}

//TikvLogger config is the config of tikv log
type TikvLogger struct {
	Path       string `cfg:"path; logs/tikv;nonempty ; the default log path"`
	Level      string `cfg:"level; info; ; log level(debug, info, warn, error, panic, fatal)"`
	Compress   bool   `cfg:"compress; false; boolean; true for enabling log compress"`
	TimeRotate string `cfg:"time-rotate; 0 0 0 * * *; ; log time rotate pattern(s m h D M W)"`
}

//Status config is the config of exported server
type Status struct {
	Listen string `cfg:"listen;0.0.0.0:7345;nonempty; listen address of http server"`
}
