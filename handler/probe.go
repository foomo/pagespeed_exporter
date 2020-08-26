package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/foomo/pagespeed_exporter/collector"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultTimeoutDuration  = 30 * time.Second
	DefaultTimeOffset       = 500 * time.Millisecond // To Allow For Processing Time
	PrometheusTimeoutHeader = "X-Prometheus-Scrape-Timeout-Seconds"
)

type httpProbeHandler struct {
	googleAPIKey     string
	parallel         bool
	collectorFactory collector.Factory
	pushGatewayUrl   string
	pushGatewayJob   string
}

func NewProbeHandler(apiKey string, parallel bool, factory collector.Factory, pushGatewayUrl string, pushGatewayJob string) http.Handler {
	return httpProbeHandler{
		googleAPIKey:     apiKey,
		parallel:         parallel,
		collectorFactory: factory,
		pushGatewayUrl:   pushGatewayUrl,
		pushGatewayJob:   pushGatewayJob,
	}
}

func (ph httpProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	targets := r.URL.Query()["target"]

	for _, target := range targets {
		log.WithField("target", target).Info("probe requested for target")
	}

	requests := collector.CalculateScrapeRequests(targets...)
	if len(requests) == 0 {
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
		ScrapeRequests: requests,
		GoogleAPIKey:   ph.googleAPIKey,
		Parallel:       ph.parallel,
		ScrapeTimeout:  timeout,
	})
	if err != nil {
		errResponse(w, "Could not initialize pagespeed collectors", err)
		return
	}

	if err := registry.Register(psc); err != nil {
		errResponse(w, "Could not register collectors", err)
		return
	}

	if ph.pushGatewayUrl != "" {
		if err := push.New(ph.pushGatewayUrl, ph.pushGatewayJob).Collector(psc).Push(); err != nil {
			errResponse(w, "Error when tried to push to pushgateaway", err)
		} else {
			log.Info(fmt.Sprintf("pushed to pushgateway %s job %s with success", ph.pushGatewayUrl, ph.pushGatewayJob))
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func errResponse(w http.ResponseWriter, message string, err error) {
	log.WithError(err).Error(message)
	http.Error(w, message, http.StatusInternalServerError)
}

func getScrapeTimeout(r *http.Request) (timeout time.Duration, err error) {
	// If a timeout is configured via the Prometheus header, add it to the request.
	if v := r.Header.Get(PrometheusTimeoutHeader); v != "" {
		timeoutSeconds, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, errors.Wrap(err, "could not parse timeout")
		}
		if timeoutSeconds == 0 {
			return DefaultTimeoutDuration, nil
		}

		return time.Duration(timeoutSeconds*float64(time.Second)) - DefaultTimeOffset, nil
	}

	return DefaultTimeoutDuration, nil
}
