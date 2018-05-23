package logs

import (
	"fmt"
	"os"
	"os/exec"
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
	u, e := user.Current()
	if e != nil {
		return e
	}
	// Default behaviour pulls znc logs from the current user
	prefix := fmt.Sprintf("/home/%s", u.Username)
	if GetConf().LogDir != "" {
		prefix = GetConf().LogDir
	}
	userRoot := fmt.Sprintf("%s/.znc/users/", prefix)
	cmd := exec.Command("chmod", "-R", userRoot, "077")
	cmd.Run()
	zncPath = fmt.Sprintf("%s*/networks/%s/moddata/log/%%s/%%s.log", userRoot, GetConf().Network)
	if zncPath != "" {
		return nil
	}
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
	//fc.now = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
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
		time.Sleep(60 * time.Second)
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
		//Debug(ftag, fmt.Sprintf("Hi, I found an offset for channel: %s on day %s with index %+v\n", channel, day, index))
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
	fc.indexes = ToMap(chanData)
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
			//Debug(ftag, fmt.Sprintf("No existing size for: %+v\n", l))
			sizes[lp.Channel] = l
		}
		if l.StartIndex > sizes[lp.Channel].StartIndex {
			//Debug(ftag, fmt.Sprintf("My start index %s is superior %s\n", l, sizes[lp.Channel]))
			sizes[lp.Channel] = l
		}
		if sizes[lp.Channel].Size < l.Size && sizes[lp.Channel].StartIndex == 0 {
			Debug(ftag, fmt.Sprintf("My size %s is bigger than %s\n", l, sizes[lp.Channel]))
			sizes[lp.Channel] = l
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
