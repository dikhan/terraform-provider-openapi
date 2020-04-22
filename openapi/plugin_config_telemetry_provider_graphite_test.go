package openapi

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"strconv"
	"testing"
)

func TestTelemetryProviderGraphite_Validate(t *testing.T) {
	testCases := []struct {
		testName    string
		host        string
		port        int
		expectedErr error
	}{
		{
			testName:    "happy path - host and port populated",
			host:        "telemetry.myhost.com",
			port:        8125,
			expectedErr: nil,
		},
		{
			testName:    "crappy path - host is empty",
			host:        "",
			port:        8125,
			expectedErr: errors.New("graphite telemetry configuration is missing a value for the 'host property'"),
		},
		{
			testName:    "crappy path - port is 0",
			host:        "telemetry.myhost.com",
			port:        0,
			expectedErr: errors.New("graphite telemetry configuration is missing a valid value (>0) for the 'port' property'"),
		},
	}

	for _, tc := range testCases {
		tpg := TelemetryProviderGraphite{
			Host: tc.host,
			Port: tc.port,
		}
		err := tpg.Validate()
		assert.Equal(t, tc.expectedErr, err, tc.testName)
	}
}

func TestTelemetryProviderGraphite_IncOpenAPIPluginVersionTotalRunsCounter(t *testing.T) {
	openAPIPluginVersion := "0.25.0"
	expectedLogMetricToSubmit := "[INFO] graphite metric to be submitted: terraform.openapi_plugin_version.0_25_0.total_runs"
	expectedLogMetricSuccess := "[INFO] graphite metric successfully submitted: terraform.openapi_plugin_version.0_25_0.total_runs"
	expectedMetric := "myPrefixName.terraform.openapi_plugin_version.0_25_0.total_runs:1|c"

	var logging bytes.Buffer
	log.SetOutput(&logging)

	metricChannel := make(chan string)
	pc, telemetryHost, telemetryPort := udpServer(metricChannel)
	defer pc.Close()

	telemetryPortInt, err := strconv.Atoi(telemetryPort)
	tpg := TelemetryProviderGraphite{
		Host:   telemetryHost,
		Port:   telemetryPortInt,
		Prefix: "myPrefixName",
	}
	err = tpg.IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion, nil)
	assert.Nil(t, err)
	assertExpectedMetricAndLogging(t, metricChannel, expectedMetric, expectedLogMetricToSubmit, expectedLogMetricSuccess, &logging)
}

func TestTelemetryProviderGraphite_IncOpenAPIPluginVersionTotalRunsCounter_BadHost(t *testing.T) {
	openAPIPluginVersion := "0.25.0"
	expectedError := &net.DNSError{Err: "no such host", Name: "bad graphite host", Server: "", IsTimeout: false, IsTemporary: false}

	tpg := createTestGraphiteProviderBadHost()
	err := tpg.IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion, nil)

	assert.Equal(t, expectedError, err)
}

func TestTelemetryProviderGraphite_IncServiceProviderTotalRunsCounter(t *testing.T) {
	providerName := "myProviderName"
	expectedLogMetricToSubmit := "[INFO] graphite metric to be submitted: terraform.providers.myProviderName.total_runs"
	expectedLogMetricSuccess := "[INFO] graphite metric successfully submitted: terraform.providers.myProviderName.total_runs"
	expectedMetric := "myPrefixName.terraform.providers.myProviderName.total_runs:1|c"

	var logging bytes.Buffer
	log.SetOutput(&logging)

	metricChannel := make(chan string)
	pc, telemetryHost, telemetryPort := udpServer(metricChannel)
	defer pc.Close()

	telemetryPortInt, err := strconv.Atoi(telemetryPort)
	tpg := TelemetryProviderGraphite{
		Host:   telemetryHost,
		Port:   telemetryPortInt,
		Prefix: "myPrefixName",
	}
	err = tpg.IncServiceProviderTotalRunsCounter(providerName, nil)
	assert.Nil(t, err)
	assertExpectedMetricAndLogging(t, metricChannel, expectedMetric, expectedLogMetricToSubmit, expectedLogMetricSuccess, &logging)
}

func TestTelemetryProviderGraphite_IncServiceProviderTotalRunsCounter_BadHost(t *testing.T) {
	providerName := "myProviderName"
	expectedError := &net.DNSError{Err: "no such host", Name: "bad graphite host", Server: "", IsTimeout: false, IsTemporary: false}

	tpg := createTestGraphiteProviderBadHost()
	err := tpg.IncServiceProviderTotalRunsCounter(providerName, nil)

	assert.Equal(t, expectedError, err)
}

func TestTelemetryProviderGraphite_BuildMetricName(t *testing.T) {
	testCases := []struct {
		testName               string
		prefix                 string
		metricName             string
		expectedFullMetricName string
	}{
		{
			testName:               "happy path - with prefix",
			prefix:                 "myPrefixName",
			metricName:             "myMetricName",
			expectedFullMetricName: "myPrefixName.myMetricName",
		},
		{
			testName:               "happy path - without prefix",
			metricName:             "myMetricName",
			expectedFullMetricName: "myMetricName",
		},
	}

	for _, tc := range testCases {
		tpg := TelemetryProviderGraphite{
			Host:   "myTelemetryHost",
			Port:   8125,
			Prefix: tc.prefix,
		}

		fullMetricName := tpg.buildMetricName(tc.metricName)

		assert.Equal(t, tc.expectedFullMetricName, fullMetricName)
	}
}

func createTestGraphiteProviderBadHost() TelemetryProviderGraphite {
	tpg := TelemetryProviderGraphite{
		Host:   "bad graphite host",
		Port:   8125,
		Prefix: "myPrefixName",
	}
	return tpg
}
