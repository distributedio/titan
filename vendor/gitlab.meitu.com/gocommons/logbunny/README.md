# logbunny 🐰
Logbunny is a go log framework mixed with serval popular logger. 
It is designed to take place of the slow old-fashion seelog.
It's so powerful quick and flexible that everyone can't just believe it is called bunny ?!?

## Feature
* 支持__zap__ 和 __logrus__ 两种log
* log级别动态调整, 提供一个RESTful API 可以用于通过 PUT 修改level
* 支持log分级别多输出。可以输出到任何实现了io.Writer的位置
* 相同实现的log可以相互Tee。即，将两个log合成一个log来用
* 支持生成GrpcLogger及GrpcLoggerV2

## Simple Benchmark for logbunny
这里提供了一个简单的benchmnark对比两种不同的log的实现，同时为了更加直观，我们也对gocommon的seelog做了benchmark。
更多的测试可以查阅docs下的文档。

test case | testing times | cost on per-operation | allocation on per-option | allocation times on per-option
----------|---------------|-----------------------|--------------------------|-------------------------------
BenchmarkBunnyZapLoggerDebug-4    | 1000000  | 1809 ns/op      |    232 B/op |     3 allocs/op
BenchmarkBunnyZapLoggerError-4    | 1000000  | 1785 ns/op      |    232 B/op |     3 allocs/op
BenchmarkBunnyLogrusLoggerDebug-4 | 200000   | 6994 ns/op      |    2202 B/op|    37 allocs/op
BenchmarkBunnyLogrusLoggerError-4 | 200000   | 6934 ns/op      |    2202 B/op|    37 allocs/op
BenchmarkGocommonError-4  |    100000        | 14404 ns/op     |    576 B/op |    19 allocs/op
BenchmarkGocommonDebug-4  |    200000        | 12720 ns/op     |    549 B/op |    16 allocs/op

## Quick Start
使用方式及其简单
```go
    import log "gitlab.meitu.com/gocommons/logbunny"

    log.Debug("demo",log.String("foo","bar"))

```

```go
    import log "gitlab.meitu.com/gocommons/logbunny"

    zaplogger, err := log.New(log.WithZapLogger(), log.WithDebugLevel(), log.WithJSONEncoder())

    log.Logger = zaplogger

    log.Debug("demo",log.String("foo","bar"))
```
    
简单的例子还不爽？可以看看[demo](https://gitlab.meitu.com/gocommons/logbunny/blob/dev/demo/client.go)下的例子！详细的介绍了所有功能的使用方法！

## Testing (TODO 目前达不到96%了)
logbunny 采用开发和单元测试相结合的方式，目前的测试覆盖率基本达到 __96%__
__test case total coverage: 95.8% __

| filename | function | coverage% |
|----------|----------|-----------|
| logbunny/config.go:46:              |  NewConfig     |  100.0%                    |
| logbunny/config.go:68:              |  WithCaller    |  100.0%                    |
| logbunny/config.go:69:              |  WithZapLogger    |   100.0%                |
| logbunny/config.go:70:              |  WithLogrusLogger |   100.0%                |
| logbunny/config.go:71:              |  WithJSONEncoder |    100.0%                |
| logbunny/config.go:72:              |  WithTextEncoder |    100.0%                |
| logbunny/config.go:73:              |  WithDebugLevel  |    100.0%                |
| logbunny/config.go:74:              |  WithInfoLevel   |    100.0%                |
| logbunny/config.go:75:              |  WithWarnLevel   |    100.0%                |
| logbunny/config.go:76:              |  WithErrorLevel  |    100.0%                |
| logbunny/config.go:77:              |  WithPanicLevel  |    100.0%                |
| logbunny/config.go:78:              |  WithFatalLevel  |    100.0%                |
| logbunny/field.go:55:               |  Skip      |      100.0%                    |
| logbunny/field.go:63:               |  Base64    |      100.0%                    |
| logbunny/field.go:73:               |  Bool      |      100.0%                    |
| logbunny/field.go:87:               |  Float64   |      100.0%                    |
| logbunny/field.go:96:               |  Int       |  100.0%                        |
| logbunny/field.go:105:              |  Int64     |      100.0%                    |
| logbunny/field.go:114:              |  Uint      |      100.0%                    |
| logbunny/field.go:123:              |  Uint64    |      100.0%                    |
| logbunny/field.go:132:              |  Uintptr    |     100.0%                    |
| logbunny/field.go:141:              |  String     |     100.0%                    |
| logbunny/field.go:150:              |  Stringer   |     100.0%                    |
| logbunny/field.go:161:              |  Time       |     100.0%                    |
| logbunny/field.go:170:              |  Error      |     100.0%                    |
| logbunny/field.go:180:              |  Duration   |     100.0%                    |
| logbunny/field.go:192:              |  Marshaler  |     100.0%                    |
| logbunny/field.go:204:              |  Object     |     100.0%                    |
| logbunny/levelHandler.go:42:        |      NewLogrusLevelHandler  | 100.0%        |
| logbunny/levelHandler.go:46:        |      Set         |100.0%                    |
| logbunny/levelHandler.go:65:        |      Get         |100.0%                    |
| logbunny/levelHandler.go:84:        |      ServeHTTP   |    100.0%                |
| logbunny/levelHandler.go:95:        |      getLevel    |    100.0%                |
| logbunny/levelHandler.go:117:       |      putLevel    |    100.0%                |
| logbunny/levelHandler.go:161:       |      error       |    100.0%                |
| logbunny/levelHandler.go:173:       |      NewZapLevelHandler | 100.0%            |
| logbunny/levelHandler.go:178:       |      Set         |100.0%                    |
| logbunny/levelHandler.go:197:       |      Get         |87.5%                     |
| logbunny/levelHandler.go:216:       |      ServeHTTP   |    100.0%                |
| logbunny/log.go:24:                 |  internalError   |    100.0%                |
| logbunny/log.go:42:                 |  New        | 100.0%                        |
| logbunny/log.go:58:                 |  Tee        | 94.7%                         |
| logbunny/log.go:91:                 |  FilterLogger    |    92.0%                 |
| logbunny/log.go:140:                |  newZapLogger    |    100.0%                |
| logbunny/log.go:188:                |  newLogrusLogger |    100.0%                |
| logbunny/logrus.go:24:              |  SetLevel        |100.0%                    |
| logbunny/logrus.go:31:              |  AddCaller       |85.7%                     |
| logbunny/logrus.go:44:              |  newLogrusSplitLogger   | 100.0%            |
| logbunny/logrus.go:87:              |  log        | 82.6%                         |
| logbunny/logrus.go:135:             |  Debug      |     100.0%                    |
| logbunny/logrus.go:139:             |  Info       |     100.0%                    |
| logbunny/logrus.go:143:             |  Warn       |     100.0%                    |
| logbunny/logrus.go:147:             |  Error      |     100.0%                    |
| logbunny/logrus.go:151:             |  Panic      |     100.0%                    |
| logbunny/logrus.go:155:             |  Fatal      |     0.0%                      |
| logbunny/logrus.go:161:             |  genField   |     100.0%                    |
| logbunny/logrusLevelLogger.go:11:    |newLogrusTeeLogger|  66.7%                  |
| logbunny/logrusLevelLogger.go:24:    |Fire           | 88.9%                      |
| logbunny/logrusLevelLogger.go:63:    |Levels         | 100.0%                     |
| logbunny/logrusLevelLogger.go:75:    |newLogrusLevelLogger  |  90.9%              |
| logbunny/logrusLevelLogger.go:111:   |Fire           | 100.0%                     |
| logbunny/logrusLevelLogger.go:120:   |Levels         | 100.0%                     |
| logbunny/logrusLevelLogger.go:130:   |Fire           | 100.0%                     |
| logbunny/logrusLevelLogger.go:139:   |Levels         | 100.0%                     |
| logbunny/logrusLevelLogger.go:149:   |Fire           | 100.0%                     |
| logbunny/logrusLevelLogger.go:158:   |Levels         | 100.0%                     |
| logbunny/logrusLevelLogger.go:168:   |Fire           | 100.0%                     |
| logbunny/logrusLevelLogger.go:177:   |Levels         | 100.0%                     |
| logbunny/logrusLevelLogger.go:187:   |Fire           | 0.0%                       |
| logbunny/logrusLevelLogger.go:196:   |Levels         | 100.0%                     |
| logbunny/logrusLevelLogger.go:206:   |Fire           | 100.0%                     |
| logbunny/logrusLevelLogger.go:215:   |Levels         | 100.0%                     |
| logbunny/zap.go:18:          |SetLevel        | 100.0%                            |
| logbunny/zap.go:25:          |newZapSplitLogger|    100.0%                        |
| logbunny/zap.go:116:         |Debug         |  100.0%                             |
| logbunny/zap.go:121:         |Info          |  100.0%                             |
| logbunny/zap.go:126:         |Warn          |  100.0%                             |
| logbunny/zap.go:131:         |Error         |  100.0%                             |
| logbunny/zap.go:136:         |Panic         |  100.0%                             |
| logbunny/zap.go:141:         |Fatal         |  0.0%                               | 
| logbunny/zap.go:147:         |zapFields     |  100.0%                             |

## Issue
Any bugs founded or feature in need, just open up issue
