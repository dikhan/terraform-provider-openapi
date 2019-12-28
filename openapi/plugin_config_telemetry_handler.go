package openapi

import (
	"log"
	"time"
)

// TelemetryHandler is responsible for making sure that metrics are shipped to all the telemetry providers registered and
// also ensures that metrics submissions are configured with timeouts. Hence, if the telemetry provider is taking longer than the
// timeout set or it errors when sending the metric, the provider execution will not be affected by it and the corresponding error
// will be logged for the reference
type TelemetryHandler interface {
	// SubmitMetrics
	SubmitMetrics()
}

const telemetryTimeout = 2

type telemetryHandlerTimeoutSupport struct {
	timeout            int
	providerName       string
	openAPIVersion     string
	telemetryProviders []TelemetryProvider
}

// MetricSubmitter is the function holding the logic that actually submits the metric
type MetricSubmitter func() error

func (t telemetryHandlerTimeoutSupport) SubmitMetrics() {
	for _, telemetryProvider := range t.telemetryProviders {
		t.submitMetric("IncServiceProviderTotalRunsCounter", func() error {
			return telemetryProvider.IncServiceProviderTotalRunsCounter(t.providerName)
		})
		t.submitMetric("IncOpenAPIPluginVersionTotalRunsCounter", func() error {
			return telemetryProvider.IncOpenAPIPluginVersionTotalRunsCounter(t.openAPIVersion)
		})
	}
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
