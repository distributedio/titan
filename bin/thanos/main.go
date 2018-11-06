package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"gitlab.meitu.com/platform/thanos"
	"gitlab.meitu.com/platform/thanos/context"
	"gitlab.meitu.com/platform/thanos/db"
)

func main() {
	// silent the tikv log message
	logrus.SetOutput(ioutil.Discard)

	addr := os.Args[1]

	store, err := db.Open(addr)
	if err != nil {
		log.Fatalln(err)
	}

	serv := thanos.New(&context.ServerContext{
		RequirePass: "",
		Store:       store,
	})

	if err := serv.ListenAndServe(":6380"); err != nil {
		log.Fatalln(err)
	}
}
