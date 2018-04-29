package logs

import (
	"fmt"
	"time"
	"os/user"
	//"path/filepath"
)

type FileCollector struct {
		// collect all the files in given paths. allow format strings
		path string
		// 2d array, first dimension is for multiple sprintf, second is sprintf args
		args [][]string
}

var zncPath = ""

func CustomZncPath(path string) error {
	zncPath = path
	return nil
}

func ZncPath(network string) error {
	u, e := user.Current()
	if e != nil {
		return e
	}
	zncPath = fmt.Sprintf("/home/%s/.znc/users/*/networks/%s/moddata/log/%%s/", u.Username, network)
	return nil
}

func GetLogsForDay(reply chan string, day time.Time) chan string {
	return GetLogsForChan(reply, day, "")
}

func GetLogsForChan(reply chan string, day time.Time, oneChan string) chan string {
	if oneChan == "" {
		oneChan = "*"
	}
	//path := fmt.Sprintf(zncPath, oneChan)
	return make(chan string)
}
