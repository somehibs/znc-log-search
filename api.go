package logs

import (
	//"fmt"
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

// Return a state object with public fields you want to serve over JSON
type ApiState interface {
	ApiState() interface{}
}

func InitApi(stateHandler http.Handler) *Api {
	a := Api{}
	a.Init(&stateHandler)
	return &a
}

func (a *Api) Init(stateHandler *http.Handler) {
	if GetConf().Prometheus != false {
		// Register prometheus' handler.
		http.Handle("/metrics", promhttp.Handler())
	}
	// Register an HTTP handler that also provides an API to query the state of various internal queues
	http.Handle("/state", *stateHandler)
	go http.ListenAndServe("127.0.0.1:9991", nil)
}
