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
	cache []map[string]int64
	c arango.Client
}

func (f *IdFeed) QueryIdsForever() {
	f.cache = make([]map[string]int64, 2)
	for ;; {
		e := f.Connect()
		if e != nil {
			panic(fmt.Sprintf("%s", e.Error()))
		}
		for ;; {
			QueryId(<-f.In)
		}
	}
}

func QueryId(l Line) {
	//
}

func (f *IdFeed) Connect() (e error) {
	f.Out = make(chan IdLine, GetConf().Queues["id"])
	conf := GetConf().Arango
	httpConnection, e := http.NewConnection(http.ConnectionConfig{Endpoints: conf.Endpoints})
	if e != nil {
		return e
	}
	c, e := arango.NewClient(arango.ClientConfig{Connection: httpConnection, Authentication: arango.BasicAuthentication(conf.User, conf.Password)})
	if e != nil {
		return e
	}
	fmt.Printf("Got it %s\n", c)
	return
}
