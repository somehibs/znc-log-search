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

func UserNetworkZncPath(network string) error {
	u, e := user.Current()
	if e != nil {
		return e
	}
	zncPath = fmt.Sprintf("/home/%s/.znc/users/*/networks/%s/moddata/log/", u.Username, network)
	fmt.Println(zncPath)
	return nil
}

func GetLogsForDay(day time.Time) chan string {
	return make(chan string)
}
