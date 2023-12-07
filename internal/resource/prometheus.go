package resource

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPReqTotal           *prometheus.CounterVec
	ConcurrentRequestCount prometheus.GaugeFunc

	PushMsgTotal *prometheus.CounterVec
)

const (
	PromLabelPath   = "path"
	PromLabelStatus = "status"

	PromLabelSVC = "service"
	PromLabelDC  = "dc"
)

func initPromtheus() {
	HTTPReqTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests made.",
	}, []string{PromLabelPath, PromLabelStatus})

	PushMsgTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "push_msg_total",
		Help: "Total number of push msg",
	}, []string{PromLabelSVC, PromLabelDC})

	ConcurrentRequestCount = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "concurrent_request_count",
		Help: "the number of concurrent request",
	}, func() float64 {
		return float64(ConcurrentReq)
	})

	prometheus.MustRegister(
		HTTPReqTotal,
		PushMsgTotal,
		ConcurrentRequestCount,
	)
}
