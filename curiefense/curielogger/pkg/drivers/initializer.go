package drivers

import (
	"github.com/spf13/viper"
	"io"
	"os"
)

const (
	STDOUT_ENABLED  = `STDOUT_ENABLED`
	GCS_ENABLED     = `GCS_ENABLED`
	FLUENTD_ENABLED = `CURIELOGGER_USES_FLUENTD`
)

func InitDrivers(v *viper.Viper) io.WriteCloser {
	output := make([]io.WriteCloser, 0)
	if v.GetBool(STDOUT_ENABLED) {
		output = append(output, os.Stdout)
	}
	if v.GetBool(GCS_ENABLED) {
		if g := NewGCS(v); g != nil {
			output = append(output, g)
		}
	}

	// DEPRECATED
	if v.GetBool(FLUENTD_ENABLED) {
		output = append(output, NewFluentD(v))
	}

	return NewTee(output)
}
