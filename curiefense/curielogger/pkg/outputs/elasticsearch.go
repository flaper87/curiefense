package outputs

import (
	"github.com/elastic/go-elasticsearch/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

const (
	ELASTICSEARCH_URL = `ELASTICSEARCH_URL`
)

type ElasticSearch struct {
	client *elasticsearch.Client
}

type ElasticsearchConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	Url                string `mapstructure:"url"`
	KibanaUrl          string `mapstructure:"kibana_url"`
	Initialize         bool   `mapstructure:"initialize"`
	Overwrite          bool   `mapstructure:"overwrite"`
	AccessLogIndexName string `mapstructure:"accesslog_index_name"`
	UseDataStream      bool   `mapstructure:"use_data_stream"`
	ILMPolicy          string `mapstructure:"ilm_policy"`
}

func NewElasticSearch(v *viper.Viper, cfg ElasticsearchConfig) *ElasticSearch {
	url := v.GetString(ELASTICSEARCH_URL)
	c, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: strings.Split(url, `,`),
	})
	if err != nil {
		log.Error(err)
		return nil
	}
	log.Info(`initialized es`)
	return &ElasticSearch{client: c}
}

func (es *ElasticSearch) Write(p []byte) (n int, err error) {
	r, err := es.client.Index(
		"curieaccesslog",
		strings.NewReader(string(p)),
		es.client.Index.WithRefresh("true"),
		es.client.Index.WithPretty(),
		es.client.Index.WithFilterPath("result", "_id"),
	)
	if err != nil {
		return 0, err
	}
	return len(p), r.Body.Close()
}

func (es *ElasticSearch) Close() error {
	return es.Close()
}
