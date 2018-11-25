package main

import (
	"flag"
	"github.com/foomo/pagespeed_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"

	"os"
	"time"
)

var (
	googleApiKey    string
	listenerAddress string
	targets         []string
	checkInterval   time.Duration
)

var (
	Version string
)

func main() {
	log.Infof("starting pagespeed exporter version %s", Version)

	parseFlags()
	psc, errCollector := collector.NewCollector(targets)
	if errCollector != nil {
		log.WithError(errCollector).Fatal("could not instantiate collector")
	}
	prometheus.MustRegister(psc)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listenerAddress, nil))
}

func parseFlags() {
	flag.StringVar(&googleApiKey, "api-key", getenv("PAGESPEED_API_KEY", ""), "sets the google API key used for pagespeed")
	flag.StringVar(&listenerAddress, "listener", getenv("PAGESPEED_LISTENER", ":9271"), "sets the listener address for the exporters")
	targetsFlag := flag.String("targets", getenv("PAGESPEED_TARGETS", ""), "comma separated list of targets to measure")

	flag.Parse()

	targets = strings.Split(*targetsFlag, ",")

	if googleApiKey == "" {
		log.Fatal("google api key parameter must be specified")
	}

	if len(targets) == 0 || targets[0] == "" {
		log.Fatal("at least one target must be specified for metrics")
	}
}

func getenv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
