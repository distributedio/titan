package integration

import (
	"fmt"
	"log"
	"net"

	"gitlab.meitu.com/platform/thanos"
	"gitlab.meitu.com/platform/thanos/conf"
	"gitlab.meitu.com/platform/thanos/context"
	"gitlab.meitu.com/platform/thanos/db"
)

var (
	svr *thanos.Server
	cfg = &conf.Server{
		Listen:        ServerAddr,
		MaxConnection: 10000,
		Auth:          "",
		Tikv: conf.Tikv{
			PdAddrs: "mocktikv://",
		},
	}

	//ServerAddr default server addr
	ServerAddr = "127.0.0.1:17369"
	lis        net.Listener
)

//SetAuth default no verify
// specify auth to enable validation
func SetAuth(auth string) {
	cfg.Auth = auth
}

//Start start server
//1.open db
//2.start server fd
func Start() {
	var err error
	store, err := db.Open(&cfg.Tikv)
	if err != nil {
		log.Fatalln(err)
	}

	svr = thanos.New(&context.ServerContext{
		RequirePass: cfg.Auth,
		Store:       store,
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
