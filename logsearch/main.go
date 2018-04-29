package main

import (
	//"fmt"

	"github.com/somehibs/znc-log-search"
)

func main() {
	collector := logs.FileCollector{}
	collector.GetLogsForever()
	//if e != nil {
	//	fmt.Printf("Error: %s\n", e)
	//}
}
