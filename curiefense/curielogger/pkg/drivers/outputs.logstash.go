package drivers

import (
	"bytes"
	"curielog/pkg"
	"encoding/json"
	"log"
	"net/http"
)

//  _    ___   ___ ___ _____ _   ___ _  _
// | |  / _ \ / __/ __|_   _/_\ / __| || |
// | |_| (_) | (_ \__ \ | |/ _ \\__ \ __ |
// |____\___/ \___|___/ |_/_/ \_\___/_||_|
// LOGSTASH

type LogstashConfig struct {
	Enabled       bool                `mapstructure:"enabled"`
	Url           string              `mapstructure:"url"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
}

type logstashLogger struct {
	pkg.logger
	config LogstashConfig
}

func (l *logstashLogger) Configure(channel_capacity int) error {
	l.name = "Logstash"
	ch := make(chan pkg.LogEntry, channel_capacity)
	l.channel = ch
	l.do_insert = l.InsertEntry

	if l.config.Elasticsearch.Url != "" {
		log.Printf("[DEBUG] elasticsearch configs set, initializing configuration steps for %s", l.config.Elasticsearch.Url)
		es := ElasticsearchLogger{config: l.config.Elasticsearch}
		return es.Configure(0)
	}

	return nil
}

func (l *logstashLogger) InsertEntry(e pkg.LogEntry) bool {
	log.Printf("[DEBUG] LogStash insertion!")
	e.cfLog.Tags = append(e.cfLog.Tags, "curieaccesslog")
	j, err := json.Marshal(e.cfLog)
	if err == nil {
		_, err := http.Post(l.config.Url, "application/json", bytes.NewReader(j))
		if err != nil {
			log.Printf("ERROR: could not POST log entry: %v", err)
			return false
		}
	} else {
		log.Printf("[ERROR] Could not convert protobuf entry into json for ES insertion.")
		return false
	}
	return true
}
