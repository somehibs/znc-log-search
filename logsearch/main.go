package main

import (
	//"fmt"

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
	parser.ParseLinesForever()
	//if e != nil {
	//	fmt.Printf("Error: %s\n", e)
	//}
}
