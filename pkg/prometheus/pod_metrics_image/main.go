package main

import (
	"math/rand"
	"net/http"
	"time"

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
	dynamicMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "aaa_dynamic_metric",
			Help: "This is my dynamic metric",
		},
	)
)

func init() {
	prometheus.MustRegister(myMetric)
	prometheus.MustRegister(dynamicMetric)
}

func updateDynamicMetric() {
	for {
		value := rand.Float64() * 100
		dynamicMetric.Set(value)
		time.Sleep(5 * time.Second)
	}
}

func main() {
	go updateDynamicMetric()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
