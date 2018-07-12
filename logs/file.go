package logs

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type FileCollector struct {
	now      time.Time
	Out      chan Logfile
	Done     chan int
	LastTime *time.Time
	Asleep	 bool
	sphinx   *SphinxFeed
	id       *IdFeed
	indexes map[string]ChanIndex
}

type Logfile struct {
	Path       string
	User       string
	Channel    string
	Size       int64
	Time       time.Time
	StartIndex int64
}

var zncPath = ""
var ftag = "FILE"

func checkPath() error {
	if zncPath != "" {
		return nil
	}
	u, e := user.Current()
	if e != nil {
		return e
	}
	prefix := fmt.Sprintf("/home/%s", u.Username)
	if GetConf().LogDir != "" {
		prefix = GetConf().LogDir
	}
	zncPath = fmt.Sprintf("%s/.znc/users/*/networks/%s/moddata/log/%%s/%%s.log", prefix, GetConf().Network)
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
	for {
		if fc.now.Before(end) { //|| fc.now == today {
			fc.Out <- Logfile{}
			fc.Done <- 0
			return nil
		}
		e := fc.GetLogsForDay(fc.Out, fc.now)
		if e != nil {
			panic(e)
		}
		fc.now = fc.now.Add(-time.Hour * 24)
	}
}

func (fc *FileCollector) DailyLogsForever(file chan Line, id chan IdLine) error {
	for {
		fc.Asleep = false
		fc.GetLogsForDay(fc.Out, StartOfDay(time.Now()))
		fc.Asleep = true
		time.Sleep(2 * time.Second)
		for {
			if len(file) > 0 || len(id) > 0 || len(fc.Out) > 0 {
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
		time.Sleep(90 * time.Second)
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
		knownOffset = index.Index + 1
		//fmt.Printf("known offset %s on %+v\n", channel, index)
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
	fc.LastTime = &day
	chanData := fc.sphinx.GetMaxChanIndexes(&day)
	chanData = fc.id.GetChannels(chanData)
	//Debug(ftag, fmt.Sprintf("Found channels: %s", chanData))
	fc.indexes = ToMap(chanData)
	//Debug(ftag, fmt.Sprintf("Found channel map: %s", fc.indexes))
	Debug(ftag, fmt.Sprintf("Chan index size: %d", len(fc.indexes)))
	path := fmt.Sprintf(zncPath, oneChan, day.String()[:10])
	match, e := filepath.Glob(path)
	Debug(ftag, fmt.Sprintf("%+v", path))
	Debug(ftag, fmt.Sprintf("Files: %d", len(match)))
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
		if sizes[lp.Channel] == nil {
			Debug(ftag, fmt.Sprintf("New channel: %s", l))
			sizes[lp.Channel] = l
		}
		if sizes[lp.Channel].Size < l.Size && l.StartIndex >= sizes[lp.Channel].StartIndex {
			Debug(ftag, fmt.Sprintf("Overriding channel: %s with %s", sizes[lp.Channel], l))
			sizes[lp.Channel] = l
		} else {
			Debug(ftag, fmt.Sprintf("Channel %s lost to %s", l, sizes[lp.Channel]))
		}
		appended += 1
	}
	//fmt.Printf("Queuing: %d for day %s\n", len(sizes), day)
	for _, l := range sizes {
		//fmt.Printf("Dispatching %s\n", l)
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
		wl := GetConf().Indexer.Whitelist
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
