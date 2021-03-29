package pkg

import (
	"curielog/curielogger"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	ald "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
	"github.com/prometheus/client_golang/prometheus"
)

//   ___ ___  __  __ __  __  ___  _  _
//  / __/ _ \|  \/  |  \/  |/ _ \| \| |
// | (_| (_) | |\/| | |\/| | (_) | .` |
//  \___\___/|_|  |_|_|  |_|\___/|_|\_|
// COMMON

type CurieProxyLog struct {
	Headers     map[string]string      `json:"headers"`
	Cookies     map[string]string      `json:"cookies"`
	Geo         map[string]interface{} `json:"geo"`
	Arguments   map[string]string      `json:"arguments"`
	Attributes  map[string]interface{} `json:"attributes"`
	Blocked     bool                   `json:"blocked"`
	BlockReason map[string]interface{} `json:"block_reason"`
	Tags        []string               `json:"tags"`
}

type RXTimer struct {
	FirstUpstreamByte float64 `json:"firstupstreambyte"`
	LastUpstreamByte  float64 `json:"lastupstreambyte"`
	LastByte          float64 `json:"lastbyte"`
}

type TXTimer struct {
	FirstUpstreamByte   float64 `json:"firstupstreambyte"`
	LastUpstreamByte    float64 `json:"lastupstreambyte"`
	FirstDownstreamByte float64 `json:"firstdownstreambyte"`
	LastDownstreamByte  float64 `json:"lastdownstreambyte"`
}

type DownstreamData struct {
	ConnectionTermination   bool   `json:"connectiontermination"`
	DirectRemoteAddress     string `json:"directremoteaddress"`
	DirectRemoteAddressPort uint32 `json:"directremoteaddressport"`
	LocalAddress            string `json:"localaddress"`
	LocalAddressPort        uint32 `json:"localaddressport"`
	ProtocolError           bool   `json:"protocolerror"`
	RemoteAddress           string `json:"remoteaddress"`
	RemoteAddressPort       uint32 `json:"remoteaddressport"`
}

type UpstreamData struct {
	Cluster                string `json:"cluster"`
	ConnectionFailure      bool   `json:"connectionfailure"`
	ConnectionTermination  bool   `json:"connectiontermination"`
	LocalAddress           string `json:"localaddress,omitempty"`
	LocalAddressPort       uint32 `json:"localaddressport,omitempty"`
	Overflow               bool   `json:"overflow"`
	RemoteAddress          string `json:"remoteaddress,omitempty"`
	RemoteAddressPort      uint32 `json:"remoteaddressport,omitempty"`
	RemoteReset            bool   `json:"remotereset"`
	RequestTimeout         bool   `json:"requesttimeout"`
	RetryLimitExceeded     bool   `json:"retrylimitexceeded"`
	TransportFailureReason string `json:"transportfailurereason"`
}

type CertificateData struct {
	Properties         string   `json:"properties"`
	PropertiesAltNames []string `json:"propertiesaltnames"`
}

type TLSData struct {
	LocalCertificate CertificateData `json:"localcertificate"`
	PeerCertificate  CertificateData `json:"peercertificate"`
	CipherSuite      string          `json:"ciphersuite"`
	SessionId        string          `json:"sessionid"`
	SNIHostname      string          `json:"snihostname"`
	Version          string          `json:"version"`
}

type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RequestData struct {
	BodyBytes    uint64                 `json:"bodybytes"`
	HeadersBytes uint64                 `json:"headersbytes"`
	OriginalPath string                 `json:"originalpath"`
	Headers      map[string]string      `json:"headers"`
	Cookies      map[string]string      `json:"cookies"`
	Arguments    map[string]string      `json:"arguments"`
	Geo          map[string]interface{} `json:"geo"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type ResponseData struct {
	BodyBytes    uint64            `json:"bodybytes"`
	Code         int               `json:"code"`
	CodeDetails  string            `json:"codedetails"`
	Headers      map[string]string `json:"headers"`
	HeadersBytes uint64            `json:"headersbytes"`
	Trailers     map[string]string `json:"trailers"`
}

type MetadataData struct {
	DelayInjected              bool    `json:"delayinjected"`
	FailedLocalHealthCheck     bool    `json:"failedlocalhealthcheck"`
	FaultInjected              bool    `json:"faultinjected"`
	InvalidEnvoyRequestHeaders bool    `json:"invalidenvoyrequestheaders"`
	LocalReset                 bool    `json:"localreset"`
	NoHealthyUpstream          bool    `json:"nohealthyupstream"`
	NoRouteFound               bool    `json:"noroutefound"`
	RateLimited                bool    `json:"ratelimited"`
	RateLimitServiceError      bool    `json:"ratelimitserviceerror"`
	RouteName                  string  `json:"routename"`
	SampleRate                 float64 `json:"samplerate"`
	StreamIdleTimeout          bool    `json:"streamidletimeout"`
	UnauthorizedDetails        string  `json:"unauthorizeddetails"`
}

type CuriefenseLog struct {
	RequestId string `json:"requestid"`
	Timestamp string `json:"timestamp"`
	Scheme    string `json:"scheme"`
	Authority string `json:"authority"`
	Port      uint32 `json:"port"`
	Method    string `json:"method"`
	Path      string `json:"path"`

	Blocked     bool                   `json:"blocked"`
	BlockReason map[string]interface{} `json:"block_reason"`
	Tags        []string               `json:"tags"`

	RXTimers RXTimer `json:"timers"`
	TXTimers TXTimer `json:"timers"`

	Upstream   UpstreamData   `json:"upstream"`
	Downstream DownstreamData `json:"downstream"`

	TLS      TLSData      `json:"tls"`
	Request  RequestData  `json:"request"`
	Response ResponseData `json:"response"`
	Metadata MetadataData `json:"metadata"`
}

type LogEntry struct {
	fullEntry     *ald.HTTPAccessLogEntry
	cfLog         *CuriefenseLog
	curieProxyLog *CurieProxyLog
}

type Logger interface {
	Configure(channel_capacity int) error
	ConfigureFromEnv(envVar string, channel_capacity int) error
	Start()
	GetLogEntry() LogEntry
	SendEntry(e LogEntry)
	InsertEntry(e LogEntry) bool
}

type logger struct {
	name      string
	channel   chan LogEntry
	url       string
	do_insert func(LogEntry) bool
}

func (l logger) SendEntry(e LogEntry) {
	if len(l.channel) >= cap(l.channel) {
		main.metric_dropped_log_entry.With(prometheus.Labels{"logger": l.name}).Inc()
		log.Printf("[WARNING] [%s] buffer full (%v/%v). Log entry dropped", l.name, len(l.channel), cap(l.channel))
	} else {
		l.channel <- e
	}
}

func (l *logger) Configure(channel_capacity int) error {
	l.name = "Generic Logger"
	l.channel = make(chan LogEntry, channel_capacity)
	return nil
}

func (l *logger) ConfigureFromEnv(envVar string, channel_capacity int) error {
	url, ok := os.LookupEnv(envVar)
	if !ok {
		return errors.New(fmt.Sprintf("Did not find %s in environment", envVar))
	}
	l.url = url
	return l.Configure(channel_capacity)
}

func (l logger) Start() {
	log.Printf("[INFO] %s logging routine started", l.name)
	for {
		entry := l.GetLogEntry()
		now := time.Now()
		l.do_insert(entry)
		main.metric_logger_latency.With(prometheus.Labels{"logger": l.name}).Observe(time.Since(now).Seconds())
	}
}

func (l logger) GetLogEntry() LogEntry {
	return <-l.channel
}

func (l logger) InsertEntry(e LogEntry) bool {
	log.Printf("[ERROR] entry insertion not implemented")
	return false
}
