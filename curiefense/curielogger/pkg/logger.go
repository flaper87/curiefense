package pkg

import (
	"github.com/curiefense/curiefense/curielogger/pkg/entities"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"io"
)

type LogSender struct {
	exportDrivers io.WriteCloser
	encoder       *jsoniter.Encoder
	metrics       *Metrics
}

func NewLogSender(v *viper.Viper, drivers io.WriteCloser, metrics *Metrics) *LogSender {
	return &LogSender{exportDrivers: drivers, encoder: jsoniter.ConfigFastest.NewEncoder(drivers), metrics: metrics}
}

func (ls *LogSender) Write(l *entities.LogEntry) error {
	// add metrics
	ls.metrics.add(l)
	// send logs
	return ls.encoder.Encode(l)
}

func (ls *LogSender) Close() error {
	// close drivers
	return nil
}
