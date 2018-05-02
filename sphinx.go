package logs

import (
	"fmt"
	"strings"
	"time"

	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

type SphinxFeed struct {
	In chan IdLine
	index int
	Db *sql.DB
	value []string
	valueData []interface{}
}

func (f *SphinxFeed) InsertSphinxForever() {
	f.index = 1
	for {
		f.QueueOne(<-f.In)
		if len(f.In) > 0 && len(f.value) < 500 {
			continue
		}
		f.Insert()
	}
}

//func (f *SphinxFeed) GetOldest(channel string) {
//	c, e := f.Db.Query(getOldest, nil)
//	if e != nil {
//		panic("Couldn't find oldest item in sphinx")
//	}
//}

func (f *SphinxFeed) GetMaxChanIndexes(day *time.Time) {
	// Clamp the day to the end of the day, add 24 hours and take a second off
	day.Add(24*time.Hour)
}

func (f *SphinxFeed) Insert() {
	// Insert a prebuilt query string
	if len(f.value) == 0 {
		fmt.Println("Ignoring attempt to insert no data")
		return
	}
	query := fmt.Sprintf("INSERT INTO irc_msg (id, timestamp, nick, channel, channel_id, msg, line_index, nick_id, permission, user_id) VALUES %s", strings.Join(f.value, ","))
	//fmt.Printf("Query: %s", query)
	cur, e := f.Db.Query(query, f.valueData...)
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
	f.value = append(f.value, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	f.valueData = append(f.valueData, f.index)
	f.valueData = append(f.valueData, l.Line.Time.Unix())
	f.valueData = append(f.valueData, l.Line.Nick)
	f.valueData = append(f.valueData, l.Line.Channel)
	f.valueData = append(f.valueData, l.ChannelId)
	f.valueData = append(f.valueData, l.Line.Message)
	f.valueData = append(f.valueData, l.Line.Index)
	f.valueData = append(f.valueData, l.NickId)
	f.valueData = append(f.valueData, permissionFor(l.Line.Channel))
	f.valueData = append(f.valueData, l.UserId)
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
	return GetConf().DefaultPermission
}

func (f *SphinxFeed) Connect() error {
	db, e := sql.Open("mysql", GetConf().Sphinx.Dsn)
	db.SetMaxOpenConns(29)
	f.Db = db
	return e
}
