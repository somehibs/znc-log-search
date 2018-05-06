package logs

import (
	"fmt"
	"net/http"

	//"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Api struct {
	prometheus *Prometheus
}

type Prometheus struct {
	gauges map[string]int64
}

func InitApi(stateHandler http.Handler) *Api {
	a := Api{}
	a.Init(&stateHandler)
	return &a
}

func (a *Api) Init(stateHandler *http.Handler) {
	if GetConf().Prometheus != false {
		fmt.Println("Adding prometheus support")
		// Register a bunch of metrics and pepper the ability to use them throughout the system
		http.Handle("/metrics", promhttp.Handler())
	}
	// Register an HTTP handler that also provides an API to query the state of various internal queues via some sort of queue callback
	http.Handle("/state", *stateHandler)
}
