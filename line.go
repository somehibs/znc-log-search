package logs

import (
	"io"
	"time"
	"bufio"
	"fmt"
	"os"
)

type LineParser struct {
	In chan Logfile
	Out chan Line
	// lastLine exists to prevent reindexing the same file seek index
	lastLine map[string]map[string]string
}

type Line struct {
	// Dunno much about this.
	Time time.Time
	Nick string
	Message string
	Channel string
}

func (p *LineParser) ParseLinesForever() {
	f := <-p.In
	for ;f.Channel != ""; {
		p.ParseLinesForFile(<-p.In)
	}
}

func (p *LineParser) ParseLinesForFile(file Logfile) {
	// Open the file.
	f, e := os.Open(file.Path)
	if e != nil {
			fmt.Printf("Failed to open file (Err: %s)\n", e)
	}
	// Seek line by line, starting from lastLine.
	rdr := bufio.NewReader(f)
	index := int64(0)
	for line := ""; e == nil; line, e = rdr.ReadString('\n') {
		p.Out <- p.ParseLine(&file, &line, index)
		index += int64(len(line))
	}
	if e != nil && e != io.EOF {
		fmt.Println("Error: " + e.Error())
	}
}

func (p *LineParser) ParseLine(file *Logfile, line *string, index int64) Line {
	fmt.Printf("Found line %d with ind %d\n", len(*line), index)
	return Line{}
}
