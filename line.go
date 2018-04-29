package logs

import (
	"errors"
	"io"
	"time"
	"bufio"
	"fmt"
	"os"
	"regexp"
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

// Common regex fields
var IRCTIME = `[0-9]{2}:[0-9]{2}:[0-9]{2}`
var IRCNICK = `[\.a-zA-Z0-9_\-\\\[\]\{\}\^\'\|\x60~]+`
var GREEDY = `.*`

// Skip joins and parts
var skiplist []*regexp.Regexp = []*regexp.Regexp {
	regexp.MustCompile(fmt.Sprintf(`\[(?P<time>%s)\] \*\*\* \w+: (?P<nick>%s) (?P<msg>%s)`, IRCTIME, IRCNICK, GREEDY)),
}

// Pick up messages, mode changes, nick changes and notices
var re []*regexp.Regexp = []*regexp.Regexp {
	regexp.MustCompile(fmt.Sprintf(`\[(?P<time>%s)\] -(?P<nick>%s)- (?P<msg>%s)`, IRCTIME, IRCNICK, GREEDY)),
	regexp.MustCompile(fmt.Sprintf(`\[(?P<time>%s)\] \* (?P<nick>%s) (?P<msg>%s)`, IRCTIME, IRCNICK, GREEDY)),
	regexp.MustCompile(fmt.Sprintf(`\[(?P<time>%s)\] \*\*\* (?P<nick>%s) (?P<msg>%s)`, IRCTIME, IRCNICK, GREEDY)),
	regexp.MustCompile(fmt.Sprintf(`\[(?P<time>%s)\] <(?P<nick>%s)> (?P<msg>%s)`, IRCTIME, IRCNICK, GREEDY)),
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
		if line == "" {
			continue
		}
		result, e := p.ParseLine(&file, &line, index)
		if e == nil {
			//p.Out <- result
			fmt.Sprintf("%s", result)
		}
		index += int64(len(line))
	}
	if e != nil && e != io.EOF {
		fmt.Println("Error: " + e.Error())
		panic("Unexpected error")
	}
}

func (p *LineParser) ParseLine(file *Logfile, line *string, index int64) (l Line, e error) {
	for _, r := range skiplist {
			m := r.FindAllString(*line, -1)
			if len(m) > 0 {
				e = errors.New("skip")
				return
			}
	}
	for _, r := range re {
			// test regexp against line
			match := r.FindStringSubmatch(*line)
			if len(match) > 0 {
				sub := r.SubexpNames()
				for i, name := range sub {
					if i == 0 { continue }
					switch name {
						case "nick":
							l.Nick = sub[i]
						case "time":
//							l.Time = sub[i]
						case "msg":
							l.Message = sub[i]
						default:
							panic("Don't understand " + name)
					}
				}
				return
			}
	}
	panic("NO MATCH")
}
