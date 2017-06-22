package main

import (
	"flag"
	"strings"
	log "github.com/sirupsen/logrus"

	"github.com/foomo/pagespeed_exporter/collector"
)

var (
	googleApiKey string
	targets      []string
)

func main() {
	parseFlags()

	exp := collector.NewCollector(":9100", googleApiKey, targets)
	log.Fatal(exp.Start())
}

func parseFlags() {
	flag.StringVar(&googleApiKey, "api-key", "", "sets the google API key used for pagespeed")
	targetsFlag := flag.String("targets", "", "comma separated list of targets to measure")

	flag.Parse()

	targets = strings.Split(*targetsFlag, ",")

	if googleApiKey == "" {
		log.Fatal("google api key parameter must be specified")
	}

	if len(targets) == 0 {
		log.Fatal("at least one target must be specified for metrics")
	}
}
