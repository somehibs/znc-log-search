package logs

import (
	"fmt"
	"strings"

	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

type SphinxFeed struct {
	In chan IdLine
	index int
	c *sql.DB
	value []string
	valueData []interface{}
}

func (f *SphinxFeed) InsertSphinxForever() {
	f.index = 1
	for {
		e := f.Connect()
		if e != nil {
			panic(fmt.Sprintf("Could not connect to Sphinx: %s", e.Error))
		}
		for {
			f.QueueOne(<-f.In)
			if len(f.In) > 0 && len(f.value) < 500 {
				continue
			}
			fmt.Println("Inserting %s users", len(f.value))
			f.Insert()
		}
	}
}

func (f *SphinxFeed) Insert() {
	// Insert a prebuilt query string
	if len(f.value) == 0 {
		fmt.Println("Ignoring attempt to insert no data")
		return
	}
	query := fmt.Sprintf("INSERT INTO irc_msg (id, timestamp, nick, channel, channel_id, msg, line_index, nick_id, permission) VALUES %s", strings.Join(f.value, ","))
	//fmt.Printf("Query: %s", query)
	cur, e := f.c.Query(query, f.valueData...)
	if e != nil {
		panic(fmt.Sprintf("Query failed %s", e))
	}
	defer cur.Close()
	//fmt.Printf("Cursor: %s", cur)
	f.value = make([]string, 0)
	f.valueData = make([]interface{}, 0)
}

func (f *SphinxFeed) QueueOne(l IdLine) {
	// Buffer this line into the query string.
	f.value = append(f.value, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	f.valueData = append(f.valueData, f.index)
	f.valueData = append(f.valueData, l.Line.Time.Unix())
	f.valueData = append(f.valueData, l.Line.Nick)
	f.valueData = append(f.valueData, l.Line.Channel)
	f.valueData = append(f.valueData, l.ChannelId)
	f.valueData = append(f.valueData, l.Line.Message)
	f.valueData = append(f.valueData, l.Line.Index)
	f.valueData = append(f.valueData, l.NickId)
	f.valueData = append(f.valueData, permissionFor(l.Line.Channel))
	f.index += 1
}

func permissionFor(channel string) int {
	// Based on the permission matrix, insert a permission.
	// If there isn't a permission, insert it as max permission
	for k, v := range GetConf().Permissions {
		for _, c := range v {
			if (c == channel) {
				return k
			}
		}
	}
	return 3
}

func (f *SphinxFeed) Connect() error {
	db, e := sql.Open("mysql", GetConf().Sphinx.Dsn)
	db.SetMaxOpenConns(30)
	f.c = db
	return e
}
