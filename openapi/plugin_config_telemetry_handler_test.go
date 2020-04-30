package openapi

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestSubmitPluginExecutionMetrics(t *testing.T) {
	stub := &telemetryProviderStub{}
	ths := telemetryHandlerTimeoutSupport{
		providerName:      "providerName",
		timeout:           1,
		openAPIVersion:    "0.25.0",
		telemetryProvider: stub,
	}
	ths.SubmitPluginExecutionMetrics()
	// The below confirm that the corresponding inc methods were called and also the info passed in was the correct one
	assert.Equal(t, ths.openAPIVersion, stub.openAPIPluginVersionReceived)
}

func TestSubmitPluginExecutionMetrics_FailsNilTelemetryProvider(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	ths := telemetryHandlerTimeoutSupport{
		providerName:      "providerName",
		timeout:           1,
		openAPIVersion:    "0.25.0",
		telemetryProvider: nil,
	}
	ths.SubmitPluginExecutionMetrics()
	assert.Contains(t, buf.String(), "[INFO] Telemetry provider not configured")
}

func TestSubmitResourceExecutionMetrics(t *testing.T) {
	expectedResourceName := "resourceName"
	expectedTfOperation := TelemetryResourceOperationCreate
	stub := &telemetryProviderStub{}
	ths := telemetryHandlerTimeoutSupport{
		providerName:      "providerName",
		timeout:           1,
		openAPIVersion:    "0.25.0",
		telemetryProvider: stub,
	}
	ths.SubmitResourceExecutionMetrics(expectedResourceName, expectedTfOperation)
	// The below confirm that the corresponding inc methods were called and also the info passed in was the correct one
	assert.Equal(t, ths.providerName, stub.providerNameReceived)
	assert.Equal(t, expectedResourceName, stub.resourceNameReceived)
	assert.Equal(t, expectedTfOperation, stub.tfOperationReceived)
}

func TestSubmitResourceExecutionMetrics_FailsNilTelemetryProvider(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	ths := telemetryHandlerTimeoutSupport{
		providerName:      "providerName",
		timeout:           1,
		openAPIVersion:    "0.25.0",
		telemetryProvider: nil,
	}
	ths.SubmitResourceExecutionMetrics("resourceName", TelemetryResourceOperationCreate)
	assert.Contains(t, buf.String(), "[INFO] Telemetry provider not configured")
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

func TestSubmitTelemetryMetric(t *testing.T) {
	var resourceNameReceived string
	var tfOperationReceived TelemetryResourceOperation
	clientOpenAPI := &clientOpenAPIStub{
		telemetryHandler: &telemetryHandlerStub{
			submitResourceExecutionMetricsFunc: func(resourceName string, tfOperation TelemetryResourceOperation) {
				resourceNameReceived = resourceName
				tfOperationReceived = tfOperation
			},
		},
	}
	specResource := &specStubResource{
		name: "resourceName",
	}
	submitTelemetryMetric(clientOpenAPI, TelemetryResourceOperationCreate, specResource, "prefix_")
	assert.Equal(t, "prefix_resourceName", resourceNameReceived)
	assert.Equal(t, TelemetryResourceOperationCreate, tfOperationReceived)
}

func TestSubmitTelemetryMetric_EmptyResourceName(t *testing.T) {
	var submitResourceExecutionMetricsFuncCalled bool
	clientOpenAPI := &clientOpenAPIStub{
		telemetryHandler: &telemetryHandlerStub{
			submitResourceExecutionMetricsFunc: func(resourceName string, tfOperation TelemetryResourceOperation) {
				submitResourceExecutionMetricsFuncCalled = true
			},
		},
	}
	specResource := &specStubResource{
		name: "",
	}
	submitTelemetryMetric(clientOpenAPI, TelemetryResourceOperationCreate, specResource, "prefix_")
	assert.False(t, submitResourceExecutionMetricsFuncCalled)
}
