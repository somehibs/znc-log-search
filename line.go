package logs

import (
	"errors"
	"io"
	"strconv"
	"strings"
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
	Index int64
	Nick string
	Message string
	Channel string
	User string
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

var akaIndex = 2;

func (p *LineParser) InitChan() {
	p.Out = make(chan Line, GetConf().Queues["line"])
}

func (p *LineParser) ParseLinesForever() {
	f := <-p.In
	for ;f.Channel != ""; {
		p.ParseLinesForFile(f)
		f = <-p.In
	}
}

func (p *LineParser) ParseLinesForFile(file Logfile) {
	//fmt.Printf("hi: %s\n", file)
	// Open the file.
	f, e := os.Open(file.Path)
	if e != nil {
			fmt.Printf("Failed to open file %s (Err: %s)\n", file.Path, e)
	}
	// Seek line by line, starting from lastLine.
	_, e = f.Seek(file.StartIndex, 0)
	//bail := file.Size - file.StartIndex
	//if file.StartIndex > 0 {
	//	fmt.Printf("%s %s bail: %d tried: %d new ind: %d e: %s\n", file.Time, file.Channel, bail, file.StartIndex, a, e)
	//}
	rdr := bufio.NewReader(f)
	index := int64(file.StartIndex)
	lc := int64(0)
	lineGuess := int64(file.Size / 48)
	if lineGuess < 1 {
		lineGuess = 1
	}
	buffer := make([]Line, lineGuess)
	for line := ""; e == nil; line, e = rdr.ReadString('\n') {
		if line == "" {
			continue
		}
		result := &buffer[0]
		if lc >= lineGuess {
			result = &Line{}
		} else {
			result = &buffer[lc]
		}
		e := p.ParseLine(&file, &line, index, result)
		//fmt.Printf("line: %+v\nlen: %s\n", result, len(line))
		if e == nil {
			p.Out <- *result
			lc += int64(1)
		}
		index += int64(len(line))
	}
	if lc > 0 {
		fmt.Printf("Sent %d lines for %s\n", lc, file.Channel)
	}
	if e != nil && e != io.EOF {
		fmt.Println("Error: " + e.Error())
		panic("Unexpected error")
	}
}

func combineTime(t *time.Time, ts string) time.Time {
	h, e := strconv.ParseInt(ts[0:2], 10, 32)
	if e != nil {
		panic(fmt.Sprintf("Panicing : %s", e))
	}
	m, e := strconv.ParseInt(ts[3:5], 10, 32)
	if e != nil {
		panic(fmt.Sprintf("Panicing combining : %s", e))
	}
	s, e := strconv.ParseInt(ts[6:8], 10, 32)
	if e != nil {
		panic(fmt.Sprintf("Panicing combining time: %s", e))
	}
	return time.Date(t.Year(), t.Month(), t.Day(), int(h), int(m), int(s), 0, time.UTC)
}

func (p *LineParser) ParseLine(file *Logfile, line *string, index int64, l *Line) (e error) {
	l.Index = index
	l.Channel = file.Channel
	l.User = file.User
	for _, r := range skiplist {
			m := r.FindAllString(*line, -1)
			if len(m) > 0 {
				e = errors.New("skip")
				return
			}
	}
	for ri, r := range re {
		// test regexp against line
		match := r.FindStringSubmatch(*line)
		if len(match) > 0 {
			sub := r.SubexpNames()
			for i, name := range sub {
				if i == 0 { continue }
				switch name {
					case "nick":
						l.Nick = strings.ToLower(match[i])
					case "time":
						l.Time = combineTime(&file.Time, match[i])
					case "msg":
						l.Message = match[i]
						if ri == akaIndex && strings.Contains(l.Message, "is now known as") {
							return errors.New("aka ignore")
						}
					default:
						panic("Don't understand " + name)
				}
			}
			return
		}
	}
	return errors.New("no match")
//	panic(fmt.Sprintf("NO MATCH %s %d", *line, len(*line)))
}
