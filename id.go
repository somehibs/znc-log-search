package logs

import (
	"fmt"
	"strconv"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

var channels = "Channels"
var nicks = "Nicks"

type IdLine struct {
	Line *Line
	NickId int64
	ChannelId int64
}

type IdFeed struct {
	In chan Line
	Out chan IdLine
	nicks map[string]int64
	chans map[string]int64
	nickLen int64
	chanLen int64
	c arango.Client
	db arango.Database
}

func (f *IdFeed) InitChan() {
	f.Out = make(chan IdLine, GetConf().Queues["id"])
	f.chans = map[string]int64{}
	f.nicks = map[string]int64{}
}

func (f *IdFeed) InitLens() {
	f.chanLen = f.GetLen(channels)
	f.nickLen = f.GetLen(nicks)
}

func (f *IdFeed) GetLen(collection string) (length int64) {
	return
}

func (f *IdFeed) QueryIdsForever() {
	for ;; {
		e := f.Connect()
		f.InitLens()
		if e != nil {
			panic(fmt.Sprintf("%s", e.Error()))
		}
		for ;; {
			f.Out <- f.QueryId(<-f.In)
		}
	}
}

func (f *IdFeed) QueryId(l Line) (id IdLine) {
	id.Line = &l
	id.NickId = f.nicks[l.Nick]
	if id.NickId == 0 {
		id.NickId = f.Get(nicks, l.Nick)
	}
	id.ChannelId = f.chans[l.Channel]
	if id.ChannelId == 0 {
		id.ChannelId = f.Get(channels, l.Channel)
	}
	return
}

type ArangoItem struct {
	Value string
}

func (f *IdFeed) Get(collection, value string) int64 {
	item := ArangoItem{}
	cur, e := f.db.Query(nil, "FOR i IN "+collection+" FILTER i.value == @val RETURN i", map[string]interface{}{"val": value})
	if e != nil {
		panic(e.Error())
	}
	m, e := cur.ReadDocument(nil, &item)
	fmt.Printf("k: %s m: %+v\n", value, item)
	if e == nil || arango.IsNoMoreDocuments(e) {
		ret, e := strconv.ParseInt(m.Key, 10, 64)
		if e != nil {
			panic(e.Error())
		}
		cache := &f.nicks
		if collection == "Channels" {
			cache = &f.chans
		}
		(*cache)[value] = ret
		return ret
	}
	panic(e.Error())
}

func (f *IdFeed) Connect() (e error) {
	conf := GetConf().Arango
	//fmt.Printf("C: %+v\n", GetConf())
	httpConnection, e := http.NewConnection(http.ConnectionConfig{Endpoints: conf.Endpoints})
	if e != nil {
		return e
	}
	f.c, e = arango.NewClient(arango.ClientConfig{Connection: httpConnection, Authentication: arango.BasicAuthentication(conf.User, conf.Password)})
	if e != nil {
		return e
	}
	f.db, e = f.c.Database(nil, conf.Db)
	return
}
