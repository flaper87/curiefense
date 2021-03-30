package pkg

import (
	"github.com/curiefense/curiefense/curielogger/pkg/entities"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"io"
	"log"
)

type LogSender struct {
	exportDrivers []io.WriteCloser
	metrics       *Metrics
}

var (
	jsonEncoder = jsoniter.ConfigFastest
)

func NewLogSender(v *viper.Viper, drivers []io.WriteCloser, metrics *Metrics) *LogSender {
	return &LogSender{exportDrivers: drivers, metrics: metrics}
}

func (ls *LogSender) Write(l entities.LogEntry) {
	// add metrics
	ls.metrics.add(l)
	// send logs
	b, err := jsonEncoder.Marshal(l)
	if err != nil {
		log.Print(err)
	}
	for _, logger := range ls.exportDrivers {
		if _, err = logger.Write(b); err != nil {
			log.Print(err)
		}
	}
}

func (ls *LogSender) Close() error {
	// close drivers
	return nil
}
