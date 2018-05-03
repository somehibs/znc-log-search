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

	// Create a line parser
	parser := logs.LineParser{In: collector.Out}
	parser.InitChan()

	// Create an ID parser
	id := logs.IdFeed{In: parser.Out}
	id.InitChan()

	// Sphinx feed
	sphinx := logs.SphinxFeed{In: id.Out}
	sphinx.Connect()
	id.Connect()

	collector.InitDb(&sphinx, &id)

	// Dispatch logs forever to the output channel
	go collector.DailyLogsForever(parser.Out, id.Out)
	go parser.ParseLinesForever()
	ExhaustChan(parser.Out)
	go id.QueryIdsForever()
	//go sphinx.InsertSphinxForever()

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
}

func ExhaustChan(c chan logs.Line) {
	e := ""
	lines := 0
	for {
		lines += 1
		line := <-c
		fmt.Printf("Last: %+v\n", line)
		panic("")
		time.Sleep(3*time.Second)
		if lines % 1000 == 0 {
			fmt.Printf("Lines: %d Last: %+v\n", lines, line)
		}
	}
	if e != "" {
		fmt.Printf("Error: %s\n", e)
	}
}
