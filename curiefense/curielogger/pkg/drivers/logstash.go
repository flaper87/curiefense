package drivers

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

const (
	CURIELOGGER_OUTPUTS_LOGSTASH_URL = `CURIELOGGER_OUTPUTS_LOGSTASH_URL`
)

type Logstash struct {
	url string
}

func NewLogstash(v *viper.Viper) *Logstash {
	log.Info(`initialized logstash`)
	return &Logstash{url: v.GetString(CURIELOGGER_OUTPUTS_LOGSTASH_URL)}
}

func (l *Logstash) Write(p []byte) (n int, err error) {
	r, err := http.Post(l.url, "application/json", bytes.NewReader(p))
	if err != nil {
		return 0, err
	}
	return len(p), r.Body.Close()
}

func (l *Logstash) Close() error {
	return nil
}
