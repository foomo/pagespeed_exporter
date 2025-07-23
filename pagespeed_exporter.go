package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/foomo/pagespeed_exporter/collector"
	"github.com/foomo/pagespeed_exporter/handler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	Version string

	credentialsFile string
	googleApiKey    string
	listenerAddress string
	targets         arrayFlags
	categories      arrayFlags
	parallel        bool
	pushGatewayUrl  string
	pushGatewayJob  string
	cacheTTL        string // as duration string, e.g. "60s"
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

	log.Infof("starting pagespeed exporter version %s on address %s for %d targets and %d categories", Version, listenerAddress, len(targets), len(categories))

	collectorFactory := collector.NewFactory()
	// Register prometheus target collectors only if there is more than one target
	if len(targets) > 0 {
		requests := collector.CalculateScrapeRequests(targets, categories)

		var parsedCacheTTL time.Duration
		if cacheTTL != "" {
			var err error
			parsedCacheTTL, err = time.ParseDuration(cacheTTL)
			if err != nil {
				log.WithError(err).Warn("invalid CACHE_TTL, disabling cache")
				parsedCacheTTL = 0
			}
		}

		psc, errCollector := collectorFactory.Create(collector.Config{
			ScrapeRequests:  requests,
			GoogleAPIKey:    googleApiKey,
			CredentialsFile: credentialsFile,
			Parallel:        parallel,
			CacheTTL:        parsedCacheTTL,
		})
		if errCollector != nil {
			log.WithError(errCollector).Fatal("could not instantiate collector")
		}
		prometheus.MustRegister(psc)
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler.NewIndexHandler())
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/probe", handler.NewProbeHandler(credentialsFile, googleApiKey, parallel, collectorFactory, pushGatewayUrl, pushGatewayJob, categories))

	server := http.Server{
		Addr:    listenerAddress,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

func parseFlags() {
	flag.StringVar(&cacheTTL, "cache-ttl", getenv("CACHE_TTL", ""), "cache TTL for API results, e.g. 60s. If empty, disables cache")
	flag.StringVar(&googleApiKey, "api-key", getenv("PAGESPEED_API_KEY", ""), "sets the google API key used for pagespeed")
	flag.StringVar(&credentialsFile, "credentials-file", getenv("PAGESPEED_CREDENTIALS_FILE", ""), "sets the location of the credentials file used for pagespeed")
	flag.StringVar(&listenerAddress, "listener", getenv("PAGESPEED_LISTENER", ":9271"), "sets the listener address for the exporters")
	flag.BoolVar(&parallel, "parallel", getenv("PAGESPEED_PARALLEL", "false") == "true", "forces parallel execution for pagespeed")
	flag.StringVar(&pushGatewayUrl, "pushGatewayUrl", getenv("PUSHGATEWAY_URL", ""), "sets the push gateway to send the metrics. leave empty to ignore it")
	flag.StringVar(&pushGatewayJob, "pushGatewayJob", getenv("PUSHGATEWAY_JOB", "pagespeed_exporter"), "sets push gateway job name")
	targetsFlag := flag.String("targets", getenv("PAGESPEED_TARGETS", ""), "comma separated list of targets to measure")
	categoriesFlag := flag.String("categories", getenv("PAGESPEED_CATEGORIES", "accessibility,best-practices,performance,pwa,seo"), "comma separated list of categories. overridden by categories in JSON targets")
	flag.Var(&targets, "t", "multiple argument parameters")
	flag.Parse()

	if *targetsFlag != "" {
		additionalTargets := strings.Split(*targetsFlag, ",")
		targets = append(targets, additionalTargets...)
	}

	if *categoriesFlag != "" {
		additionalCategories := strings.Split(*categoriesFlag, ",")
		categories = append(categories, additionalCategories...)
	}

	if len(targets) == 0 || targets[0] == "" {
		log.Info("no targets specified, listening from collector")
	}
}

func getenv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
