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
	Db *sql.DB
	BufferedLines int64
	InsertedLines int64
	index int64
	value []string
	valueData []interface{}
}

func (f *SphinxFeed) InsertSphinxForever() {
	f.index = f.GetMaxId()+1
	for {
		f.BufferOne(<-f.In)
		if (GetConf().Daily && len(f.In) > 0) || len(f.value) < 1000 {
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

type ChanIndex struct {
	Index int64

	Channel string
	User string

	ChannelId int64
	UserId int64
}

func ToMap(ind []ChanIndex) map[string]ChanIndex {
	ret := map[string]ChanIndex{}
	for _, v := range ind {
		ret[fmt.Sprintf("%s%s", v.User, v.Channel)] = v
	}
	return ret
}

func (f *SphinxFeed) GetMaxId() (r int64) {
	cur, e := f.Db.Query("SELECT MAX(id) FROM irc_msg")
	if e != nil {
		panic(fmt.Sprintf("%s", e))
	}
	if cur.Next() {
		var max int64
		e = cur.Scan(&max)
		if e != nil {
			panic(fmt.Sprintf("%s", e))
		}
		r = max
	}
	return
}

func (f *SphinxFeed) GetMaxChanIndexes(day *time.Time) []ChanIndex {
	// Clamp the day to the end of the day, add 24 hours and take a second off
	max := day.Add(24*time.Hour)
	max = max.Add(-1*time.Second)
	query := fmt.Sprintf("SELECT MAX(line_index) as li, channel_id, user_id FROM irc_msg WHERE timestamp >= %d AND timestamp <= %d GROUP BY channel_id LIMIT 9999 option max_matches=%d", day.Unix(), max.Unix(), 100000)
	//fmt.Printf("%s\n", query)
	cur, e := f.Db.Query(query)
	if e != nil {
		panic(fmt.Sprintf("Could not query chan indexes %s", e))
	}
	//fmt.Printf("Max: %s Min: %s\n", max, day)
	m := make([]ChanIndex, 0)
	for ;cur.Next(); {
		var channel, user, index int64
		e = cur.Scan(&index, &channel, &user)
	//	mapKey := fmt.Sprintf("%d%d", channel, user)
	//	if chanMap[mapKey] > index {
	//		// Supersceded by another channel entry
	//		continue
	//	}
		m = append(m, ChanIndex{Index: index, UserId: user, ChannelId: channel})
		if e != nil {
			panic(fmt.Sprintf("Could not scan row %s", e))
		}
	}
	return m
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
	f.InsertedLines += f.BufferedLines
	f.BufferedLines = 0
	//fmt.Printf("Cursor: %s", cur)
	f.value = make([]string, 0)
	f.valueData = make([]interface{}, 0)
}

func (f *SphinxFeed) BufferOne(l IdLine) {
	// Buffer this line into the query string.
	f.BufferedLines += 1
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
	db.SetMaxOpenConns(26)
	f.Db = db
	return e
}
