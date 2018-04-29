package main

import (
	"fmt"
	"github.com/somehibs/znc-log-search"
)

func main() {
	conf := logs.GetConf()
	fmt.Println(logs.UserNetworkZncPath(conf.Network))
}
