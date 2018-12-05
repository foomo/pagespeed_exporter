package collector

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_getConstLabels(t *testing.T) {
	type args struct {
		scrape *Scrape
	}
	getArgs := func(target string, strategy Strategy) args {
		return args{&Scrape{Target: target, Strategy: strategy}}
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
