package googleapi

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"bytes"
	"net/http"
	"net/http/httptest"
)

func TestParseResultFromData(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/exampleresult.json")
	if err != nil {
		t.Error(err)
	}

	resultData, errResultData := ParseResultFromData(data)
	if errResultData != nil {
		t.Error(errResultData)
	}

	assert.Equal(t, "https://www.globus.ch/", resultData.Target)
	assert.Equal(t, int64(61), resultData.PageStats.NumberResources)
	assert.Equal(t, int64(15), resultData.PageStats.NumberHosts)
	assert.Equal(t, int64(14381), resultData.PageStats.TotalRequestBytes)
	assert.Equal(t, int64(41), resultData.PageStats.NumberStaticResources)
	assert.Equal(t, int64(73765), resultData.PageStats.HtmlResponseBytes)
	assert.Equal(t, int64(17460), resultData.PageStats.TextResponseBytes)
	assert.Equal(t, int64(371007), resultData.PageStats.CssResponseBytes)
	assert.Equal(t, int64(653870), resultData.PageStats.ImageResponseBytes)
	assert.Equal(t, int64(1078707), resultData.PageStats.JavascriptResponseBytes)
	assert.Equal(t, int64(8356), resultData.PageStats.OtherResponseBytes)
	assert.Equal(t, int64(16), resultData.PageStats.NumberJsResources)
	assert.Equal(t, int64(3), resultData.PageStats.NumberCssResources)
	assert.Equal(t, int64(84), resultData.RuleGroups["SPEED"].Score)
}

func TestParseResultFromReader(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/exampleresult.json")
	if err != nil {
		t.Error(err)
	}
	reader := bytes.NewReader(data)

	readerData, errResultData := ParseResultFromReader(reader)
	if errResultData != nil {
		t.Error(errResultData)
	}

	resultData, errResultData := ParseResultFromData(data)
	if errResultData != nil {
		t.Error(errResultData)
	}

	assert.Equal(t, resultData, readerData)
}

func TestNewGoogleAPIService(t *testing.T) {
	service := NewGoogleAPIService("api-key")
	assert.Equal(t, "api-key", service.(Service).apiKey)
}

func TestGetPagespeedResults(t *testing.T) {
	http.DefaultClient.Transport = &mockTransport{}

	service := NewGoogleAPIService("api-key")
	result, errResult := service.GetPagespeedResults("testdata/exampleresult.json")
	if errResult != nil {
		t.Error(errResult)
	}

	assert.NotNil(t, result)
}

func TestGetPagespeedResultsBadRequest(t *testing.T) {
	http.DefaultClient.Transport = &mockTransport{}

	service := NewGoogleAPIService("api-key")
	_, errResult := service.GetPagespeedResults("nope.jpg")
	if errResult == nil {
		t.Error("there should be an error thrown")
	}

}

type mockTransport struct {
	http.RoundTripper
}

func (lt mockTransport) RoundTrip(r *http.Request) (res *http.Response, err error) {
	response := httptest.ResponseRecorder{}
	filepath := r.URL.Query().Get("url")
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		response.Body = bytes.NewBuffer([]byte("Not found"))
		response.WriteHeader(404)
	} else {
		response.Body = bytes.NewBuffer(data)
		response.WriteHeader(200)
	}
	return response.Result(), nil
}
