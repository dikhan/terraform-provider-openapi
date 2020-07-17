package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTelemetryProviderHttpEndpoint_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expectedErr error
	}{
		{
			name:        "happy path - host and port populated",
			url:         "http://telemetry.myhost.com/v1/metrics",
			expectedErr: nil,
		},
		{
			name:        "url is empty",
			url:         "",
			expectedErr: errors.New("http endpoint telemetry configuration is missing a value for the 'url property'"),
		},
		{
			name:        "url is wrongly formatter",
			url:         "htop://something-wrong.com",
			expectedErr: errors.New("http endpoint telemetry configuration does not have a valid URL 'htop://something-wrong.com'"),
		},
	}

	Convey("Given a TelemetryProviderHTTPEndpoint", t, func() {
		for _, tc := range testCases {
			tpg := TelemetryProviderHTTPEndpoint{
				URL: tc.url,
			}
			Convey(fmt.Sprintf("When Validate method is called: %s", tc.name), func() {
				err := tpg.Validate()
				Convey("Then the result returned should be the expected one", func() {
					So(err, ShouldResemble, tc.expectedErr)
				})
			})
		}
	})
}

func TestCreateNewCounterMetric(t *testing.T) {
	testCases := []struct {
		name           string
		prefix         string
		expectedMetric telemetryMetric
	}{
		{
			name:           "prefix is not empty",
			prefix:         "prefix",
			expectedMetric: telemetryMetric{metricTypeCounter, "prefix.metric_name", []string{"tag_name:tag_value"}},
		},
		{
			name:           "prefix is empty",
			prefix:         "",
			expectedMetric: telemetryMetric{metricTypeCounter, "metric_name", []string{"tag_name:tag_value"}},
		},
	}

	for _, tc := range testCases {
		Convey(fmt.Sprintf("When createNewCounterMetric method is called: %s", tc.name), t, func() {
			telemetryMetric := createNewCounterMetric(tc.prefix, "metric_name", []string{"tag_name:tag_value"})
			Convey("Then the result returned should be the expected one", func() {
				So(telemetryMetric, ShouldResemble, tc.expectedMetric)
			})
		})
	}

}

func TestCreateNewRequest(t *testing.T) {
	testCases := []struct {
		name                           string
		url                            string
		expectedCounterMetric          telemetryMetric
		expectedHeaders                map[string]string
		telemetryProviderConfiguration *telemetryProviderConfigurationHTTPEndpoint
		expectedErr                    error
	}{
		{
			name: "happy path - request is created with the expected Header and telemetryMetric",
			expectedCounterMetric: telemetryMetric{
				MetricType: metricTypeCounter,
				MetricName: "prefix.terraform.openapi_plugin_version.total_runs",
				Tags:       []string{"openapi_plugin_version:version"},
			},
			telemetryProviderConfiguration: &telemetryProviderConfigurationHTTPEndpoint{
				Headers: map[string]string{
					"property_name": "property_value",
				},
			},
			expectedHeaders: map[string]string{contentType: "application/json", userAgentHeader: "OpenAPI Terraform Provider", "property_name": "property_value"},
			expectedErr:     nil,
		},
		{
			name:        "crappy path - bad url",
			url:         "&^%",
			expectedErr: errors.New("parse &^%: invalid URL escape \"%\""),
		},
	}

	Convey("Given a TelemetryProviderHTTPEndpoint", t, func() {
		for _, tc := range testCases {
			var err error
			var request *http.Request
			var reqBody []byte

			tph := TelemetryProviderHTTPEndpoint{
				URL: tc.url,
			}
			Convey(fmt.Sprintf("When createNewRequest method is called: %s", tc.name), func() {
				request, err = tph.createNewRequest(tc.expectedCounterMetric, tc.telemetryProviderConfiguration)
				Convey("Then the result returned should be the expected one", func() {
					if tc.expectedErr != nil {
						So(err.Error(), ShouldResemble, tc.expectedErr.Error())
					} else {
						for expectedHeaderName, expectedHeaderValue := range tc.expectedHeaders {
							So(request.Header.Get(expectedHeaderName), ShouldContainSubstring, expectedHeaderValue)
						}
						reqBody, err = ioutil.ReadAll(request.Body)
						telemetryMetric := telemetryMetric{}
						err = json.Unmarshal(reqBody, &telemetryMetric)
						So(telemetryMetric, ShouldResemble, tc.expectedCounterMetric)
					}
				})
			})
		}
	})
}

func TestTelemetryProviderHttpEndpointSubmitMetric(t *testing.T) {
	testCases := []struct {
		testName                       string
		returnedResponseCode           int
		telemetryProviderConfiguration TelemetryProviderConfiguration
		expectedHeaders                map[string]string
		expectedErr                    error
	}{
		{
			testName:             "happy path with no telemetryProviderConfiguration",
			returnedResponseCode: http.StatusOK,
			expectedHeaders: map[string]string{
				contentType:     "application/json",
				userAgentHeader: "OpenAPI Terraform Provider",
			},
			telemetryProviderConfiguration: nil,
			expectedErr:                    nil,
		},
		{
			testName:             "happy path with expected telemetryProviderConfigurationHTTPEndpoint",
			returnedResponseCode: http.StatusOK,
			telemetryProviderConfiguration: telemetryProviderConfigurationHTTPEndpoint{
				Headers: map[string]string{
					"prop_name": "prop_value",
				},
			},
			expectedHeaders: map[string]string{
				contentType:     "application/json",
				userAgentHeader: "OpenAPI Terraform Provider",
				"prop_name":     "prop_value",
			},
			expectedErr: nil,
		},
		{
			testName:                       "happy path with wrong TelemetryProviderConfiguration",
			returnedResponseCode:           http.StatusOK,
			telemetryProviderConfiguration: struct{}{}, // random struct
			expectedErr:                    errors.New("telemetryProviderConfiguration object not the expected one: telemetryProviderConfigurationHTTPEndpoint"),
		},
		{
			testName:             "api server returns non 2xx code",
			returnedResponseCode: http.StatusNotFound,
			expectedErr:          errors.New("/v1/metrics' returned a non expected status code 404"),
		},
	}

	for _, tc := range testCases {

		expectedCounterMetric := telemetryMetric{
			MetricType: metricTypeCounter,
			MetricName: "prefix.terraform.openapi_plugin_version.total_runs",
			Tags:       []string{"openapi_plugin_version:version"},
		}

		api := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Method, http.MethodPost, tc.testName)
			assert.Equal(t, "/v1/metrics", req.URL.String(), tc.testName)
			for headerName, headerValue := range tc.expectedHeaders {
				assert.Contains(t, req.Header.Get(headerName), headerValue, tc.testName)
			}
			reqBody, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, tc.testName)
			telemetryMetric := telemetryMetric{}
			err = json.Unmarshal(reqBody, &telemetryMetric)
			assert.Nil(t, err, tc.testName)
			assert.Equal(t, expectedCounterMetric.MetricType, telemetryMetric.MetricType, tc.testName)
			assert.Equal(t, expectedCounterMetric.MetricName, telemetryMetric.MetricName, tc.testName)
			assert.Equal(t, expectedCounterMetric.Tags, telemetryMetric.Tags, tc.testName)
			rw.WriteHeader(tc.returnedResponseCode)
		}))
		// Close the server when test finishes
		defer api.Close()

		tph := TelemetryProviderHTTPEndpoint{
			URL: fmt.Sprintf("%s/v1/metrics", api.URL),
		}
		err := tph.submitMetric(expectedCounterMetric, tc.telemetryProviderConfiguration)
		if tc.expectedErr == nil {
			assert.NoError(t, err, tc.testName)
		} else {
			assert.Error(t, err, tc.testName)
			assert.Contains(t, err.Error(), tc.expectedErr.Error(), tc.testName)
		}
	}

}

func TestTelemetryProviderHttpEndpointSubmitMetricFailureScenarios(t *testing.T) {
	testCases := []struct {
		testName    string
		inputURL    string
		expectedErr error
	}{
		{
			testName:    "url is missing the protocol",
			inputURL:    "?",
			expectedErr: errors.New("request POST ? failed. Response Error: 'Post ?: unsupported protocol scheme \"\"'"),
		},
		{
			testName:    "url contains invalid characters",
			inputURL:    "&^%",
			expectedErr: errors.New("parse &^%: invalid URL escape \"%\""),
		},
	}

	for _, tc := range testCases {
		tph := TelemetryProviderHTTPEndpoint{
			URL: tc.inputURL,
		}
		err := tph.submitMetric(telemetryMetric{metricTypeCounter, "prefix.terraform.openapi_plugin_version.version.total_runs", []string{"openapi_plugin_version:version"}}, nil)
		assert.EqualError(t, err, tc.expectedErr.Error())
	}
}

func TestTelemetryProviderHttpEndpointIncOpenAPIPluginVersionTotalRunsCounter(t *testing.T) {
	testCases := []struct {
		testName             string
		returnedResponseCode int
		expectedErr          error
	}{
		{
			testName:             "happy path",
			returnedResponseCode: http.StatusOK,
			expectedErr:          nil,
		},
		{
			testName:             "metric submission fails",
			returnedResponseCode: http.StatusNotFound,
			expectedErr:          errors.New("/v1/metrics' returned a non expected status code 404"),
		},
	}

	for _, tc := range testCases {

		api := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			reqBody, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, tc.testName)
			telemetryMetric := telemetryMetric{}
			err = json.Unmarshal(reqBody, &telemetryMetric)
			assert.Nil(t, err, tc.testName)
			assert.Equal(t, metricTypeCounter, telemetryMetric.MetricType, tc.testName)
			assert.Equal(t, "terraform.openapi_plugin_version.total_runs", telemetryMetric.MetricName, tc.testName)
			assert.Equal(t, []string{"openapi_plugin_version:0_26_0"}, telemetryMetric.Tags, tc.testName)
			rw.WriteHeader(tc.returnedResponseCode)
		}))
		// Close the server when test finishes
		defer api.Close()

		tph := TelemetryProviderHTTPEndpoint{
			URL: fmt.Sprintf("%s/v1/metrics", api.URL),
		}
		err := tph.IncOpenAPIPluginVersionTotalRunsCounter("0.26.0", nil)
		if tc.expectedErr == nil {
			assert.NoError(t, err, tc.testName)
		} else {
			assert.Error(t, err, tc.testName)
			assert.Contains(t, err.Error(), tc.expectedErr.Error(), tc.testName)
		}
	}
}

func TestTelemetryProviderHttpEndpointIncServiceProviderResourceTotalRunsCounter(t *testing.T) {
	testCases := []struct {
		testName             string
		returnedResponseCode int
		expectedErr          error
	}{
		{
			testName:             "happy path",
			returnedResponseCode: http.StatusOK,
			expectedErr:          nil,
		},
		{
			testName:             "metric submission fails",
			returnedResponseCode: http.StatusNotFound,
			expectedErr:          errors.New("/v1/metrics' returned a non expected status code 404"),
		},
	}

	for _, tc := range testCases {

		api := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			reqBody, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, tc.testName)
			telemetryMetric := telemetryMetric{}
			err = json.Unmarshal(reqBody, &telemetryMetric)
			assert.Nil(t, err, tc.testName)
			assert.Equal(t, metricTypeCounter, telemetryMetric.MetricType, tc.testName)
			assert.Equal(t, "terraform.provider", telemetryMetric.MetricName, tc.testName)
			assert.Equal(t, []string{"provider_name:cdn", "resource_name:cdn_resource", fmt.Sprintf("terraform_operation:%s", TelemetryResourceOperationCreate)}, telemetryMetric.Tags, tc.testName)
			rw.WriteHeader(tc.returnedResponseCode)
		}))
		// Close the server when test finishes
		defer api.Close()

		tph := TelemetryProviderHTTPEndpoint{
			URL: fmt.Sprintf("%s/v1/metrics", api.URL),
		}
		err := tph.IncServiceProviderResourceTotalRunsCounter("cdn", "cdn_resource", TelemetryResourceOperationCreate, nil)
		if tc.expectedErr == nil {
			assert.NoError(t, err, tc.testName)
		} else {
			assert.Error(t, err, tc.testName)
			assert.Contains(t, err.Error(), tc.expectedErr.Error(), tc.testName)
		}
	}
}

func TestGetTelemetryProviderConfiguration(t *testing.T) {
	tp := TelemetryProviderHTTPEndpoint{
		ProviderSchemaProperties: []string{"prop_name"},
	}
	propSchema := newStringSchemaDefinitionPropertyWithDefaults("prop_name", "", true, false, "prop_value")
	testSchema := newTestSchema(propSchema)
	tpConfig := tp.GetTelemetryProviderConfiguration(testSchema.getResourceData(t))
	assert.IsType(t, telemetryProviderConfigurationHTTPEndpoint{}, tpConfig)
	assert.Equal(t, "prop_value", tpConfig.(telemetryProviderConfigurationHTTPEndpoint).Headers["prop_name"])
}
