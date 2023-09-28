package slo

import "github.com/prometheus/client_golang/prometheus"

var (
	sloOngoingSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lunettes_slo_ongoing_size",
			Help: "slo map size",
		},
		// queue name
		[]string{"name"},
	)
)

func init() {
	prometheus.MustRegister(sloOngoingSize)
}
