package openapi

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestSubmitMetrics(t *testing.T) {
	stub := &telemetryProviderStub{}
	testCases := []struct {
		name            string
		ths             telemetryHandlerTimeoutSupport
		expectedLogging string
	}{
		{
			name: "submitMetrics works fine",
			ths: telemetryHandlerTimeoutSupport{
				providerName:   "providerName",
				timeout:        1,
				openAPIVersion: "0.25.0",
				telemetryProviders: []TelemetryProvider{
					stub,
				},
			},
			expectedLogging: "",
		},
	}
	for _, tc := range testCases {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		tc.ths.SubmitMetrics()
		assert.Contains(t, buf.String(), tc.expectedLogging, tc.name)
		// The below confirm that the corresponding inc methods were called and also the info passed in was the correct one
		assert.Equal(t, tc.ths.openAPIVersion, stub.openAPIPluginVersionReceived, tc.name)
		assert.Equal(t, tc.ths.providerName, stub.providerNameReceived, tc.name)
	}
}

func TestSubmitMetric(t *testing.T) {
	testCases := []struct {
		name                 string
		ths                  telemetryHandlerTimeoutSupport
		inputMetricName      string
		inputMetricSubmitter func() error
		expectedLogging      string
	}{
		{
			name: "submitMetric method is called with a metric name and a metric submitter that runs before the timeout",
			ths: telemetryHandlerTimeoutSupport{
				timeout: 1,
			},
			inputMetricName: "someMetricName",
			inputMetricSubmitter: func() error {
				return nil
			},
			expectedLogging: "",
		},
		{
			name: "submitMetric method is called with a metric name and a metric submitter timeout",
			ths: telemetryHandlerTimeoutSupport{
				timeout: 0,
			},
			inputMetricName: "someMetricName",
			inputMetricSubmitter: func() error {
				time.Sleep(2 * time.Second)
				return nil
			},
			expectedLogging: "metric 'someMetricName' submission did not finish within the expected time 0s\n",
		},
		{
			name: "submitMetric method is called with a metric name and a metric submitter errors out",
			ths: telemetryHandlerTimeoutSupport{
				timeout: 1,
			},
			inputMetricName: "someMetricName",
			inputMetricSubmitter: func() error {
				return errors.New("some error")
			},
			expectedLogging: "metric 'someMetricName' submission failed: some error",
		},
	}
	for _, tc := range testCases {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		tc.ths.submitMetric(tc.inputMetricName, tc.inputMetricSubmitter)
		assert.Contains(t, buf.String(), tc.expectedLogging, tc.name)
	}
}
