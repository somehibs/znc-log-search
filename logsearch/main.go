package main

import (
	"fmt"
	"github.com/somehibs/znc-log-search"
)

func main() {
	// Global config
	conf := logs.GetConf()

	// Configure path searcher
	err := logs.ZncPath(conf.Network)
	if err != nil {
		fmt.Println(err)
		return
	}
}
