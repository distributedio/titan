package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	stub "github.com/arthurkiller/rollingWriter"
	"github.com/shafreeck/configo"
	"github.com/shafreeck/continuous"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gitlab.meitu.com/platform/thanos"
	"gitlab.meitu.com/platform/thanos/conf"
	"gitlab.meitu.com/platform/thanos/context"
	"gitlab.meitu.com/platform/thanos/db"
	"gitlab.meitu.com/platform/thanos/metrics"
)

func main() {
	var showVersion bool
	var confPath string

	flag.BoolVar(&showVersion, "v", false, "Show Version")
	flag.StringVar(&confPath, "c", "conf/thanos.toml", "conf file path")
	flag.Parse()

	if showVersion {
		thanos.PrintVersionInfo()
		return
	}

	config := &conf.Thanos{}
	if err := configo.Load(confPath, config); err != nil {
		fmt.Printf("unmarshal config file failed, %s\n", err)
		os.Exit(1)
	}

	writer, err := Writer(config.Logger.Path, config.Logger.TimeRotate, config.Logger.Compress)
	if err != nil {
		fmt.Printf("create logger write failed, %s\n", err)
		os.Exit(1)
	}

	if err := GlobalLogger(config.Logger.Level, config.Logger.Name, writer); err != nil {
		fmt.Printf("create logger failed, %s\n", err)
		os.Exit(1)
	}

	tlog := config.TikvLog
	if err := TikvLogrus(tlog.Path, tlog.Level, tlog.TimeRotate, tlog.Compress); err != nil {
		fmt.Printf("create tikv logger failed, %s\n", err)
		os.Exit(1)
	}

	store, err := db.Open(&config.Server.Tikv)
	if err != nil {
		zap.L().Fatal("open db failed", zap.Error(err))
		os.Exit(1)
	}

	svr := metrics.NewServer(&config.Status)

	serv := thanos.New(&context.ServerContext{
		RequirePass: config.Server.Auth,
		Store:       store,
	})

	cont := continuous.New(continuous.LoggerOutput(writer), continuous.PidFile(config.PIDFileName))
	if err := cont.AddServer(serv, &continuous.ListenOn{Network: "tcp", Address: config.Server.Listen}); err != nil {
		zap.L().Fatal("add thanos server failed:", zap.Error(err))
	}

	if err := cont.AddServer(svr, &continuous.ListenOn{Network: "tcp", Address: config.Status.Listen}); err != nil {
		zap.L().Fatal("add statues server failed:", zap.Error(err))
	}

	if err := cont.Serve(); err != nil {
		zap.L().Fatal("run server failed:", zap.Error(err))
	}
}

//Logger zap logger
func GlobalLogger(level, name string, write io.Writer) error {
	var lv = zap.NewAtomicLevel()
	switch level {
	case "debug":
		lv.SetLevel(zap.DebugLevel)
	case "info":
		lv.SetLevel(zap.InfoLevel)
	case "warn":
		lv.SetLevel(zap.WarnLevel)
	case "error":
		lv.SetLevel(zap.ErrorLevel)
	case "panic":
		lv.SetLevel(zap.PanicLevel)
	case "fatal":
		lv.SetLevel(zap.FatalLevel)
	default:
		return fmt.Errorf("unknown log level(%s)\n", level)
	}
	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Local().Format("2006-01-02 15:04:05.999999999"))
	}

	encoderCfg := zapcore.EncoderConfig{
		NameKey:        "Name",
		StacktraceKey:  "Stack",
		MessageKey:     "Message",
		LevelKey:       "Level",
		TimeKey:        "TimeStamp",
		CallerKey:      "Caller",
		EncodeTime:     timeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	output := zapcore.AddSync(write)
	var zapOpts []zap.Option
	zapOpts = append(zapOpts, zap.AddCaller())
	zapOpts = append(zapOpts, zap.Hooks(metrics.Measure))

	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), output, lv), zapOpts...)
	logger.Named(name)
	log := logger.With(zap.Int("PID", os.Getpid()))
	zap.ReplaceGlobals(log)
	//http change log level
	http.Handle("/thanos/log/level", lv)

	return nil
}

//TikvLogrus deal tikv log
func TikvLogrus(path, level, pattern string, compress bool) error {
	writer, err := Writer(path, pattern, compress)
	if err != nil {
		return err
	}
	logrus.SetOutput(writer)
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		return fmt.Errorf("unknown log level(%s)\n", level)
	}
	return nil
}

//Writer generate the rollingWriter
func Writer(path, pattern string, compress bool) (io.Writer, error) {
	var opts []stub.Option
	opts = append(opts, stub.WithRollingTimePattern(pattern))
	if compress {
		opts = append(opts, stub.WithCompress())
	}
	opts = append(opts, stub.WithLogPath(path))
	writer, err := stub.NewWriter(opts...)
	if err != nil {
		return nil, fmt.Errorf("create IOWriter failed, %s", err)
	}
	return writer, nil
}
