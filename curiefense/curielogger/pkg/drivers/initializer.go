package drivers

import (
	"github.com/spf13/viper"
	"io"
)

func InitDrivers(v *viper.Viper) []io.WriteCloser {
	output := make([]io.WriteCloser, 0)
	if v.GetBool(`STDOUT_ENABLED`) {
		output = append(output, NewBufferedStdout())
	}
	return output
}
