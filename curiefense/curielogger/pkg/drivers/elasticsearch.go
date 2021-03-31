package drivers

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/spf13/viper"
	"log"
	"strings"
)

const (
	ELASTICSEARCH_URL = `ELASTICSEARCH_URL`
)

type ElasticSearch struct {
	client *elasticsearch.Client
}

func NewElasticSearch(v *viper.Viper) *ElasticSearch {
	c, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: strings.Split(v.GetString(ELASTICSEARCH_URL), `,`),
	})
	if err != nil {
		log.Print(err)
		return nil
	}
	log.Print(`initialized es`)
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
