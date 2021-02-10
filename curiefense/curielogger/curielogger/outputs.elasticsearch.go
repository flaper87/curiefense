package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
)

//  ___ _      _   ___ _____ ___ ___ ___ ___   _   ___  ___ _  _
// | __| |    /_\ / __|_   _|_ _/ __/ __| __| /_\ | _ \/ __| || |
// | _|| |__ / _ \\__ \ | |  | | (__\__ \ _| / _ \|   / (__| __ |
// |___|____/_/ \_\___/ |_| |___\___|___/___/_/ \_\_|_\\___|_||_|
// ELASTICSEARCH

const (
	INDEX_TEMPLATE = `
{
  "index_patterns": ["{{.AccessLogIndexName}}*"],
  "priority": 200,
  "composed_of": [],
  "version": 1,
  {{if .UseDataStream}}"data_stream": {},{{end}}
  "_meta": {
    "description": "Curiefense's index template "
  },
  "template": {
    "settings": {
      "index": {
        "lifecycle": {
          "name": "{{.AccessLogIndexName}}",
          "rollover_alias": "{{.AccessLogIndexName}}"
        }
      }
    },
    "mappings": {
      "dynamic_templates": [],
      "properties": {
        "authority": {
          "type": "keyword"
        },
        "block_reason": {
          "type": "object"
        },
        "blocked": {
          "type": "boolean"
        },
        "downstream": {
          "properties": {
            "connectiontermination": {
              "type": "boolean"
            },
            "directremoteaddress": {
              "type": "ip"
            },
            "directremoteaddressport": {
              "type": "integer"
            },
            "localaddress": {
              "type": "ip"
            },
            "localaddressport": {
              "type": "integer"
            },
            "protocolerror": {
              "type": "boolean"
            },
            "remoteaddress": {
              "type": "ip"
            },
            "remoteaddressport": {
              "type": "integer"
            }
          }
        },
        "metadata": {
          "properties": {
            "delayinjected": {
              "type": "boolean"
            },
            "failedlocalhealthcheck": {
              "type": "boolean"
            },
            "faultinjected": {
              "type": "boolean"
            },
            "invalidenvoyrequestheaders": {
              "type": "boolean"
            },
            "localreset": {
              "type": "boolean"
            },
            "nohealthyupstream": {
              "type": "boolean"
            },
            "noroutefound": {
              "type": "boolean"
            },
            "ratelimited": {
              "type": "boolean"
            },
            "ratelimitserviceerror": {
              "type": "boolean"
            },
            "routename": {
              "type": "keyword"
            },
            "samplerate": {
              "type": "float"
            },
            "streamidletimeout": {
              "type": "boolean"
            },
            "unauthorizeddetails": {
              "type": "keyword"
            }
          }
        },
        "method": {
          "type": "keyword"
        },
        "path": {
          "type": "keyword"
        },
        "port": {
          "type": "integer"
        },
        "request": {
          "properties": {
            "geo": {
              "type": "object",
              "properties": {
                "location": {
                  "type": "geo_point",
                  "ignore_malformed": true,
                  "ignore_z_value": true
                }
              }
            },
            "arguments": {
              "type": "flattened"
            },
            "attributes": {
              "type": "flattened"
            },
            "bodybytes": {
              "type": "integer"
            },
            "cookies": {
              "type": "flattened"
            },
            "headers": {
              "type": "flattened"
            },
            "headersbytes": {
              "type": "integer"
            },
            "originalpath": {
              "type": "keyword"
            }
          }
        },
        "requestid": {
          "type": "keyword"
        },
        "response": {
          "properties": {
            "bodybytes": {
              "type": "integer"
            },
            "code": {
              "type": "integer"
            },
            "codedetails": {
              "type": "keyword"
            },
            "headers": {
              "type": "keyword"
            },
            "headersbytes": {
              "type": "integer"
            },
            "trailers": {
              "type": "keyword"
            }
          }
        },
        "rxtimers": {
          "properties": {
            "firstupstreambyte": {
              "type": "float"
            },
            "lastbyte": {
              "type": "float"
            },
            "lastupstreambyte": {
              "type": "float"
            }
          }
        },
        "scheme": {
          "type": "keyword"
        },
        "tags": {
          "type": "keyword"
        },
        "@timestamp": {
          "type": "date_nanos"
        },
        "tls": {
          "properties": {
            "ciphersuite": {
              "type": "keyword"
            },
            "sessionid": {
              "type": "keyword"
            },
            "snihostname": {
              "type": "keyword"
            },
            "version": {
              "type": "keyword"
            }
          }
        },
        "txtimers": {
          "properties": {
            "firstdownstreambyte": {
              "type": "float"
            },
            "firstupstreambyte": {
              "type": "float"
            },
            "lastdownstreambyte": {
              "type": "float"
            },
            "lastupstreambyte": {
              "type": "float"
            }
          }
        },
        "upstream": {
          "properties": {
            "cluster": {
              "type": "keyword"
            },
            "connectionfailure": {
              "type": "boolean"
            },
            "connectiontermination": {
              "type": "boolean"
            },
            "localaddress": {
              "type": "ip"
            },
            "localaddressport": {
              "type": "integer"
            },
            "overflow": {
              "type": "boolean"
            },
            "remoteaddress": {
              "type": "ip"
            },
            "remoteaddressport": {
              "type": "integer"
            },
            "remotereset": {
              "type": "boolean"
            },
            "requesttimeout": {
              "type": "boolean"
            },
            "retrylimitexceeded": {
              "type": "boolean"
            },
            "transportfailurereason": {
              "type": "keyword"
            }
          }
        }
      }
    },
    "aliases": {}
  }
}}

`

	ILM_POLICY = `
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover" : {
            "max_age": "1d",
            "max_size": "50GB",
            "max_docs": 10000000
          }
        }
      },
      "delete": {
        "min_age": "7d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
	`

	INDEX_SETTINGS = `
{
  "aliases": {
    "{{.AccessLogIndexName}}": {
      "is_write_index": true
    }
  }
}
	`

	METRICS_ES_PREFIX   = "curiemetrics"
	ACCESSLOG_ES_PREFIX = "curieaccesslog"
)

type ElasticsearchConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	Url                string `mapstructure:"url"`
	Initialize         bool   `mapstructure:"initialize"`
	Overwrite          bool   `mapstructure:"overwrite"`
	AccessLogIndexName string `mapstructure:"accesslog_index_name"`
	UseDataStream      bool   `mapstructure:"use_data_stream"`
}

type ElasticsearchLogger struct {
	logger
	client *elasticsearch.Client
	config ElasticsearchConfig
}

func (l ElasticsearchLogger) getESClient() *elasticsearch.Client {
	for l.client == nil {
		cfg := elasticsearch.Config{
			Addresses: []string{l.config.Url},
		}
		conn, err := elasticsearch.NewClient(cfg)
		if err == nil {
			l.client = conn
			log.Printf("[DEBUG] Connected to elasticsearch %v\n", l.config.Url)
			break
		}
		log.Printf("[ERROR] Could not connect to elasticsearch [%v]: %v\n", l.config.Url, err)
		time.Sleep(time.Second)
	}
	return l.client
}

func (l *ElasticsearchLogger) Configure(channel_capacity int) error {
	l.name = "Elasticsearch"
	ch := make(chan LogEntry, channel_capacity)
	l.channel = ch
	l.do_insert = l.InsertEntry

	if l.config.AccessLogIndexName == "" {
		l.config.AccessLogIndexName = ACCESSLOG_ES_PREFIX
	}

	if !l.config.Initialize {
		return nil
	}

	log.Printf("[DEBUG] Initializing Elasticsearch configs for %s\n", l.config.Url)
	client := l.getESClient()

	// Initialize ILM Policy
	res, err := client.ILM.GetLifecycle()

	if err != nil {
		log.Printf("[ERROR] There was an error while querying the ILM Policies %v", err)
		return nil
	}

	var ilm map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&ilm); err != nil {
		return nil
	}

	_, exists := ilm[l.config.AccessLogIndexName]
	if l.config.Overwrite || !exists {
		log.Printf("[DEBUG] creating / overwriting elasticsearch ilm policy %s for %s\n", l.config.AccessLogIndexName, l.config.Url)

		var iTpl bytes.Buffer
		gTpl := template.Must(template.New("it").Parse(ILM_POLICY))
		gTpl.Execute(&iTpl, l.config)

		body := client.ILM.PutLifecycle.WithBody(bytes.NewReader(iTpl.Bytes()))
		resp, err := client.ILM.PutLifecycle(l.config.AccessLogIndexName, body)
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
	tplExists, err := client.Indices.ExistsIndexTemplate(l.config.AccessLogIndexName)

	if err != nil {
		log.Printf("[ERROR] there was an error while querying the template %v", err)
		return nil
	}

	if l.config.Overwrite || tplExists.IsError() {
		log.Printf("[DEBUG] creating / overwriting elasticsearch index template %s for %s\n", ACCESSLOG_ES_PREFIX, l.config.Url)
		var iTpl bytes.Buffer
		gTpl := template.Must(template.New("it").Parse(INDEX_TEMPLATE))
		gTpl.Execute(&iTpl, l.config)

		resp, err := client.Indices.PutIndexTemplate(l.config.AccessLogIndexName, bytes.NewReader(iTpl.Bytes()))
		if err != nil || resp.IsError() {
			log.Printf("[ERROR] index template creation failed %v %v", err, resp)
		}
	}

	// Data streams take care of creating the initila index, assigning an ILM policy
	// to it, and all the internal management. The index will be `hidden` and prefixed
	// with `.ds` so, in kibana, it is necessary to flag the "show hidden indeces" option.
	//
	// For non data stream configs, we have to create the initial index to make sure the
	// alias is assigned, the policy is attached to the index, etc.
	if !l.config.UseDataStream {
		log.Printf("[DEBUG] data streams disabled: creating initial index")

		indexName := fmt.Sprintf("%s-000001", l.config.AccessLogIndexName)
		iExists, err := client.Indices.Exists([]string{indexName})

		if err != nil {
			log.Printf("[ERROR] there was an error while querying the template %v", err)
			return nil
		}

		if !iExists.IsError() {
			log.Printf("[DEBUG] elasticsearch index %s exists: doing noting", indexName)
			return nil
		}

		var iTpl bytes.Buffer
		gTpl := template.Must(template.New("it").Parse(INDEX_SETTINGS))
		gTpl.Execute(&iTpl, l.config)

		resp, err := client.Indices.Create(indexName, client.Indices.Create.WithBody(bytes.NewReader(iTpl.Bytes())))
		if err != nil || resp.IsError() {
			log.Printf("[ERROR] index template creation failed %v %v", err, resp)
		}
	}

	return nil
}

func (l *ElasticsearchLogger) InsertEntry(e LogEntry) bool {
	log.Printf("[DEBUG] ES insertion!")
	client := l.getESClient() // needed to ensure ES cnx is not closed and reopen it if needed
	j, err := json.Marshal(e.cfLog)
	if err == nil {
		client.Index(
			"log",
			strings.NewReader(string(j)),
			client.Index.WithRefresh("true"),
			client.Index.WithPretty(),
			client.Index.WithFilterPath("result", "_id"),
		)
	} else {
		log.Printf("[ERROR] Could not convert protobuf entry into json for ES insertion.")
		return false
	}
	return true
}
