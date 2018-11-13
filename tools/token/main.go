package main

import (
	"flag"
	"fmt"
	"time"

	"gitlab.meitu.com/platform/titan/command"
)

var (
	key       string
	token     string
	namespace string
)

func main() {
	flag.StringVar(&key, "key", "", "server key")
	flag.StringVar(&token, "token", "", "client token")
	flag.StringVar(&namespace, "namespace", "", "biz name")
	flag.Parse()
	if token != "" {
		ns, err := command.Verify([]byte(token), []byte(key))
		if err != nil {
			fmt.Printf("auth failed :%s\n", err)
			return
		}
		fmt.Println("auth sucess")
		fmt.Println("Namespace:", string(ns))
	} else {
		token, err := command.Token([]byte(key), []byte(namespace), time.Now().Unix())
		if err != nil {
			fmt.Printf("create token failed %s\n", err)
			return
		}
		fmt.Printf("token : %s\n", token)
	}
}
