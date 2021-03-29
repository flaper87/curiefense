package drivers

import (
	"curielog/pkg"
	"encoding/json"
	"log"
	"net/http"
	neturl "net/url"
)

//  ___ _   _   _ ___ _  _ _____ ___
// | __| | | | | | __| \| |_   _|   \
// | _|| |_| |_| | _|| .` | | | | |) |
// |_| |____\___/|___|_|\_| |_| |___/
// FLUENTD

type fluentdLogger struct {
	pkg.logger
}

func (l *fluentdLogger) Configure(channel_capacity int) error {
	l.name = "FluentD"
	ch := make(chan pkg.LogEntry, channel_capacity)
	l.channel = ch
	l.do_insert = l.InsertEntry
	return nil
}

func (l *fluentdLogger) InsertEntry(e pkg.LogEntry) bool {
	log.Printf("[DEBUG] Fluentd insertion!")
	j, err := json.Marshal(e.cfLog)
	if err == nil {
		_, err := http.PostForm(l.url+"curiefense.log", neturl.Values{"json": {string(j)}})
		if err != nil {
			log.Printf("ERROR: could not POST log entry: %v", err)
		}
	} else {
		log.Printf("[ERROR] Could not convert protobuf entry into json for ES insertion.")
	}
	return true
}
