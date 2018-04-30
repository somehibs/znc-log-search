package logs

import (
	"fmt"
)

type SphinxFeed struct {
	In chan IdLine
	queue []string
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
	return nil
}
