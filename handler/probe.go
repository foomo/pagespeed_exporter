package handler

import (
	"context"
	"github.com/foomo/pagespeed_exporter/collector"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultTimeoutDuration = 30 * time.Second
	DefaultTimeOffset      = 500 * time.Millisecond // To Allow For Processing Time
)

type httpProbeHandler struct {
	googleAPIKey     string
	parallel         bool
	collectorFactory collector.Factory
}

func NewProbeHandler(apiKey string, parallel bool, factory collector.Factory) http.Handler {
	return httpProbeHandler{
		googleAPIKey:     apiKey,
		parallel:         parallel,
		collectorFactory: factory,
	}
}

func (ph httpProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	targets := getTargets(r)
	if len(targets) == 0 {
		http.Error(w, "Probe requires at least one target", http.StatusBadRequest)
		return
	}

	timeout, err := getScrapeTimeout(r)
	if err != nil {
		errResponse(w, "Could not parse scrape timeout", err)
		return
	}

	// set correct timeout without offset
	ctx, cancel := context.WithTimeout(context.Background(), timeout) //Offset to calculate inits
	defer cancel()
	r = r.WithContext(ctx)

	registry := prometheus.NewRegistry()

	psc, err := ph.collectorFactory.Create(collector.Config{
		Targets:       targets,
		GoogleAPIKey:  ph.googleAPIKey,
		Parallel:      ph.parallel,
		ScrapeTimeout: timeout,
	})
	if err != nil {
		errResponse(w, "Could not initialize pagespeed collectors", err)
		return
	}

	if err := registry.Register(psc); err != nil {
		errResponse(w, "Could not register collectors", err)
		return
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func errResponse(w http.ResponseWriter, message string, err error) {
	log.WithError(err).Error(message)
	http.Error(w, message, http.StatusInternalServerError)
}

func getTargets(r *http.Request) []string {
	return r.URL.Query()["target"]
}

func getScrapeTimeout(r *http.Request) (timeout time.Duration, err error) {
	// If a timeout is configured via the Prometheus header, add it to the request.
	var timeoutSeconds float64

	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		var err error
		timeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, errors.Wrap(err, "could not parse timeout")
		}
	}
	if timeoutSeconds == 0 {
		return DefaultTimeoutDuration, nil
	}

	return time.Duration(timeoutSeconds*float64(time.Second)) - DefaultTimeOffset, nil
}
