package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	myMetric = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "aaa_my_metric",
			Help: "This is my metric",
		},
	)
)

func init() {
	prometheus.MustRegister(myMetric)
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
