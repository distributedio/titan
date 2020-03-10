package conf

import "time"

// Titan configuration center
type Titan struct {
	Server      Server `cfg:"server"`
	Status      Status `cfg:"status"`
	TiKV        TiKV   `cfg:"tikv"`
	Logger      Logger `cfg:"logger"`
	PIDFileName string `cfg:"pid-filename; titan.pid; ; the file name to record connd PID"`
}

// Server config is the config of titan server
type Server struct {
	Auth                       string `cfg:"auth;;;client connetion auth"`
	Listen                     string `cfg:"listen; 0.0.0.0:7369; netaddr; address to listen"`
	SSLCertFile                string `cfg:"ssl-cert-file;;;server SSL certificate file (enables SSL support)"`
	SSLKeyFile                 string `cfg:"ssl-key-file;;;server SSL key file"`
	MaxConnection              int64  `cfg:"max-connection;1000;numeric;client connection count"`
	ListZipThreshold           int    `cfg:"list-zip-threshold;100;numeric;the max limit length of elements in list"`
	DumpRedisCommand           string `cfg:"dump-redis-command;./;;dump redis command and duration to file"`
	DumpRedisCommandRotateSize int    `cfg:"dump-redis-command-rotate-size; 500;;dump redis command file rotate size"`
}

// TiKV config is the config of tikv sdk
type TiKV struct {
	PdAddrs string     `cfg:"pd-addrs; mocktikv://; ;pd address in tidb"`
	GC      GC         `cfg:"gc"`
	Expire  Expire     `cfg:"expire"`
	ZT      ZT         `cfg:"zt"`
	TiKVGC  TiKVGC     `cfg:"tikv-gc"`
	Logger  TiKVLogger `cfg:"logger"`
}

// TiKVGC config is the config of implement tikv sdk gcwork
type TiKVGC struct {
	Disable           bool          `cfg:"disable; false; boolean; false is used to disable tikvgc "`
	Interval          time.Duration `cfg:"interval;20m;;gc work tick interval"`
	LeaderLifeTime    time.Duration `cfg:"leader-life-time;30m;;lease flush leader interval"`
	SafePointLifeTime time.Duration `cfg:"safe-point-life-time;10m;;safe point life time "`
	Concurrency       int           `cfg:"concurrency;2;;gc work concurrency"`
}

// GC config is the config of Titan GC work
type GC struct {
	Disable        bool          `cfg:"disable; false; boolean; false is used to disable gc"`
	Interval       time.Duration `cfg:"interval;1s;;gc work tick interval"`
	LeaderLifeTime time.Duration `cfg:"leader-life-time;3m;;lease flush leader interval"`
	BatchLimit     int           `cfg:"batch-limit;256;numeric;key count limitation per-transection"`
}

// Expire config is the config of Titan expire work
type Expire struct {
	Disable        bool          `cfg:"disable; false; boolean; false is used to disable expire"`
	Interval       time.Duration `cfg:"interval;1s;;expire work tick interval"`
	LeaderLifeTime time.Duration `cfg:"leader-life-time;3m;;lease flush leader interval"`
	BatchLimit     int           `cfg:"batch-limit;256;numeric;key count limitation per-transection"`
}

// ZT config is the config of zlist
type ZT struct {
	Disable    bool          `cfg:"disable; false; boolean; false is used to disable  zt"`
	Workers    int           `cfg:"workers;5;numeric;parallel workers count"`
	BatchCount int           `cfg:"batch;10;numeric;object transfer limitation per-transection"`
	QueueDepth int           `cfg:"depth;100;numeric;ZT Worker queue depth"`
	Interval   time.Duration `cfg:"interval;1000ms; ;Queue fill interval in milsecond"`
}

// Logger config is the config of default zap log
type Logger struct {
	Name       string `cfg:"name; titan; ; the default logger name"`
	Path       string `cfg:"path; logs/titan; ; the default log path"`
	Level      string `cfg:"level; info; ; log level(debug, info, warn, error, panic, fatal)"`
	Compress   bool   `cfg:"compress; false; boolean; true for enabling log compress"`
	TimeRotate string `cfg:"time-rotate; 0 0 0 * * *; ; log time rotate pattern(s m h D M W)"`
}

// TiKVLogger config is the config of tikv log
type TiKVLogger struct {
	Path       string `cfg:"path; logs/tikv;nonempty ; the default log path"`
	Level      string `cfg:"level; info; ; log level(debug, info, warn, error, panic, fatal)"`
	Compress   bool   `cfg:"compress; false; boolean; true for enabling log compress"`
	TimeRotate string `cfg:"time-rotate; 0 0 0 * * *; ; log time rotate pattern(s m h D M W)"`
}

// Status config is the config of exported server
type Status struct {
	Listen      string `cfg:"listen;0.0.0.0:7345;nonempty; listen address of http server"`
	SSLCertFile string `cfg:"ssl-cert-file;;;status server SSL certificate file (enables SSL support)"`
	SSLKeyFile  string `cfg:"ssl-key-file;;;status server SSL key file"`
}
