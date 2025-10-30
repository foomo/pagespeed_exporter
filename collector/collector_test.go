package collector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/api/option"
)

func loadTestEnv() {
	envPath := filepath.Join("..", ".env")
	_ = godotenv.Load(envPath)
}

func Test_getConstLabels(t *testing.T) {
	type args struct {
		scrape *ScrapeResult
	}
	getArgs := func(target string, strategy Strategy) args {
		return args{&ScrapeResult{Request: ScrapeRequest{Url: target, Strategy: strategy}}}
	}
	tests := []struct {
		name       string
		args       args
		wantLabels prometheus.Labels
		wantErr    bool
	}{
		{"valid mobile", getArgs("https://host/path", StrategyMobile),
			prometheus.Labels{"host": "https://host", "path": "/path", "strategy": string(StrategyMobile)}, false},

		{"valid desktop", getArgs("https://host/path", StrategyDesktop),
			prometheus.Labels{"host": "https://host", "path": "/path", "strategy": string(StrategyDesktop)}, false},

		{"valid query", getArgs("https://host/path?some=query", StrategyDesktop),
			prometheus.Labels{"host": "https://host", "path": "/path?some=query", "strategy": string(StrategyDesktop)}, false},

		{"invalid url", getArgs("http://[fe80::1%en0]:8080/", StrategyMobile),
			nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLabels, err := getConstLabels(tt.args.scrape)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConstLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotLabels, tt.wantLabels) {
				t.Errorf("getConstLabels() = %v, want %v", gotLabels, tt.wantLabels)
			}
		})
	}
}

func Test_convertCategoryToScore(t *testing.T) {
	type args struct {
		category string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"average", args{"AVERAGE"}, 0.5},
		{"fast", args{"FAST"}, 1},
		{"none", args{"NONE"}, -1},
		{"jibberish", args{""}, -1},
		{"slow", args{"SLOW"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertCategoryToScore(tt.args.category); got != tt.want {
				t.Errorf("convertCategoryToScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fqname(t *testing.T) {
	type args struct {
		values []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{[]string{}}, "pagespeed_"},
		{"single", args{[]string{"one"}}, "pagespeed_one"},
		{"multi", args{[]string{"one", "two", "three"}}, "pagespeed_one_two_three"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fqname(tt.args.values...); got != tt.want {
				t.Errorf("fqname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_CollectorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	loadTestEnv()

	var options []option.ClientOption

	if apiKey := os.Getenv("PAGESPEED_API_KEY"); apiKey != "" {
		options = append(options, option.WithAPIKey(apiKey))
	}

	if cf := os.Getenv("PAGESPEED_CREDENTIALS_FILE"); cf != "" {
		options = append(options, option.WithCredentialsFile(cf))
	}

	if len(options) == 0 {
		t.Skip("skipping integration test unless PAGESPEED_API_KEY or PAGESPEED_CREDENTIALS_FILE is set")
	}

	config := Config{
		ScrapeRequests: []ScrapeRequest{
			{
				Url:        "https://www.example.com",
				Strategy:   StrategyMobile,
				Categories: []string{CategoryPerformance},
			},
		},
		GoogleAPIKey: os.Getenv("PAGESPEED_API_KEY"),
		Parallel:     false,
	}

	if cf := os.Getenv("PAGESPEED_CREDENTIALS_FILE"); cf != "" {
		config.CredentialsFile = cf
	}

	coll, err := newCollector(config)
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}

	ch := make(chan prometheus.Metric, 100)
	done := make(chan bool)

	var metrics []prometheus.Metric
	go func() {
		for m := range ch {
			metrics = append(metrics, m)
		}
		done <- true
	}()

	coll.Collect(ch)
	close(ch)
	<-done

	if len(metrics) == 0 {
		t.Fatal("expected at least one metric to be collected")
	}

	t.Logf("Successfully collected %d metrics from https://www.example.com", len(metrics))

	hasExpectedLabels := false
	for _, m := range metrics {
		desc := m.Desc().String()
		if contains(desc, "host") && contains(desc, "path") && contains(desc, "strategy") {
			hasExpectedLabels = true
			break
		}
	}

	if !hasExpectedLabels {
		t.Error("expected metrics to contain host, path, and strategy labels")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
