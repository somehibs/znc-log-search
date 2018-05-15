package logs

import (
	"fmt"
)

func Debug(tag, log string) {
	if (GetConf().Debug) {
		fmt.Printf("D (%s): %s\n", tag, log)
	}
}

func Log(tag, log string) {
	fmt.Println(tag + ": " + log)
}
