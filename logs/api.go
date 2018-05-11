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

// New creates an http listener from a map of URLs to handlers.
// Depending on configuration, it may also register a prometheus handler.
func NewApi(handlers map[string]http.Handler) *Api {
	a := Api{}

	for k, v := range handlers {
		http.Handle("/"+k, v)
	}
	// unwritten debug mode
	//http.Handle("/debug", a.gopsAgent())
	if GetConf().Prometheus != false {
		// Register prometheus' handler.
		http.Handle("/metrics", promhttp.Handler())
	}
	return &a
}

// Listen calls ListenAndServe on http with either configuration or an override url
func (a *Api) Listen() {
	if GetConf().ApiUrl != "" {
		go http.ListenAndServe(GetConf().ApiUrl, nil)
	}
}
