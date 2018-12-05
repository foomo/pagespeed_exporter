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
			prometheus.Labels{"host": "host", "path": "/path", "strategy": string(StrategyMobile)}, false},

		{"valid desktop", getArgs("https://host/path", StrategyDesktop),
			prometheus.Labels{"host": "host", "path": "/path", "strategy": string(StrategyDesktop)}, false},

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
