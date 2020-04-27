package openapi

type telemetryHandlerStub struct {
	submitPluginExecutionMetricsFunc   func()
	submitResourceExecutionMetricsFunc func(resourceName string, tfOperation TelemetryResourceOperation)
}

func (t *telemetryHandlerStub) SubmitPluginExecutionMetrics() {
	t.submitPluginExecutionMetricsFunc()
}

func (t *telemetryHandlerStub) SubmitResourceExecutionMetrics(resourceName string, tfOperation TelemetryResourceOperation) {
	t.submitResourceExecutionMetricsFunc(resourceName, tfOperation)
}
