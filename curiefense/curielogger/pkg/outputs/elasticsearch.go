package outputs

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	ELASTICSEARCH_URL   = `ELASTICSEARCH_URL`
	ACCESSLOG_ES_PREFIX = "curieaccesslog"
)

type EmbeddedResource struct {
	Filename string
	Contents string
	Length   int
}

var ES_RESOURCES = map[string]EmbeddedResource{

	"files/elasticsearch/es_index_template.json": {
		Filename: "files/elasticsearch/es_index_template.json",
		Contents: "H4sIAAAAAAAC/9xYT2/rNgy/51MIvgwYioe3HXvasF0GDLsMOw2DwUh0rFUWNYrOq1/g7z7ISfOncdw6VoFhp9ay+Av//Uiau5VShfUGn8sAIsg+Fo/qz2K3+/Sj1hjjr7T5Jb3+DRrs+2+Lvx6SRGBLbKUrHtX3nz8PR5qaQBFNSVVC2N/bIkdLvnhU36Xn3c5W6tMfEX8Ggd+FEZq+LwwIlHF4Kh7Vrn/Y7dCbvk8AZYMC6XSllFKFwajZBtljFj+1bLFCH/GbqAYrlGATHAiqYqVUPyjxcnSCiShi/SYeT16ccHagVOFshbrTDi+OlSo8NOnslpeKh/PLTM7RFrkEZyFOiR2l+tX5334PVzQQwmutTeehsbp8sTEefT+8DkwBWSzGS8uglfolgBeWSRcGy56w+0JszlQ62VSsHemnkhHiEIZRAFr/jVom5NHcEl0TOQQ/Lmvoiz/lyrn4DVuHd5q8R53SRpAb6+GQQZfXpnR4pcdeF8uohbEhQTCGMcYpTBtmwwVimYT0ghvkaVxHGlwG/c5xsigWmIQ0OWQmXhyNbHHIEIHV2P/naZzKWqp7c5LYoIPO+sSrK/LMd1cF1qEZYlojOKl1jfopA2zrJJuS1m/BWYN+Sx3jPy1GqREMclwMPVjOGFEWQ3nae7Brw2hpugeRqRWsqPXLncgg6GxjcwTkiBWRt1ZjJuomYw9N9SbSVUsaQ4rQBId83u1HoCpHIG8ADYG0xqHYBqldniWtP3Tdr2gMClgX55v7nsJSk7mrrQeQ+j7B6xp5uz6eCx4oPacIbpCmvHYYOx5ev5+APFaDGzPBBf4GqQxk/fVP7OvVxhNj2YCriJuBb8ItTt39Wm7BtXi4eXWxX009v84x4E3boJc4nf1p0vf4BpVAhO26FcwBtibTrbs3sN41NGiiJ5tFp3d0knlIeQwkthvrwY3QMUeBOHDO3lcjGGMgH3EOZXPG3mAWkLvr7/wUmoOTx0vCabTjj+gu/Jy64RX0ZPgry1FeJqNk4cLO7CBKJpglWr3pq6hrvJpo3scygU28S/CHFJ4o0IRb4gYESw+e4o2fdrOCq22okWNrJcPshjFa8tYsR/K2pihZJsrTCis3meReMp12IBl48BH0zKrghxP1xlfbdN67Ngry4uw6LabS93jLuPgz42NWXf/ZHRJtkStHX/5X66NRxDy7isP4l+ujllG42y8XnjWiybBiEAY/+O3AiNE98931d3SzPmzm9zRPL/pV36/+DQAA//8BaAX9GhkAAA==",
		Length:   6426,
	},

	"files/elasticsearch/ilm_policy.json": {
		Filename: "files/elasticsearch/ilm_policy.json",
		Contents: "H4sIAAAAAAAC/3SPz8oCIRTF9z7F4a6/xcziI5hlmx4jRC+N4MyNUaI/+O5hYWlNLsSfP/Ucbwqgo3hnLjQgU+ZRBw4vBmiUWCFA2kQnc2g2AVrEeznxQmgFQJM+7/WBaQD1lv6+ZXDXh/3vdts1b8XkvL57jupAUp+rVO6TZc+R2/KTm0uVTV3l16/eb6wkqTInldQ9AAD//wL4qZJPAQAA",
		Length:   335,
	},

	"files/elasticsearch/index_settings.json": {
		Filename: "files/elasticsearch/index_settings.json",
		Contents: "H4sIAAAAAAAC/6rmUlBQSszJTCxOLVayUgBxFRSUqqv1HJOTU4uLffLTPfNSUiv8EnNTa2vhChQUlDKL48uLMktS4zNB8kpWCiVFpalgyVouEK7l4gIEAAD//1j8ho1cAAAA",
		Length:   92,
	},

	"files/kibana/dashboard.ndjson": {
		Filename: "files/kibana/dashboard.ndjson",
		Contents: "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",
		Length:   34326,
	},
}

//
// Return the contents of a resource.
//
func getResource(path string) ([]byte, error) {
	if entry, ok := ES_RESOURCES[path]; ok {
		var raw bytes.Buffer
		var err error

		// Decode the data.
		in, err := base64.StdEncoding.DecodeString(entry.Contents)
		if err != nil {
			return nil, err
		}

		// Gunzip the data to the client
		gr, err := gzip.NewReader(bytes.NewBuffer(in))
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		data, err := ioutil.ReadAll(gr)
		if err != nil {
			return nil, err
		}
		_, err = raw.Write(data)
		if err != nil {
			return nil, err
		}

		// Return it.
		return raw.Bytes(), nil
	}
	return nil, fmt.Errorf("failed to find resource '%s'", path)
}

//
// Return the available resources in a slice.
//
func getResources() []EmbeddedResource {
	i := 0
	ret := make([]EmbeddedResource, len(ES_RESOURCES))
	for _, v := range ES_RESOURCES {
		ret[i] = v
		i++
	}
	return ret
}

type ElasticSearch struct {
	client *elasticsearch.Client
	cfg    ElasticsearchConfig
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
	es := &ElasticSearch{client: c, cfg: cfg}
	es.ConfigureEs()
	return es
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
	return nil
}

func (es *ElasticSearch) ConfigureKibana() {

	var body, ktpl bytes.Buffer
	var fw io.Writer
	var err error
	res, _ := getResource("files/kibana/dashboard.ndjson")
	gTpl := template.Must(template.New("it").Parse(string(res)))
	gTpl.Execute(&ktpl, es.cfg)

	mwriter := multipart.NewWriter(&body)
	if fw, err = mwriter.CreateFormFile("file", "dashboard.ndjson"); err != nil {
		log.Error("Error creating writer: %v", err)
		return
	}
	if _, err := io.Copy(fw, bytes.NewReader(ktpl.Bytes())); err != nil {
		log.Error("Error with io.Copy: %v", err)
		return
	}
	mwriter.Close()

	log.Debug("configuring kibana")
	kbUrl := fmt.Sprintf("%s/api/saved_objects/_import?overwrite=true", es.cfg.KibanaUrl)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	bReader := bytes.NewReader(body.Bytes())
	req, err := http.NewRequest("POST", kbUrl, bReader)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", mwriter.FormDataContentType())
	req.Header.Set("kbn-xsrf", "true")

	for i := 0; i < 60; i++ {
		rst, err := client.Do(req)
		if rst != nil {
			if rst.StatusCode == 200 {
				log.Debugf("kibana dashboard imported %s", kbUrl)
				break
			}

			if rst.StatusCode == 409 {
				log.Debugf("kibana index pattern already exists %s", kbUrl)
				break
			}
		}

		log.Errorf("kibana index pattern creation failed (retrying in 5s) %s %v %v %v", kbUrl, err, req, rst)
		time.Sleep(5 * time.Second)
		bReader.Seek(0, 0)
	}
}

func (es *ElasticSearch) ConfigureEs() {
	var res *esapi.Response
	var err error
	for i := 0; i < 60; i++ {
		res, err = es.client.ILM.GetLifecycle()
		if err != nil {
			log.Errorf("There was an error while querying the ILM Policies %v", err)
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	var ilm map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&ilm); err != nil {
		log.Errorf("There was an error while reading the ILM Policies %v", err)
		return
	}

	_, exists := ilm[es.cfg.AccessLogIndexName]
	if es.cfg.Overwrite || !exists {
		log.Debugf("creating / overwriting elasticsearch ilm policy %s for %s\n", es.cfg.AccessLogIndexName, es.cfg.Url)

		policy := es.cfg.ILMPolicy
		if policy == "" {
			var iTpl bytes.Buffer
			res, _ := getResource("files/elasticsearch/ilm_policy.json")
			gTpl := template.Must(template.New("it").Parse(string(res)))
			gTpl.Execute(&iTpl, es.cfg)
			policy = string(iTpl.Bytes())
		}

		body := es.client.ILM.PutLifecycle.WithBody(strings.NewReader(policy))
		resp, err := es.client.ILM.PutLifecycle(es.cfg.AccessLogIndexName, body)
		if err != nil || resp.IsError() {
			log.Printf("[ERROR] index template creation failed %v %v", err, resp)
		}
	}

	// Create the Index Template
	//
	// This is how the mapping, ILM policies, and rollover aliases are assigned to
	// the indices or datastreams. There should always be an index template.
	//
	// TODO: Version the index template, as we may have to change the index mapping
	// in the future. Elastic's beats handle this in a decent way, look them up before
	// working on this task.
	tplExists, err := es.client.Indices.ExistsIndexTemplate(es.cfg.AccessLogIndexName)

	if err != nil {
		log.Error("there was an error while querying the template %v", err)
		return
	}

	if es.cfg.Overwrite || tplExists.IsError() {
		log.Printf("[DEBUG] creating / overwriting elasticsearch index template %s for %s\n", ACCESSLOG_ES_PREFIX, es.cfg.Url)
		var iTpl bytes.Buffer
		res, _ := getResource("files/elasticsearch/es_index_template.json")
		gTpl := template.Must(template.New("it").Parse(string(res)))
		gTpl.Execute(&iTpl, es.cfg)

		resp, err := es.client.Indices.PutIndexTemplate(es.cfg.AccessLogIndexName, bytes.NewReader(iTpl.Bytes()))
		if err != nil || resp.IsError() {
			log.Errorf("[ERROR] index template creation failed %v %v", err, resp)
		}
	}

	// Data streams take care of creating the initila index, assigning an ILM policy
	// to it, and all the internal management. The index will be `hidden` and prefixed
	// with `.ds` so, in kibana, it is necessary to flag the "show hidden indeces" option.
	//
	// For non data stream configs, we have to create the initial index to make sure the
	// alias is assigned, the policy is attached to the index, etc.
	if !es.cfg.UseDataStream {
		log.Debugf("[DEBUG] data streams disabled: creating initial index")

		indexName := fmt.Sprintf("%s-000001", es.cfg.AccessLogIndexName)
		iExists, err := es.client.Indices.Exists([]string{indexName})

		if err != nil {
			log.Error("[ERROR] there was an error while querying the template %v", err)
			return
		}

		if !iExists.IsError() {
			log.Debugf("[DEBUG] elasticsearch index %s exists: doing noting", indexName)
			return
		}

		var iTpl bytes.Buffer
		res, _ := getResource("files/elasticsearch/index_settings.json")
		gTpl := template.Must(template.New("it").Parse(string(res)))
		gTpl.Execute(&iTpl, es.cfg)

		resp, err := es.client.Indices.Create(indexName, es.client.Indices.Create.WithBody(bytes.NewReader(iTpl.Bytes())))
		if err != nil || resp.IsError() {
			log.Errorf("[ERROR] index template creation failed %v %v", err, resp)
			return
		}
	}

	// Attempt to configure Kibana's index patterns
	// and dashboards in the background.
	go es.ConfigureKibana()

}
