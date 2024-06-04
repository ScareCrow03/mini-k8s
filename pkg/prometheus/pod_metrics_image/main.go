package main

import (
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ( // 以下3个指标每1s更新一次
	// 简单自增型指标
	myMetric = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "aaa_my_metric",
			Help: "This is my inc metric",
		},
	)
	// 随机数指标
	dynamicMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "aaa_dynamic_metric",
			Help: "This is my dynamic metric",
		},
	)
	// 正弦曲线指标
	sineMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "aaa_sine_metric",
			Help: "This is my sine metric",
		},
	)
)

func init() {
	prometheus.MustRegister(myMetric)
	prometheus.MustRegister(dynamicMetric)
	prometheus.MustRegister(sineMetric)
}

func updateDynamicMetric() {
	for {
		value := rand.Float64() * 100
		dynamicMetric.Set(value)
		time.Sleep(1 * time.Second)
	}
}

func updateSineMetric() {
	for {
		value := math.Sin(float64(time.Now().Unix()))
		sineMetric.Set(value)
		time.Sleep(1 * time.Second)
	}
}

func updateMyMetric() {
	for {
		myMetric.Inc()
		time.Sleep(1 * time.Second)
	}
}

func main() {
	go updateDynamicMetric()
	go updateSineMetric()
	go updateMyMetric()
	// 暴露2112端口
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
