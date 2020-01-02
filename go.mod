module github.com/distributedio/titan

require (
	github.com/arthurkiller/rollingwriter v1.1.2
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.18+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/cznic/sortutil v0.0.0-20181122101858-f5f958428db8 // indirect
	github.com/distributedio/configo v0.0.0-20190610140513-0d38d0d8590a
	github.com/distributedio/continuous v0.0.0-20190527021358-1768e41f22b9
	github.com/facebookgo/ensure v0.0.0-20160127193407-b4ab57deab51 // indirect
	github.com/facebookgo/freeport v0.0.0-20150612182905-d4adf43b75b9 // indirect
	github.com/facebookgo/grace v0.0.0-20180706040059-75cf19382434 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20150612182917-8dac2c3c4870 // indirect
	github.com/golang/groupcache v0.0.0-20191227052852-215e87163ea7 // indirect
	github.com/golang/protobuf v1.2.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/myesui/uuid v1.0.0 // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pingcap/check v0.0.0-20191216031241-8a5a85928f12 // indirect
	github.com/pingcap/goleveldb v0.0.0-20191226122134-f82aafb29989 // indirect
	github.com/pingcap/kvproto v0.0.0-20200102065152-5d51d93be892
	github.com/pingcap/tidb v1.1.0-beta.0.20191227152506-8f13cf1449bd
	github.com/pingcap/tipb v0.0.0-20191230123656-568726749cb7 // indirect
	github.com/prometheus/client_golang v0.9.2
	github.com/robfig/cron v1.2.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/shafreeck/retry v0.0.0-20180827080527-71c8c3fbf8f8
	github.com/sirupsen/logrus v1.3.0
	github.com/stretchr/testify v1.4.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/twinj/uuid v1.0.0
	go.etcd.io/etcd v3.3.18+incompatible // indirect
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/zap v1.12.0
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gopkg.in/stretchr/testify.v1 v1.0.0-00010101000000-000000000000 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

go 1.13

replace gopkg.in/stretchr/testify.v1 => github.com/stretchr/testify v1.2.2
