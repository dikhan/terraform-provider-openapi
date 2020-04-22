package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/dikhan/terraform-provider-openapi/openapi/version"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// TelemetryProviderHTTPEndpoint defines the configuration for HTTPEndpoint. This struct also implements the TelemetryProvider interface
// and ships metrics to the following namespace by default <prefix>.terraform.* where '<prefix>' can be configured.
type TelemetryProviderHTTPEndpoint struct {
	// URL describes the HTTP endpoint to send the metric to
	URL string `yaml:"url"`
	// Prefix enables to append a prefix to the metrics pushed to the HTTP endpoint
	Prefix string `yaml:"prefix,omitempty"`
	// ProviderSchemaProperties defines what specific provider configuration properties and their values that will be injected into
	// metric API request headers. Values must match a real property name in provider schema configuration.
	ProviderSchemaProperties []string `yaml:"provider_schema_properties,omitempty"`
}

type TelemetryProviderConfigurationHTTPEndpoint struct {
	Headers map[string]string
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
func (g TelemetryProviderHTTPEndpoint) Validate() error {
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
func (g TelemetryProviderHTTPEndpoint) IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	version := strings.Replace(openAPIPluginVersion, ".", "_", -1)
	metricName := fmt.Sprintf("terraform.openapi_plugin_version.%s.total_runs", version)
	metric := createNewCounterMetric(g.Prefix, metricName)
	if err := g.submitMetric(metric, telemetryProviderConfiguration); err != nil {
		return err
	}
	return nil
}

// IncServiceProviderTotalRunsCounter will submit an increment to 1 the metric type counter '<prefix>.terraform.providers.%s.total_runs'. The
// %s will be replaced by the provider name used at runtime
func (g TelemetryProviderHTTPEndpoint) IncServiceProviderTotalRunsCounter(providerName string, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	metricName := fmt.Sprintf("terraform.providers.%s.total_runs", providerName)
	metric := createNewCounterMetric(g.Prefix, metricName)
	if err := g.submitMetric(metric, telemetryProviderConfiguration); err != nil {
		return err
	}
	return nil
}

func (g TelemetryProviderHTTPEndpoint) GetTelemetryProviderConfiguration(data *schema.ResourceData) TelemetryProviderConfiguration {
	tpConfig := TelemetryProviderConfigurationHTTPEndpoint{
		Headers: map[string]string{},
	}
	for _, propSchemaName := range g.ProviderSchemaProperties {
		propSchemaValue := data.Get(propSchemaName)
		if propSchemaValue != nil {
			tpConfig.Headers[propSchemaName] = propSchemaValue.(string)
		}
	}
	return tpConfig
}

func (g TelemetryProviderHTTPEndpoint) submitMetric(metric telemetryMetric, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	var telemetryConfiguration TelemetryProviderConfigurationHTTPEndpoint
	if telemetryProviderConfiguration != nil {
		var ok bool
		telemetryConfiguration, ok = telemetryProviderConfiguration.(TelemetryProviderConfigurationHTTPEndpoint)
		if !ok {
			return fmt.Errorf("wrong TelemetryProviderConfiguration object")
		}
	}

	log.Printf("[INFO] http endpoint metric to be submitted: %s", metric.MetricName)
	req, err := g.createNewRequest(metric, &telemetryConfiguration)
	if err != nil {
		return err
	}
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("request POST %s failed. Response Error: '%s'", g.URL, err.Error())
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("response returned from POST '%s' returned a non expected status code %d", g.URL, resp.StatusCode)
	}
	log.Printf("[INFO] http endpoint metric successfully submitted: %s", metric)
	return nil
}

func (g TelemetryProviderHTTPEndpoint) createNewRequest(metric telemetryMetric, telemetryProviderConfiguration *TelemetryProviderConfigurationHTTPEndpoint) (*http.Request, error) {
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
	if telemetryProviderConfiguration != nil && telemetryProviderConfiguration.Headers != nil {
		for schemaPropertyName, schemaPropertyValue := range telemetryProviderConfiguration.Headers {
			req.Header.Set(schemaPropertyName, schemaPropertyValue)
		}
	}
	return req, nil
}
