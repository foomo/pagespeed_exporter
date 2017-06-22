package googleapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Strategy string

const (
	StrategyMobile  = Strategy("mobile")
	StrategyDesktop = Strategy("desktop")
)

const (
	pagespeedApiTemplate = "https://www.googleapis.com/pagespeedonline/v2/runPagespeed?url=%s&key=%s&strategy=%s&prettyprint=false"
)

type APIService interface {
	GetPagespeedResults(target string, strategy Strategy) (result *Result, err error)
}

type Service struct {
	apiKey string
}

func NewGoogleAPIService(apiKey string) APIService {
	return Service{
		apiKey: apiKey,
	}
}
func (service Service) GetPagespeedResults(target string, strategy Strategy) (result *Result, err error) {
	requestUrl := fmt.Sprintf(pagespeedApiTemplate, target, service.apiKey, strategy)
	resp, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New("Invalid response status " + resp.Status + " and body " + string(responseBody))
	}

	result, err = ParseResultFromReader(resp.Body)
	if result != nil {
		result.Strategy = strategy
	}
	return
}

func ParseResultFromReader(reader io.Reader) (result *Result, err error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ParseResultFromData(data)
}

func ParseResultFromData(data []byte) (result *Result, err error) {
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return
}
