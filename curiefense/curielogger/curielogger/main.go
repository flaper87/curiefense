package main

import (
	"curielog/pkg"
	"curielog/pkg/drivers"
	"encoding/json"

	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	als "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	duration "github.com/golang/protobuf/ptypes/duration"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"

	"net/http"

	"github.com/hashicorp/logutils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//   ___ ___ ___  ___     _   ___ ___ ___ ___ ___   _    ___   ___ ___
//  / __| _ \ _ \/ __|   /_\ / __/ __| __/ __/ __| | |  / _ \ / __/ __|
// | (_ |   /  _/ (__   / _ \ (_| (__| _|\__ \__ \ | |_| (_) | (_ \__ \
//  \___|_|_\_|  \___| /_/ \_\___\___|___|___/___/ |____\___/ \___|___/
// GRPC ACCESS LOGS

func DurationToFloat(d *duration.Duration) float64 {
	if d != nil {
		return float64(d.GetSeconds()) + float64(d.GetNanos())*1e-9
	}
	return 0
}

func TimestampToRFC3339(d *timestamp.Timestamp) string {
	var v time.Time
	if d != nil {
		v = time.Unix(int64(d.GetSeconds()), int64(d.GetNanos()))
	} else {
		v = time.Now()
	}
	return v.Format(time.RFC3339Nano)
}

func MapToNameValue(m map[string]string) []pkg.NameValue {
	var res []pkg.NameValue
	for k, v := range m {
		res = append(res, pkg.NameValue{k, v})
	}
	return res
}

//  __  __   _   ___ _  _
// |  \/  | /_\ |_ _| \| |
// | |\/| |/ _ \ | || .` |
// |_|  |_/_/ \_\___|_|\_|
// MAIN

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func check_env_flag(env_var string) bool {
	value, ok := os.LookupEnv(env_var)
	return ok && value != "" && value != "0" && strings.ToLower(value) != "false"
}

func readPassword(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		return scanner.Text()
	}
	log.Fatal("Could not read password")
	return ""
}

var (
	grpc_addr = getEnv("CURIELOGGER_GRPC_LISTEN", ":9001")
	prom_addr = getEnv("CURIELOGGER_PROMETHEUS_LISTEN", ":2112")
)

func main() {
	log.Print("Starting curielogger v0.3-dev1")

	pflag.String("log_level", "info", "Debug mode")
	pflag.Int("channel_capacity", 65536, "log channel capacity")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	config, err := pkg.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// configure log level
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(config.LogLevel)),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	log.Printf("[INFO] Log level set at %v", config.LogLevel)
	log.Printf("[INFO] Channel capacity set at %v", config.ChannelCapacity)

	// set up prometheus server
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Prometheus exporter listening on %v", prom_addr)
	go http.ListenAndServe(prom_addr, nil)

	////////////////////
	// set up loggers //
	////////////////////

	grpcParams := grpcServerParams{loggers: []pkg.Logger{}}

	// Prometheus
	if check_env_flag("CURIELOGGER_METRICS_PROMETHEUS_ENABLED") {
		prom := drivers.promLogger{pkg.logger{name: "prometheus", channel: make(chan pkg.LogEntry, config.ChannelCapacity)}}
		grpcParams.loggers = append(grpcParams.loggers, &prom)
	}

	configRetry := func(params *grpcServerParams, logger pkg.Logger) {
		for i := 0; i < 60; i++ {
			err := logger.Configure(config.ChannelCapacity)

			if err == nil {
				grpcParams.loggers = append(params.loggers, logger)
				logger.Start()
				break
			}

			log.Printf("[ERROR]: failed to configure logger (retrying in 5s) %v %v", logger, err)
			time.Sleep(5 * time.Second)
		}
	}

	// ElasticSearch
	if config.Outputs.Elasticsearch.Enabled {
		log.Printf("[DEBUG] Elasticsearch enabled with URL: %s", config.Outputs.Elasticsearch.Url)
		es := drivers.ElasticsearchLogger{config: config.Outputs.Elasticsearch}
		go configRetry(&grpcParams, &es)
	}

	// Logstash
	if config.Outputs.Logstash.Enabled {
		log.Printf("[DEBUG] Logstash enabled with URL: %s", config.Outputs.Logstash.Url)
		ls := drivers.logstashLogger{config: config.Outputs.Logstash}
		go configRetry(&grpcParams, &ls)
	}

	// Webhoook
	if config.Outputs.Webhook.Enabled {
		log.Printf("[DEBUG] Webhook enabled with URL: %s", config.Outputs.Webhook.Url)
		wh := webhookLogger{config: config.Outputs.Webhook}
		go configRetry(&grpcParams, &wh)
	}

	// Fluentd
	if check_env_flag("CURIELOGGER_USES_FLUENTD") {
		fd := drivers.fluentdLogger{}
		fd.ConfigureFromEnv("CURIELOGGER_FLUENTD_URL", config.ChannelCapacity)
		grpcParams.loggers = append(grpcParams.loggers, &fd)
	}

	for _, l := range grpcParams.loggers {
		go l.Start()
	}

	////////////////////////
	// set up GRPC server //
	////////////////////////

	sock, err := net.Listen("tcp", grpc_addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("GRPC server listening on %v", grpc_addr)
	s := grpc.NewServer()

	als.RegisterAccessLogServiceServer(s, &grpcParams)
	if err := s.Serve(sock); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
