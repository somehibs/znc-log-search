package main

import (
	"fmt"

	"github.com/somehibs/znc-log-search"
)

func main() {
	// Create a channel file collector
	collector := logs.FileCollector{}
	collector.InitChan()
	// Dispatch logs forever to the output channel
	go collector.GetLogsForever()

	// Create a line parser
	parser := logs.LineParser{In: collector.Out, Out: make(chan logs.Line, 10000)}
	go parser.ParseLinesForever()

	// Create an ID parser
	id := logs.IdFeed{In: parser.Out}
	id.InitChan()
	go id.QueryIdsForever()

	lines := 0
	for {
		lines += 1
		line := <-id.Out
		if lines % 1000 == 0 {
			fmt.Printf("Lines: %d Last: %+v\n", lines, line)
		}
	}
	//if e != nil {
	//	fmt.Printf("Error: %s\n", e)
	//}
}
