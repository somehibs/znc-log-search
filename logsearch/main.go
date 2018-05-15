package main

import (
	"fmt"
	"time"

	"github.com/somehibs/znc-log-search/logs"
)

var feed *logs.SphinxFeed

func main() {
	m := logs.NewManager()
	if logs.GetConf().Indexer.Daily {
		m.Daily()
	} else {
		m.Historical()
	}
	m.WaitUntilCompletion()
}

func ExhaustChan(c chan logs.IdLine) {
	lines := 0
	for {
		lines += 1
		logs.Debug("main", "Waiting for line\n")
		line := <-c
		logs.Debug("main", fmt.Sprintf("Last: %+v\n", line))
		time.Sleep(3 * time.Second)
		if lines%1000 == 0 {
			logs.Debug("main", fmt.Sprintf("Lines: %d Last: %+v\n", lines, line))
		}
	}
}
