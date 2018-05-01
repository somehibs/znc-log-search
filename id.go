package logs

import (
	"fmt"
	"strconv"
	"encoding/json"
	"strings"

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

type ArangoLen struct {
	Value int64
	Len bool
}

func (f *IdFeed) GetLen(collection string) (length int64) {
	cur, e := f.db.Query(nil, "FOR x IN " + collection + " FILTER x._key == \"0\" RETURN x", nil)
	if e != nil {
		panic(fmt.Sprintf("error querying len %s because %s", collection, e))
	}
	doc := ArangoLen{}
	_, e = cur.ReadDocument(nil, &doc)
	if e != nil || doc.Value == 0 {
		fmt.Printf("\n\nNo len for collection %s!\n\n", collection)
		m, e := f.UpsertNow(collection, map[string]string{"_key":"0"}, map[string]string{"_key":"0", "Value":"1"}, map[string]string{})
		if e != nil {
			panic(fmt.Sprintf("db failed to create len %s", e))
		}
		fmt.Printf("Metadata %+v\n", m)
		doc.Value = 1
	}
	length = doc.Value
	return
}

func (f *IdFeed) QueryIdsForever() {
	for ;; {
		e := f.Connect()
		if e != nil {
			panic(fmt.Sprintf("%s", e.Error()))
		}
		f.InitLens()
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

func qu(i map[string]string) string {
	s, e := json.Marshal(i)
	if e != nil {
		panic(fmt.Sprintf("Could not json input: %s (E: %s)", i, e))
	}
	return string(s)
}

func Upsert(collection string, filter map[string]string, insert map[string]string, update map[string]string) (u string) {
	u = "UPSERT " + qu(filter)
	u += " INSERT " + qu(insert)
	u += " UPDATE " + qu(update)
	u += " IN " + collection + " RETURN {NEW}"
	return
}

func (f *IdFeed) UpsertNow(collection string, query, insert, update map[string]string) (r interface{}, e error) {
	q := Upsert(collection, query, insert, update)
	cur, e := f.db.Query(nil, q, nil)
	if e != nil {
		fmt.Printf("Error upserting: %s\n", e)
		panic("")
	}
	defer cur.Close()
	var item map[string]map[string]string
	_, e = cur.ReadDocument(nil, &item)
	if e == nil || arango.IsNoMoreDocuments(e) {
		r = item
		return
	}
	return
}

func (f *IdFeed) Get(collection, value string) int64 {
	length := &f.chanLen
	if (collection == nicks) {
		length = &f.nickLen
	} else if (collection != channels) {
		panic("Don't know this collection")
	}
	if *length == 0 {
		panic("bad configuration")
	}

	query := Upsert(collection,
									map[string]string{"value": value},
									map[string]string{"_key": fmt.Sprintf("%d", *length), "value": value},
									map[string]string{})
	cur, e := f.db.Query(nil, query, nil)
	if e != nil {
		fmt.Printf("query: %s\n", query)
		if strings.Contains(e.Error(), "unique constraint violated") {
			*length += 1
			return f.Get(collection, value)
		}
		panic(e.Error())
	}
	defer cur.Close()

	var item map[string]map[string]string
	_, e = cur.ReadDocument(nil, &item)
	if e == nil || arango.IsNoMoreDocuments(e) {
		key := item["NEW"]["_key"]
		ret, e := strconv.ParseInt(key, 10, 64)
		if e != nil {
			panic(e.Error())
		}
		if ret == *length {
			// Key changed!
			*length = (*length)+1
			f.SaveLen(collection, *length)
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

func (f *IdFeed) SaveLen(collection string, newLength int64) {
	// Forcibly update the index variable
	lenItem := ArangoLen{newLength, true}
	c, e := f.db.Collection(nil, collection)
	//fmt.Printf("C: %+v E: %+v\n", c, e)
	m, e := c.UpdateDocument(nil, "0", &lenItem)
	if e != nil {
		panic(fmt.Sprintf("M: %+v E: %+v\n", m, e))
	}
}

func (f *IdFeed) Connect() (e error) {
	conf := GetConf().Arango
	fmt.Printf("C: %+v\n", GetConf())
	httpConnection, e := http.NewConnection(http.ConnectionConfig{Endpoints: conf.Endpoints})
	if e != nil {
		return e
	}
	f.c, e = arango.NewClient(arango.ClientConfig{Connection: httpConnection, Authentication: arango.BasicAuthentication(conf.User, conf.Password)})
	if e != nil {
		return e
	}
	f.db, e = f.c.Database(nil, conf.Db)
	if e != nil {
		return e
	}
	return
}
