package drivers

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	neturl "net/url"
)

const (
	CURIELOGGER_FLUENTD_URL = `CURIELOGGER_FLUENTD_URL`
)

type FluentD struct {
	url string
}

func NewFluentD(v *viper.Viper) *FluentD {
	return &FluentD{url: fmt.Sprintf("%scuriefense.log", v.GetString(CURIELOGGER_FLUENTD_URL))}
}

func (b *FluentD) Write(p []byte) (n int, err error) {
	r, err := http.PostForm(b.url, neturl.Values{"json": {string(p)}})
	if r != nil && r.Body != nil {
		r.Body.Close()
	}
	return len(p), err
}

func (b *FluentD) Close() error {
	return nil
}
