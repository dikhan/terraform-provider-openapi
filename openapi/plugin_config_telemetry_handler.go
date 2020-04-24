package openapi

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"time"
)

// telemetryHandler is responsible for making sure that metrics are shipped to all the telemetry providers registered and
// also ensures that metrics submissions are configured with timeouts. Hence, if the telemetry provider is taking longer than the
// timeout set or it errors when sending the metric, the provider execution will not be affected by it and the corresponding error
// will be logged for the reference
type telemetryHandler interface {
	// submitPluginExecutionMetrics submits the metrics for the total number of times the plugin and specific OpenAPI plugin version
	// have been executed
	submitPluginExecutionMetrics()
	// submitResourceExecutionMetrics submits the metrics related to resource operation execution
	submitResourceExecutionMetrics(resourceName string, tfOperation TelemetryResourceOperation)
}

const telemetryTimeout = 2

type telemetryHandlerTimeoutSupport struct {
	timeout           int
	providerName      string
	openAPIVersion    string
	telemetryProvider TelemetryProvider
	data              *schema.ResourceData
}

// MetricSubmitter is the function holding the logic that actually submits the metric
type MetricSubmitter func() error

func (t telemetryHandlerTimeoutSupport) submitPluginExecutionMetrics() {
	if t.telemetryProvider == nil {
		log.Println("[INFO] Telemetry provider not configured")
		return
	}
	telemetryConfig := t.telemetryProvider.GetTelemetryProviderConfiguration(t.data)
	t.submitMetric("IncServiceProviderTotalRunsCounter", func() error {
		return t.telemetryProvider.IncServiceProviderTotalRunsCounter(t.providerName, telemetryConfig)
	})
	t.submitMetric("IncOpenAPIPluginVersionTotalRunsCounter", func() error {
		return t.telemetryProvider.IncOpenAPIPluginVersionTotalRunsCounter(t.openAPIVersion, telemetryConfig)
	})
}

func (t telemetryHandlerTimeoutSupport) submitResourceExecutionMetrics(resourceName string, tfOperation TelemetryResourceOperation) {
	if t.telemetryProvider == nil {
		log.Println("[INFO] Telemetry provider not configured")
		return
	}
	telemetryConfig := t.telemetryProvider.GetTelemetryProviderConfiguration(t.data)
	t.submitMetric("IncServiceProviderResourceTotalRunsCounter", func() error {
		return t.telemetryProvider.IncServiceProviderResourceTotalRunsCounter(t.providerName, resourceName, tfOperation, telemetryConfig)
	})
}

func (t telemetryHandlerTimeoutSupport) submitMetric(metricName string, metricSubmitter MetricSubmitter) {
	doneChan := make(chan error)
	go func() {
		doneChan <- metricSubmitter()
	}()
	// Wait till metric submission is completed or it times out
	select {
	case err := <-doneChan:
		if err != nil {
			log.Printf("metric '%s' submission failed: %s", metricName, err)
		}
	case <-time.After(time.Duration(t.timeout) * time.Second):
		log.Printf("metric '%s' submission did not finish within the expected time %ds", metricName, t.timeout)
	}
}
