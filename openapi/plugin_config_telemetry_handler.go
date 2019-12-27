package openapi

import (
	"log"
	"time"
)

type TelemetryHandler interface {
	SubmitMetrics()
}

type telemetryHandlerTimeoutSupport struct {
	timeout            int
	providerName       string
	openAPIVersion     string
	terraformVersion   string
	telemetryProviders []TelemetryProvider
}

type MetricSubmitter func() error

func (t telemetryHandlerTimeoutSupport) SubmitMetrics() {
	for _, telemetryProvider := range t.telemetryProviders {
		t.submitMetric("IncServiceProviderTotalRunsCounter", func() error {
			return telemetryProvider.IncServiceProviderTotalRunsCounter(t.providerName)
		})
		t.submitMetric("IncOpenAPIPluginVersionTotalRunsCounter", func() error {
			return telemetryProvider.IncOpenAPIPluginVersionTotalRunsCounter(t.openAPIVersion)
		})
		t.submitMetric("IncTerraformVersionTotalRunsCounter", func() error {
			return telemetryProvider.IncTerraformVersionTotalRunsCounter(t.terraformVersion)
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
