package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/distributedio/titan/command"
)

var (
	key       string
	token     string
	namespace string
)

func run(key, token, namespace string) (string, error) {
	if token != "" {
		ns, err := command.Verify([]byte(token), []byte(key))
		if err != nil {
			return "", fmt.Errorf("auth failed :%s", err)
		}
		return fmt.Sprintf("auth sucess\nNamespace: %s", string(ns)), nil
	} else {
		token, err := command.Token([]byte(key), []byte(namespace), time.Now().Unix())
		if err != nil {
			return "", fmt.Errorf("create token failed %s", err)
		}
		return fmt.Sprintf("token : %s", token), nil
	}
}

func main() {
	flag.StringVar(&key, "key", "", "server key")
	flag.StringVar(&token, "token", "", "client token")
	flag.StringVar(&namespace, "namespace", "", "biz name")
	flag.Parse()

	out, err := run(key, token, namespace)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(out)
}
