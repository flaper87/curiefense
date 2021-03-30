package pkg

import (
	"github.com/curiefense/curiefense/curielogger/pkg/entities"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

const (
	PROMETHEUS_EXPORT_PORT    = `CURIELOGGER_PROMETHEUS_LISTEN`
	PROMETHEUS_EXPORT_ENABLED = `CURIELOGGER_METRICS_PROMETHEUS_ENABLED`
)

type Metrics struct {
}

func NewMetrics(v *viper.Viper) *Metrics {

	// set up prometheus server
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(v.GetString(`PROMETHEUS_EXPORT_PORT`), nil)
	log.Printf("Prometheus exporter listening on %v", v.GetString(`PROMETHEUS_EXPORT_PORT`))
	return &Metrics{}
}

func (m *Metrics) add(l entities.LogEntry) {

}
