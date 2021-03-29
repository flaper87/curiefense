package drivers

import (
	"curielog/pkg"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//  ___ ___  ___  __  __ ___ _____ _  _ ___ _   _ ___
// | _ \ _ \/ _ \|  \/  | __|_   _| || | __| | | / __|
// |  _/   / (_) | |\/| | _|  | | | __ | _|| |_| \__ \
// |_| |_|_\\___/|_|  |_|___| |_| |_||_|___|\___/|___/
// PROMETHEUS

const (
	namespace = "curiemetric" // For Prometheus metrics.
)

/** Prometheus metrics **/

var (
	metric_requests = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_total",
			Help:      "Total number of HTTP requests",
		},
	)

	metric_session_details = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "session_details_total",
		Help:      "number of requests per label",
	}, []string{
		"status_code",
		"status_class",
		"origin",
		"origin_status_code",
		"origin_status_class",
		"method",
		"path",
		"blocked",
		"asn",
		"geo",
		"aclid",
		"aclname",
		"wafid",
		"wafname",
		"urlmap",
		"urlmap_entry",
		"container",
	})

	metric_dropped_log_entry = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "dropped_log_entries",
		Help:      "number of dropped log entries per logger",
	}, []string{"logger"})

	metric_request_bytes = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "request_bytes",
		Help:      "The total number of request bytes",
	})
	metric_response_bytes = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "response_bytes",
		Help:      "The total number of response bytes",
	})

	metric_requests_tags = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "session_tags_total",
		Help:      "Number of requests per label",
	}, []string{"tag"})

	metric_logger_latency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "logger_latency",
		Help:      "latency per logger",
	}, []string{"logger"})
)

/**** \\\ auto labeling /// ****/

func isStaticTag(tag string) bool {
	if tag == "all" {
		return true
	}
	parts := strings.Split(tag, ":")
	if len(parts) > 1 {
		prefix := parts[0]
		var static_tags = map[string]bool{
			"ip":           true,
			"asn":          true,
			"geo":          true,
			"aclid":        true,
			"aclname":      true,
			"wafid":        true,
			"wafname":      true,
			"urlmap":       true,
			"urlmap-entry": true,
			"container":    true,
		}
		return static_tags[prefix]
	}
	return false
}

func extractTagByPrefix(prefix string, tags map[string]interface{}) string {

	for name := range tags {
		tagsplit := strings.Split(name, ":")
		if len(tagsplit) == 2 {
			tag_prefix, value := tagsplit[0], tagsplit[1]
			if tag_prefix == prefix {
				return value
			}
		}

	}

	return "N/A"
}

func makeTagMap(tags []string) map[string]string {
	res := make(map[string]string)
	for _, k := range tags {
		tspl := strings.Split(k, ":")
		if len(tspl) == 2 {
			res[tspl[0]] = tspl[1]
		}
	}
	return res
}

func makeLabels(status_code int, method, path, upstream, blocked string, tags []string) prometheus.Labels {

	// classes and specific response code
	// icode := int(status_code)
	class_label := "status_Nxx"

	switch {
	case status_code < 200:
		class_label = "status_1xx"
	case status_code > 199 && status_code < 300:
		class_label = "status_2xx"
	case status_code > 299 && status_code < 400:
		class_label = "status_3xx"
	case status_code > 399 && status_code < 500:
		class_label = "status_4xx"
	case status_code > 499 && status_code < 600:
		class_label = "status_5xx"
	}

	status_code_str := strconv.Itoa(status_code)

	origin := "N/A"
	origin_status_code := "N/A"
	origin_status_class := "N/A"

	if len(upstream) > 0 {
		origin = upstream
		origin_status_code = fmt.Sprintf("origin_%s", status_code_str)
		origin_status_class = fmt.Sprintf("origin_%s", class_label)
	}

	tm := makeTagMap(tags)

	return prometheus.Labels{
		"status_code":         status_code_str,
		"status_class":        class_label,
		"origin":              origin,
		"origin_status_code":  origin_status_code,
		"origin_status_class": origin_status_class,
		"method":              method,
		"path":                path,
		"blocked":             blocked,
		"asn":                 tm["asn"],
		"geo":                 tm["geo"],
		"aclid":               tm["aclid"],
		"aclname":             tm["aclname"],
		"wafid":               tm["wafid"],
		"wafname":             tm["wafname"],
		"urlmap":              tm["urlmap"],
		"urlmap_entry":        tm["urlmap-entry"],
		"container":           tm["container"],
	}
}

type promLogger struct {
	pkg.logger
}

func (l *promLogger) Configure(channel_capacity int) error {
	l.name = "Prometheus"
	ch := make(chan pkg.LogEntry, channel_capacity)
	l.channel = ch
	l.do_insert = l.InsertEntry
	return nil
}

func (l promLogger) Start() {
	log.Printf("[INFO] Prometheus (%s) metrics updating routine started", l.name)

	for {
		e := l.GetLogEntry()
		log.Printf("[DEBUG] new log entry cflog=%v", *e.cfLog)
		// **** Update prometheus metrics ****
		metric_requests.Inc()
		metric_request_bytes.Add(float64(e.cfLog.Request.HeadersBytes + e.cfLog.Request.BodyBytes))
		metric_response_bytes.Add(float64(e.cfLog.Response.HeadersBytes + e.cfLog.Response.BodyBytes))

		tags := e.cfLog.Tags
		blocked := strconv.FormatBool(e.cfLog.Blocked)

		labels := makeLabels(e.cfLog.Response.Code, e.cfLog.Method, e.cfLog.Path,
			e.cfLog.Upstream.RemoteAddress, blocked, tags)
		metric_session_details.With(labels).Inc()

		for _, name := range tags {
			if !isStaticTag(name) {
				metric_requests_tags.WithLabelValues(name).Inc()
			}
		}
	}
}
