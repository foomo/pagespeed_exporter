package main

import (
	"flag"
	"strings"
	log "github.com/sirupsen/logrus"

	"github.com/foomo/pagespeed_exporter/collector"
	"time"
)

var (
	googleApiKey    string
	listenerAddress string
	targets         []string
	checkInterval   time.Duration
)

func main() {
	parseFlags()

	exp := collector.NewCollector(listenerAddress, googleApiKey, targets, checkInterval)
	log.Fatal(exp.Start())
}

func parseFlags() {
	flag.StringVar(&googleApiKey, "api-key", "", "sets the google API key used for pagespeed")
	flag.StringVar(&listenerAddress, "listener", ":80", "sets the listener address for the exporters")
	targetsFlag := flag.String("targets", "", "comma separated list of targets to measure")
	intervalFlag := flag.String("interval", "30m", "check interval (e.g. 3s 4h 5d ...)")

	flag.Parse()

	targets = strings.Split(*targetsFlag, ",")

	if googleApiKey == "" {
		log.Fatal("google api key parameter must be specified")
	}

	if len(targets) == 0 {
		log.Fatal("at least one target must be specified for metrics")
	}

	if duration, err := time.ParseDuration(*intervalFlag); err != nil {
		log.Fatal("could not parse the interval flag '", intervalFlag, "'")
	} else {
		checkInterval = duration
	}
}
