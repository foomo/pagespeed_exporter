package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/foomo/pagespeed_exporter/collector"
)

var (
	_ collector.Factory = mockCollector{}
)

type mockCollector struct {
}

func (mockCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (mockCollector) Collect(ch chan<- prometheus.Metric) {
	desc := prometheus.NewDesc("test", "test", nil, nil)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1)
}

func (mockCollector) Create(config collector.Config) (prometheus.Collector, error) {
	return mockCollector{}, nil
}

func TestProbeHandler(t *testing.T) {
	handler := NewProbeHandler("KEY", false, mockCollector{})
	require.NotNil(t, handler)

	require.HTTPSuccess(t, handler.ServeHTTP, "GET", "/probe", map[string][]string{"target": {"derp"}})
	require.HTTPBodyContains(t, handler.ServeHTTP, "GET", "/probe?target=derp", map[string][]string{"target": {"derp"}}, "test 1")
}

func Test_getScrapeTimeout(t *testing.T) {
	type args struct {
		headers http.Header
	}
	tests := []struct {
		name        string
		args        args
		wantTimeout time.Duration
		wantErr     bool
	}{
		{"default", args{}, DefaultTimeoutDuration, false},
		{"set", args{map[string][]string{"X-Prometheus-Scrape-Timeout-Seconds": {"30.5"}}}, 30 * time.Second, false},
		{"invalid", args{map[string][]string{"X-Prometheus-Scrape-Timeout-Seconds": {"derp"}}}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/", nil)
			request.Header = tt.args.headers

			gotTimeout, err := getScrapeTimeout(request)
			if (err != nil) != tt.wantErr {
				t.Errorf("getScrapeTimeout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTimeout, tt.wantTimeout) {
				t.Errorf("getScrapeTimeout() = %v, want %v", gotTimeout, tt.wantTimeout)
			}
		})
	}
}
