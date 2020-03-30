package openapi

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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
		expectedCounterMetric telemetryMetric
		expectedContentType   string
		expectedUserAgent     string
	}{
		{
			testName: "happy path - request is created with the expected Header and telemetryMetric ",
			expectedCounterMetric: telemetryMetric{
				MetricType: metricTypeCounter,
				MetricName: "prefix.terraform.openapi_plugin_version.version.total_runs",
			},
			expectedContentType: "application/json",
			expectedUserAgent:   "OpenAPI Terraform Provider",
		},
	}

	for _, tc := range testCases {
		telemetryMetric := telemetryMetric{}
		tph := TelemetryProviderHttpEndpoint{}

		request, err := tph.createNewRequest(tc.expectedCounterMetric)
		assert.Nil(t, err)
		reqBody, err := ioutil.ReadAll(request.Body)
		assert.Nil(t, err)
		err = json.Unmarshal(reqBody, &telemetryMetric)
		assert.Nil(t, err)

		assert.Equal(t, tc.expectedContentType, request.Header.Get(contentType))
		assert.Contains(t, request.Header.Get(userAgentHeader), tc.expectedUserAgent)
		assert.Equal(t, tc.expectedCounterMetric, telemetryMetric)
	}
}
