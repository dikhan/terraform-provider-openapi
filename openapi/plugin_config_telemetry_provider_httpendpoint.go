package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/dikhan/terraform-provider-openapi/openapi/version"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// TelemetryProviderHttpEndpoint defines the configuration for HttpEndpoint. This struct also implements the TelemetryProvider interface
// and ships metrics to the following namespace by default <prefix>.terraform.* where '<prefix>' can be configured.
type TelemetryProviderHttpEndpoint struct {
	// URL describes the HTTP endpoint to send the metric to
	URL string `yaml:"url"`
	// Prefix enables to append a prefix to the metrics pushed to graphite
	Prefix string `yaml:"prefix,omitempty"`
	// HttpClient holds the http client used to submit the metrics to the API
	HttpClient http.Client
}

type metricType string

const (
	metricTypeCounter metricType = "IncCounter"
)

type telemetryMetric struct {
	MetricType metricType `json:"metric_type"`
	MetricName string     `json:"metric_name"`
}

func createNewCounterMetric(prefix, metricName string) telemetryMetric {
	if prefix != "" {
		metricName = fmt.Sprintf("%s.%s", prefix, metricName)
	}
	return telemetryMetric{MetricType: metricTypeCounter, MetricName: metricName}
}

// Validate checks whether the provider is configured correctly. This validation is performed upon telemetry provider registration. If this
// method returns an error the error will be logged but the telemetry will be disabled. Otherwise, the telemetry will be enabled
// and the corresponding metrics will be shipped to Graphite
func (g TelemetryProviderHttpEndpoint) Validate() error {
	if g.URL == "" {
		return errors.New("http endpoint telemetry configuration is missing a value for the 'url property'")
	}
	if !govalidator.IsURL(g.URL) {
		return fmt.Errorf("http endpoint telemetry configuration does not have a valid URL '%s'", g.URL)
	}
	return nil
}

// IncOpenAPIPluginVersionTotalRunsCounter will submit an increment to 1 the metric type counter '<prefix>.terraform.openapi_plugin_version.%s.total_runs'. The
// %s will be replaced by the OpenAPI plugin version used at runtime
func (g TelemetryProviderHttpEndpoint) IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string) error {
	version := strings.Replace(openAPIPluginVersion, ".", "_", -1)
	metricName := fmt.Sprintf("terraform.openapi_plugin_version.%s.total_runs", version)
	metric := createNewCounterMetric(g.Prefix, metricName)
	if err := g.submitMetric(metric); err != nil {
		return err
	}
	return nil
}

// IncServiceProviderTotalRunsCounter will submit an increment to 1 the metric type counter '<prefix>.terraform.providers.%s.total_runs'. The
// %s will be replaced by the provider name used at runtime
func (g TelemetryProviderHttpEndpoint) IncServiceProviderTotalRunsCounter(providerName string) error {
	metricName := fmt.Sprintf("terraform.providers.%s.total_runs", providerName)
	metric := createNewCounterMetric(g.Prefix, metricName)
	if err := g.submitMetric(metric); err != nil {
		return err
	}
	return nil
}

func (g TelemetryProviderHttpEndpoint) submitMetric(metric telemetryMetric) error {
	log.Printf("[INFO] http endpoint metric to be submitted: %s", metric.MetricName)
	req, err := g.createNewRequest(metric)
	if err != nil {
		return err
	}
	resp, err := g.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request POST %s failed. Response Error: '%s'", g.URL, err.Error())
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("response returned from POST '%s' returned a non expected status code %d", g.URL, resp.StatusCode)
	}
	log.Printf("[INFO] http endpoint metric successfully submitted: %s", metric)
	return nil
}

func (g TelemetryProviderHttpEndpoint) createNewRequest(metric telemetryMetric) (*http.Request, error) {
	var body []byte
	var err error
	body, err = json.Marshal(metric)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, g.URL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set(contentType, "application/json")
	req.Header.Set(userAgentHeader, version.BuildUserAgent(runtime.GOOS, runtime.GOARCH))
	return req, nil
}
