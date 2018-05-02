package logs

import (
	"fmt"
	"time"
	"os/user"
	"os"
	"strings"
	"path/filepath"
)

type FileCollector struct {
	now time.Time
	Out chan Logfile
	Done chan int
	sphinx SphinxFeed
	id IdFeed
}

type Logfile struct {
	Path string
	User string
	Channel string
	Size int64
	Time time.Time
	StartOffset int
}

var zncPath = ""
func checkPath() error {
	if zncPath != "" {
		return nil
	}
	u, e := user.Current()
	if e != nil {
		return e
	}
	zncPath = fmt.Sprintf("/home/%s/.znc/users/*/networks/%s/moddata/log/%%s/%%s.log", u.Username, GetConf().Network)
	return nil
}

func (fc *FileCollector) InitChan() {
	fc.Out = make(chan Logfile, GetConf().Queues["file"])
	fc.Done = make(chan int)
}

func (fc *FileCollector) InitDb(sphinx SphinxFeed, id IdFeed) {
	fc.sphinx = sphinx
	fc.id = id
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (fc *FileCollector) GetLogsForever() error {
	// TODO: find oldest file in sphinx
	today := StartOfDay(time.Now())
	end := time.Date(2015, 9, 23, 0, 0, 0, 0, time.UTC)
	fc.now = today
	for ;; {
		if fc.now.Before(end) {//|| fc.now == today {
			fc.Out <- Logfile{}
			fc.Done <- 0
			return nil
		}
		fc.GetLogsForDay(fc.Out, fc.now)
		fc.now = fc.now.Add(-time.Hour*24)
	}
	return nil
}

func (fc *FileCollector) GetLogsForDay(reply chan Logfile, day time.Time) error {
	return fc.GetLogsForChan(reply, day, "*")
}

func LogfilePath(match string, day *time.Time) *Logfile {
	return LogfilePathExist(match, day, nil)
}

func LogfilePathExist(match string, day *time.Time, exist *Logfile) *Logfile {
	exploded := strings.Split(match, "/")
	user := ""
	channel := ""
	for i := range exploded {
		if i > 0 {
			switch exploded[i-1] {
			case "users":
				user = exploded[i]
			case "log":
				channel = exploded[i]
			}
		}
	}
	//file, e := os.Open(match)
	fileSize := int64(0)
	stat, e := os.Stat(match)
	if e == nil {
		fileSize = stat.Size()
	}
	if exist == nil {
		lf := Logfile{match, user, channel, fileSize, *day}
		return &lf
	} else {
		exist.Path = match
		exist.User = user
		exist.Channel = channel
		exist.Size = fileSize
		exist.Time = *day
		return exist
	}
}

func (fc *FileCollector) GetLogsForChan(reply chan Logfile, day time.Time, oneChan string) error {
	checkPath()
	chanData := fc.sphinx.GetMaxChanIndexes(day)
	path := fmt.Sprintf(zncPath, oneChan, day.String()[:10])
	match, e := filepath.Glob(path)
	if e != nil {
		return e
	}
	MergePaths(reply, match, &day)
	return nil
}

func MergePaths(reply chan Logfile, match []string, day *time.Time) {
	subset := make([]Logfile, len(match))
	sizes := make(map[string]*Logfile, len(match))
	appended := 0
	//fmt.Printf("Parsing: %d\n", len(match))
	for _, m := range match {
		lp := subset[appended]
		l := LogfilePathExist(m, day, &lp)
		if Whitelist(l.Channel) == false {
			continue
		}
		if sizes[lp.Channel] == nil || sizes[lp.Channel].Size < l.Size {
			sizes[lp.Channel] = l
		}
		appended += 1
	}
	fmt.Printf("Queuing: %d for day %s\n", len(sizes), day)
	for _, l := range sizes {
		reply <- *l
	}
}

var whitelist = map[string]bool{}
var notified = map[string]bool{}

func Whitelist(channel string) bool {
	if whitelist == nil {
		return true
	}
	if len(whitelist) == 0 {
		wl := GetConf().Whitelist
		if len(wl) == 0 {
			whitelist = nil
			return true
		}
		for _, k := range wl {
			whitelist[k] = true
		}
	}
	ok := whitelist[channel]
	if ok == false && channel[0] == '#' && notified[channel] == false {
		notified[channel] = true
		fmt.Println("Ignoring " + channel)
	}
	return ok
}
