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
	sphinx *SphinxFeed
	id *IdFeed
	indexes map[string]ChanIndex
}

type Logfile struct {
	Path string
	User string
	Channel string
	Size int64
	Time time.Time
	StartIndex int64
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

func (fc *FileCollector) InitDb(sphinx *SphinxFeed, id *IdFeed) {
	fc.sphinx = sphinx
	fc.id = id
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (fc *FileCollector) GetLogsBackwards() error {
	today := StartOfDay(time.Now())
	end := time.Date(2015, 9, 23, 0, 0, 0, 0, time.UTC)
	fc.now = today
	//fc.now = time.Date(2015, 12, 23, 0, 0, 0, 0, time.UTC)
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

func (fc *FileCollector) DailyLogsForever(file chan Line, id chan IdLine) error {
	for {
		fmt.Println("dailylogsforever")
		fc.GetLogsForDay(fc.Out, StartOfDay(time.Now()))
		fmt.Println("gotten")
		time.Sleep(2*time.Second)
		// Wait for all the queues to drain.
		// Sleep for a little bit
		for {
			if len(file) > 0 || len(id) > 0 || len(fc.Out) > 0 {
				time.Sleep(2*time.Second)
			} else {
				fmt.Println("dozing")
				break
			}
		}
		fmt.Println("zzzz")
		time.Sleep(90*time.Second)
	}
}

func (fc *FileCollector) GetLogsForDay(reply chan Logfile, day time.Time) error {
	return fc.GetLogsForChan(reply, day, "*")
}

func (fc *FileCollector) LogfilePath(match string, day *time.Time) *Logfile {
	return fc.LogfilePathExist(match, day, nil)
}

func (fc *FileCollector) LogfilePathExist(match string, day *time.Time, exist *Logfile) *Logfile {
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
	knownOffset := int64(0)
	uid := fmt.Sprintf("%s%s", user, channel)
	index := fc.indexes[uid]
	if index.Channel != "" {
		knownOffset = index.Index+1
		fmt.Printf("known offset %s on %+v\n", channel, index)
	} else {
		//fmt.Printf("Couldn't find index %s on %+v\n", uid, index)
		//for _, v := range fc.indexes {
		//	fmt.Printf("%+v\n", v)
		//}
		//fmt.Printf("Ok: %d\n", len(fc.indexes))
		//panic("")
	}
	//file, e := os.Open(match)
	fileSize := int64(0)
	stat, e := os.Stat(match)
	if e == nil {
		fileSize = stat.Size()
	}
	if exist == nil {
		lf := Logfile{match, user, channel, fileSize, *day, knownOffset}
		return &lf
	} else {
		exist.Path = match
		exist.User = user
		exist.Channel = channel
		exist.Size = fileSize
		exist.Time = *day
		exist.StartIndex = knownOffset
		return exist
	}
}

func (fc *FileCollector) GetLogsForChan(reply chan Logfile, day time.Time, oneChan string) error {
	checkPath()
	chanData := fc.sphinx.GetMaxChanIndexes(&day)
	chanData = fc.id.GetChannels(chanData)
	fc.indexes = ToMap(chanData)
	path := fmt.Sprintf(zncPath, oneChan, day.String()[:10])
	match, e := filepath.Glob(path)
	if e != nil {
		return e
	}
	fc.MergePaths(reply, match, &day)
	return nil
}

func (fc *FileCollector) MergePaths(reply chan Logfile, match []string, day *time.Time) {
	subset := make([]Logfile, len(match))
	sizes := make(map[string]*Logfile, len(match))
	appended := 0
	//fmt.Printf("Parsing: %d\n", len(match))
	for _, m := range match {
		lp := subset[appended]
		l := fc.LogfilePathExist(m, day, &lp)
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
		fmt.Printf("Dispatching %s\n", l)
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
