package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	stub "github.com/arthurkiller/rollingWriter"
	"github.com/shafreeck/configo"
	"github.com/sirupsen/logrus"

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
	//TODO  conf path
	flag.StringVar(&confPath, "c", "../../conf/thanos.toml", "conf file path")
	flag.Parse()

	if showVersion {
		//TODO
		// server.PrintVersionInfo()
		return
	}

	config := &conf.Thanos{}
	if err := configo.Load(confPath, config); err != nil {
		fmt.Printf("unmarshal config file failed, %s\n", err)
		os.Exit(1)
	}

	/*
		debug, err := Logger(&config.DebugLog)
		if err != nil {
			fmt.Printf("create logger failed, %s\n", err)
			os.Exit(1)
		}
	*/

	tlog := config.TikvLog
	if err := CreateLogrus(tlog.LogPath, tlog.LogLevel, tlog.LogTimeRotate, tlog.LogCompress); err != nil {
		fmt.Printf("create tikv logger failed, %s\n", err)
		os.Exit(1)
	}

	// silent the tikv log message
	store, err := db.Open(&config.Server.Tikv)
	if err != nil {
		log.Fatalln(err)
	}

	msvr := metrics.NewServer(&config.Status)

	serv := thanos.New(&context.ServerContext{
		RequirePass: config.Server.Auth,
		Store:       store,
	})

	// cont := continuous.New(continuous.UseLogger(debug), continuous.PidFile(config.PIDFileName))

	if err := serv.ListenAndServe(config.Server.Listen); err != nil {
		log.Fatalln(err)
	}

	if err := msvr.ListenAndServe(config.Status.Listen); err != nil {
		log.Fatalln(err)
	}
}

/*
//TODO zap logger
func CreateLogger(path, level, pattern, name string, compress bool) (log.Logger, error) {

		// create custom log handler for connd
		var wopts []stub.Option
		wopts = append(wopts, stub.WithTimePattern(pattern))
		if compress {
			wopts = append(wopts, stub.WithCompress())
		}

		writer, err := stub.NewIOWriter(path, stub.TimeRotate, wopts...)
		if err != nil {
			return nil, fmt.Errorf("create IOWriter failed, %s", err)
		}

		var options []log.Option
		switch level {
		case "debug":
			options = append(options, log.WithDebugLevel())
		case "info":
			options = append(options, log.WithInfoLevel())
		case "warn":
			options = append(options, log.WithWarnLevel())
		case "error":
			options = append(options, log.WithErrorLevel())
		case "panic":
			options = append(options, log.WithPanicLevel())
		case "fatal":
			options = append(options, log.WithFatalLevel())
		default:
			return nil, fmt.Errorf("unknown log level(%s)\n", level)
		}
		options = append(options, log.WithOutput(writer), log.WithCaller(), log.WithMetrics(), log.WithName(name))

		logger, err := log.New(options...)
		if err != nil {
			return nil, fmt.Errorf("init log failed, %s", err)
		}
		return logger.With(log.Int("PID", os.Getpid())), nil

	return nil, nil
}
*/

func CreateLogrus(path, level, pattern string, compress bool) error {
	var wopts []stub.Option
	wopts = append(wopts, stub.WithRollingTimePattern(pattern))
	if compress {
		wopts = append(wopts, stub.WithCompress())
	}
	wopts = append(wopts, stub.WithLogPath(path))

	writer, err := stub.NewWriter(wopts...)
	if err != nil {
		return fmt.Errorf("create IOWriter failed, %s", err)
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
