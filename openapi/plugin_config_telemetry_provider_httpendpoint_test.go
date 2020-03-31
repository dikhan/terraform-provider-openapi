package openapi

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestTelemetryProviderHttpEndpoint_Validate(t *testing.T) {
	testCases := []struct {
		testName    string
		url         string
		expectedErr error
	}{
		{
			testName:    "happy path - host and port populated",
			url:         "http://telemetry.myhost.com/v1/metrics",
			expectedErr: nil,
		},
		{
			testName:    "url is empty",
			url:         "",
			expectedErr: errors.New("http endpoint telemetry configuration is missing a value for the 'url property'"),
		},
		{
			testName:    "url is wrongly formatter",
			url:         "htop://something-wrong.com",
			expectedErr: errors.New("http endpoint telemetry configuration does not have a valid URL 'htop://something-wrong.com'"),
		},
	}

	for _, tc := range testCases {
		tpg := TelemetryProviderHttpEndpoint{
			URL: tc.url,
		}
		err := tpg.Validate()
		assert.Equal(t, tc.expectedErr, err, tc.testName)
	}
}

func TestCreateNewRequest(t *testing.T) {
	testCases := []struct {
		testName              string
		url                   string
		expectedCounterMetric telemetryMetric
		expectedContentType   string
		expectedUserAgent     string
		expectedErr           error
	}{
		{
			testName: "happy path - request is created with the expected Header and telemetryMetric",
			expectedCounterMetric: telemetryMetric{
				MetricType: metricTypeCounter,
				MetricName: "prefix.terraform.openapi_plugin_version.version.total_runs",
			},
			expectedContentType: "application/json",
			expectedUserAgent:   "OpenAPI Terraform Provider",
			expectedErr:         nil,
		},
		{
			testName:    "crappy path - bad url",
			url:         "&^%",
			expectedErr: errors.New("parse &^%: invalid URL escape \"%\""),
		},
	}

	for _, tc := range testCases {
		var err error
		var request *http.Request
		var reqBody []byte
		telemetryMetric := telemetryMetric{}
		tph := TelemetryProviderHttpEndpoint{
			URL: tc.url,
		}

		request, err = tph.createNewRequest(tc.expectedCounterMetric)
		if tc.expectedErr != nil {
			assert.Equal(t, tc.expectedErr, errors.New(err.Error()), tc.testName)
		} else {
			reqBody, err = ioutil.ReadAll(request.Body)
			err = json.Unmarshal(reqBody, &telemetryMetric)
			assert.Nil(t, err, tc.testName)
			assert.Equal(t, tc.expectedContentType, request.Header.Get(contentType), tc.testName)
			assert.Contains(t, request.Header.Get(userAgentHeader), tc.expectedUserAgent, tc.testName)
			assert.Equal(t, tc.expectedCounterMetric, telemetryMetric, tc.testName)
		}
	}
}
