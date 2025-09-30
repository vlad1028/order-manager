package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var IssuedOrdersLabel = "issued_orders_total"

var (
	IssuedOrdersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "issued_orders_total",
			Help: "Total number of issued orders",
		},
		[]string{IssuedOrdersLabel},
	)
)

func AddIssuedOrdersTotal(cnt int, label string) {
	IssuedOrdersTotal.With(prometheus.Labels{
		IssuedOrdersLabel: label,
	}).Add(float64(cnt))
}

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("failed to start metrics server: %v", err)
		}
	}()
}
