package integration

import (
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"

	"github.com/distributedio/titan"
	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/context"
	"github.com/distributedio/titan/db"
)

var (
	svr *titan.Server
	cfg = &conf.Server{
		Listen:        ServerAddr,
		MaxConnection: 10000,
		Auth:          "",
	}
	tikvConf = conf.MockConf().Tikv
	//ServerAddr default server addr
	ServerAddr = "127.0.0.1:17369"
	lis        net.Listener
)

//SetAuth default no verify
// specify auth to enable validation
func SetAuth(auth string) {
	cfg.Auth = auth
}

// SetAddr set server listen addr
func SetAddr(addr string) {
	cfg.Listen = addr
}

//Start start server
//1.open db
//2.start server fd
func Start() {
	zap.ReplaceGlobals(zap.NewNop())
	var err error
	store, err := db.Open(&tikvConf)
	if err != nil {
		log.Fatalln(err)
	}

	limitersMgr, err := db.NewLimitersMgr(store, &tikvConf.RateLimit)
	if err != nil {
		log.Fatalln(err)
	}
	svr = titan.New(&context.ServerContext{
		RequirePass:      cfg.Auth,
		Store:            store,
		ListZipThreshold: 100,
		LimitersMgr:      limitersMgr,
	})
	err = svr.ListenAndServe(cfg.Listen)
	if err != nil {
		return
	}
}

//Close close server listen fd
func Close() {
	if err := svr.Stop(); err != nil {
		fmt.Println(err)
	}
}
