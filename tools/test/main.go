package main

import (
	"flag"
	"testing"

	"github.com/meitu/titan/tools/autotest"
)

func main() {
	t := &testing.T{}
	var testcase string
	var addr string

	flag.StringVar(&addr, "addr", ":7369", "titan server addr")
	flag.StringVar(&testcase, "testcase", "", "default run testcase all")
	flag.Parse()
	client := autotest.NewAutoClient()
	client.Start(addr)
	switch testcase {
	case "string":
		client.StringCase(t)
	case "list":
		client.ListCase(t)
	case "key":
		client.KeyCase(t)
	default:
		client.StringCase(t)
		client.ListCase(t)
		client.KeyCase(t)
	}
	client.Close()
}
