package main

import (
	"fmt"
	"time"

	"github.com/somehibs/znc-log-search"
)

var feed *logs.SphinxFeed

func main() {
	//Collect(true)
	//c := make(chan int,0)
	//<-c
	Collect(false)
}

func Collect(today bool) {
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
	if feed == nil {
		feed = &sphinx
		sphinx.Connect()
	} else {
		sphinx = *feed
	}
	id.Connect()

	collector.InitDb(&sphinx, &id)

	// Dispatch logs forever to the output channel
	if today {
		go collector.DailyLogsForever(parser.Out, id.Out)
	} else {
		go collector.GetLogsBackwards()
	}
	go parser.ParseLinesForever()
	go id.QueryIdsForever()
	//go ExhaustChan(id.Out)
	go sphinx.InsertSphinxForever()

	if today == false {
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
		time.Sleep(3*time.Second)
		if lines % 1000 == 0 {
			fmt.Printf("Lines: %d Last: %+v\n", lines, line)
		}
	}
	if e != "" {
		fmt.Printf("Error: %s\n", e)
	}
}
