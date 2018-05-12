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
	e := ""
	lines := 0
	for {
		lines += 1
		fmt.Printf("pending\n")
		line := <-c
		fmt.Printf("Last: %+v\n", line)
		panic("")
		time.Sleep(3 * time.Second)
		if lines%1000 == 0 {
			fmt.Printf("Lines: %d Last: %+v\n", lines, line)
		}
	}
	if e != "" {
		fmt.Printf("Error: %s\n", e)
	}
}
