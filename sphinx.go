package logs

import (
	"fmt"

	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

type SphinxFeed struct {
	In chan IdLine
	queue []string
	c *sql.DB
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
	// Build a query string based on Sphinx entries
}

func (f *SphinxFeed) QueueOne(l IdLine) {
	// Instead of inserting one, we should buffer
}

func (f *SphinxFeed) Connect() error {
	db, e := sql.Open("mysql", "")//GetConf().Sphinx.Dsn)
	f.c = db
	return e
}
