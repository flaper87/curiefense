package drivers

import (
	"github.com/spf13/viper"
	"io"
)

func InitDrivers(v *viper.Viper) io.WriteCloser {
	output := make([]io.WriteCloser, 0)
	if v.GetBool(`STDOUT_ENABLED`) {
		output = append(output, NewBufferedStdout())
	}
	if v.GetBool(`GCS_ENABLED`) {
		if g := NewGCS(v); g != nil {
			output = append(output, NewGCS(v))
		}
	}
	return NewTee(output)
}
