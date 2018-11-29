package scraper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_pagespeedService_Scrape(t *testing.T) {
	svc := New()
	res, err := svc.Scrape([]string{"https://www.globus.ch/"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
