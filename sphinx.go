package logs

import (
	"fmt"
	"strings"

	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

type SphinxFeed struct {
	In chan IdLine
	queue []string
	c *sql.DB
	value []string
	valueData []interface{}
}

func (f *SphinxFeed) InsertSphinxForever() {
	for {
		e := f.Connect()
		if e != nil {
			panic(fmt.Sprintf("Could not connect to Sphinx: %s", e.Error))
		}
		for {
			f.QueueOne(<-f.In)
			if len(f.In) > 0 && len(f.queue) < 500 {
				continue
			}
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
	cur, e := f.c.Query(query, f.valueData)
	if e != nil {
		panic(fmt.Sprintf("Query failed %s", e))
	}
	fmt.Printf("Cursor: %s", cur)
}

func (f *SphinxFeed) QueueOne(l IdLine) {
	// Buffer this line into the query string.
	f.value = append(f.value, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	f.valueData = append(f.valueData, 1)
	f.valueData = append(f.valueData, l.Line.Time)
	f.valueData = append(f.valueData, l.Line.Nick)
	f.valueData = append(f.valueData, l.Line.Channel)
	f.valueData = append(f.valueData, l.ChannelId)
	f.valueData = append(f.valueData, l.Line.Message)
	f.valueData = append(f.valueData, l.Line.Index)
	f.valueData = append(f.valueData, l.NickId)
	f.valueData = append(f.valueData, permissionFor(l.Line.Channel))
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
	return 9999
}

func (f *SphinxFeed) Connect() error {
	db, e := sql.Open("mysql", GetConf().Sphinx.Dsn)
	f.c = db
	return e
}
