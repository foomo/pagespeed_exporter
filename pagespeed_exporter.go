package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/foomo/pagespeed_exporter/collector"
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
	exp := collector.NewCollector(listenerAddress, googleApiKey, targets, checkInterval)
	log.Fatal(exp.Start())
}

func parseFlags() {
	flag.StringVar(&googleApiKey, "api-key", getenv("PAGESPEED_API_KEY", ""), "sets the google API key used for pagespeed")
	flag.StringVar(&listenerAddress, "listener", getenv("PAGESPEED_LISTENER", ":9271"), "sets the listener address for the exporters")
	targetsFlag := flag.String("targets", getenv("PAGESPEED_TARGETS", ""), "comma separated list of targets to measure")
	intervalFlag := flag.String("interval", getenv("PAGESPEED_INTERVAL", "1h"), "check interval (e.g. 3s 4h 5d ...)")

	flag.Parse()

	targets = strings.Split(*targetsFlag, ",")

	if googleApiKey == "" {
		log.Fatal("google api key parameter must be specified")
	}

	if len(targets) == 0 || targets[0] == "" {
		log.Fatal("at least one target must be specified for metrics")
	}

	if duration, err := time.ParseDuration(*intervalFlag); err != nil {
		log.Fatal("could not parse the interval flag '", intervalFlag, "'")
	} else {
		checkInterval = duration
	}
}

func getenv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
