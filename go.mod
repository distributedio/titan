module github.com/distributedio/titan

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/arthurkiller/rollingwriter v1.1.2
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.0.0 // indirect
	github.com/coreos/license-bill-of-materials v0.0.0-20190913234955-13baff47494e // indirect
	github.com/cznic/sortutil v0.0.0-20181122101858-f5f958428db8 // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/distributedio/configo v0.0.0-20200107073829-efd79b027816
	github.com/distributedio/continuous v0.0.0-20190527021358-1768e41f22b9
	github.com/facebookgo/freeport v0.0.0-20150612182905-d4adf43b75b9 // indirect
	github.com/facebookgo/grace v0.0.0-20180706040059-75cf19382434 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
	github.com/golang/protobuf v1.3.4
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/myesui/uuid v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pingcap/goleveldb v0.0.0-20191226122134-f82aafb29989 // indirect
	github.com/pingcap/kvproto v0.0.0-20201208043834-923c9609272c
	github.com/pingcap/tidb v1.1.0-beta.0.20201210112752-c33e90a7aef4
	github.com/prometheus/client_golang v1.5.1
	github.com/robfig/cron v1.2.0 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/shafreeck/retry v0.0.0-20180827080527-71c8c3fbf8f8
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/struCoder/pidusage v0.1.2 // indirect
	github.com/tikv/pd v1.1.0-beta.0.20201125070607-d4b90eee0c70
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/twinj/uuid v1.0.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200824191128-ae9734ed278b
	go.uber.org/zap v1.16.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/stretchr/testify.v1 v1.0.0-00010101000000-000000000000 // indirect
)

go 1.13

replace gopkg.in/stretchr/testify.v1 => github.com/stretchr/testify v1.2.2
