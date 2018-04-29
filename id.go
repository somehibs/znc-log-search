package logs

import (
	"fmt"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

var channels = "Channels"
var nicks = "Nicks"

type IdLine struct {
	
}

type IdFeed struct {
	In chan Line
	Out chan IdLine
}

func (f *IdFeed) QueryIdsForever() {
	for ;; {
		f.Out = make(chan IdLine, GetConf().Queues["id"])
		conf := GetConf().Arango
		httpConnection, e := http.NewConnection(http.ConnectionConfig{Endpoints: conf.Endpoints})
		if e != nil {
			panic("Cannot connect to ArangoDb")
		}
		driver, e := arango.NewClient(arango.ClientConfig{Connection: httpConnection, Authentication: arango.BasicAuthentication(conf.User, conf.Password)})
		if e != nil {
			panic("Cannot open Arango connection")
		}
		fmt.Printf("Got driver %s", driver)
		panic("")
	}
}
