package main

import (
	"flag"
	"fmt"

	"github.com/m-asama/nff-go-upf/config"
	"github.com/m-asama/nff-go-upf/upf"
)

func main() {
	var err error
	f := flag.String("config", "config.yaml", "config file path")
	flag.Parse()
	conf, err := config.Parse(*f)
	if err != nil {
		fmt.Println("config parse error:", err)
		return
	}
	err = upf.Init(conf)
	if err != nil {
		fmt.Println("upf init error:", err)
		return
	}
	upf.Debug()
	upf.Run()
}
