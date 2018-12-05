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
)

var (
	googleApiKey    string
	listenerAddress string
	targets         arrayFlags
)

var (
	Version string
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	parseFlags()

	log.Infof("starting pagespeed exporter version %s on address %s for %d targets", Version, listenerAddress, len(targets))

	psc, errCollector := collector.NewCollector(targets, googleApiKey)
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
	flag.Var(&targets, "t", "Targets on a per-line basis")

	flag.Parse()

	additionalTargets := strings.Split(*targetsFlag, ",")
	targets = append(targets, additionalTargets...)

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
