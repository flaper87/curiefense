package pkg

import (
	"github.com/curiefense/curiefense/curielogger/pkg/entities"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"io"
)

type LogSender struct {
	exportDrivers io.WriteCloser
	metrics       *Metrics

	closed *atomic.Bool
}

func NewLogSender(v *viper.Viper, drivers io.WriteCloser, metrics *Metrics) *LogSender {
	return &LogSender{exportDrivers: drivers, metrics: metrics, closed: atomic.NewBool(false)}
}

func (ls *LogSender) Write(l *entities.LogEntry) error {
	ls.metrics.add(l)
	b, _ := jsoniter.Marshal(l.CfLog)
	_, err := ls.exportDrivers.Write(b)
	return err
}

func (ls *LogSender) Close() error {
	ls.closed.Store(true)
	return ls.exportDrivers.Close()
}

func (ls *LogSender) Closed() bool {
	return ls.closed.Load()
}
