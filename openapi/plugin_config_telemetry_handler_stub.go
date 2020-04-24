package openapi

type telemetryHandlerStub struct {
	submitPluginExecutionMetricsFunc   func()
	submitResourceExecutionMetricsFunc func(resourceName string, tfOperation TelemetryResourceOperation)
}

func (t *telemetryHandlerStub) submitPluginExecutionMetrics() {
	t.submitPluginExecutionMetricsFunc()
}

func (t *telemetryHandlerStub) submitResourceExecutionMetrics(resourceName string, tfOperation TelemetryResourceOperation) {
	t.submitResourceExecutionMetricsFunc(resourceName, tfOperation)
}
