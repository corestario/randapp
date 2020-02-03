package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MsgMetrics struct {
	NumMsgs *prometheus.CounterVec
}

const (
	PrometheusLabelStatus   = "status"
	PrometheusLabelMsgType  = "msg_type"
	PrometheusValueReceived = "Received"
	PrometheusValueAccepted = "Accepted"
)

func NewPrometheusMsgMetrics(module string) *MsgMetrics {
	numMsgs := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "Randapp",
		Subsystem: module + "_MetricsSubsystem",
		Name:      "NumMsgs",
		Help:      "number of messages since start",
	},
		[]string{PrometheusLabelStatus, PrometheusLabelMsgType},
	)
	prometheus.MustRegister(numMsgs)
	return &MsgMetrics{
		NumMsgs: numMsgs,
	}
}
