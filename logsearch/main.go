package main

import (
	"fmt"
	"time"

	"github.com/somehibs/znc-log-search"
)

func main() {
	// Create a channel file collector
	collector := logs.FileCollector{}
	collector.InitChan()
	// Dispatch logs forever to the output channel
	go collector.GetLogsForever()

	// Create a line parser
	parser := logs.LineParser{In: collector.Out}
	parser.InitChan()
	go parser.ParseLinesForever()

	// Create an ID parser
	id := logs.IdFeed{In: parser.Out}
	id.InitChan()
	go id.QueryIdsForever()

	sphinx := logs.SphinxFeed{In: id.Out}
	go sphinx.InsertSphinxForever()

	<-collector.Done
	fmt.Println("Collector finished queueing files.")
	for {
		if len(parser.Out) > 0 ||
			len(id.Out) > 0 {
			fmt.Println("Queues not empty. Waiting for queues to empty...")
			time.Sleep(2*time.Second)
		} else {
			fmt.Println("Queues are complete. Finishing.")
			return
		}
	}

	//lines := 0
	//for {
	//	lines += 1
	//	line := <-id.Out
	//	if lines % 1000 == 0 {
	//		fmt.Printf("Lines: %d Last: %+v\n", lines, line)
	//	}
	//}
	//if e != nil {
	//	fmt.Printf("Error: %s\n", e)
	//}
}
