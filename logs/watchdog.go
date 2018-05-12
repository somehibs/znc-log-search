package logs

import (
	"time"
)

type Watchdog struct {
	interval time.Duration
	timer    *time.Timer
}

func New(interval time.Duration, callback func()) *Watchdog {
	w := Watchdog{
		interval,
		time.AfterFunc(interval, callback),
	}
	return &w
}

func (w *Watchdog) Stop() {
	w.timer.Stop()
}

func (w *Watchdog) Kick() {
	w.Stop()
	w.timer.Reset(w.interval)
}
