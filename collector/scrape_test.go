package collector

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_pagespeedService_Scrape(t *testing.T) {
	svc := newPagespeedScrapeService(60 * time.Second,"")
	res, err := svc.Scrape([]string{"https://www.google.com/"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
