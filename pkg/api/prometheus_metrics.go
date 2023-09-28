package api

import (
	"net/http"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ApiCalledCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lunettes_api_called_counter",
			Help: "",
		},
		[]string{"apiname", "useragent", "ip"},
	)
)

func init() {
	prometheus.MustRegister(ApiCalledCounter)
}

func debugApiCalledCounter(handleName string, request *http.Request) {
	useragent := request.UserAgent()
	ipAddr := utils.ClientIP(request)
	ApiCalledCounter.WithLabelValues(handleName, useragent, ipAddr).Inc()
}
