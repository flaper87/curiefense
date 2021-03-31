package drivers

import (
	"bytes"
	"github.com/spf13/viper"
	"net/http"
)

const (
	CURIELOGGER_OUTPUTS_LOGSTASH_URL = `CURIELOGGER_OUTPUTS_LOGSTASH_URL`
)

type Logstash struct {
	url string
}

func NewLogstash(v *viper.Viper) *FluentD {
	return &FluentD{url: v.GetString(CURIELOGGER_OUTPUTS_LOGSTASH_URL)}
}

func (l *Logstash) Write(p []byte) (n int, err error) {
	r, err := http.Post(l.url, "application/json", bytes.NewReader(p))
	if r != nil && r.Body != nil {
		r.Body.Close()
	}
	return len(p), nil
}

func (l *Logstash) Close() error {
	return nil
}
