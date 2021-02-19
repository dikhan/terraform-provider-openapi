package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// telemetryProviderConfigurationHTTPEndpoint defines the specific telemetry configuration for the  HTTPEndpoint telemetry provider. This
// struct is populated inside the GetTelemetryProviderConfiguration method given the resource data received.
type telemetryProviderConfigurationHTTPEndpoint struct {
	Headers map[string]string
}

type metricType string

const (
	metricTypeCounter metricType = "IncCounter"
)

type telemetryMetric struct {
	MetricType metricType `json:"metric_type"`
	MetricName string     `json:"metric_name"`
	Tags       []string   `json:"tags"`
}

func createNewCounterMetric(prefix, metricName string, tags []string) telemetryMetric {
	if prefix != "" {
		metricName = fmt.Sprintf("%s.%s", prefix, metricName)
	}
	return telemetryMetric{MetricType: metricTypeCounter, MetricName: metricName, Tags: tags}
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

// IncOpenAPIPluginVersionTotalRunsCounter will submit an increment to 1 the metric type counter '<prefix>.terraform.openapi_plugin_version.total_runs' including
// any other tag present in the TelemetryProviderConfiguration.
func (g TelemetryProviderHTTPEndpoint) IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	version := strings.Replace(openAPIPluginVersion, ".", "_", -1)
	tags := []string{"openapi_plugin_version:" + version}
	metricName := "terraform.openapi_plugin_version.total_runs"
	metric := createNewCounterMetric(g.Prefix, metricName, tags)
	if err := g.submitMetric(metric, telemetryProviderConfiguration); err != nil {
		return err
	}
	return nil
}

// IncServiceProviderResourceTotalRunsCounter will submit an increment to 1 the metric type counter '<prefix>.terraform.provider'.
// In addition, it will send tags with the provider name, resource name, and terrraform operation called.
func (g TelemetryProviderHTTPEndpoint) IncServiceProviderResourceTotalRunsCounter(providerName, resourceName string, tfOperation TelemetryResourceOperation, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	tags := []string{"provider_name:" + providerName, "resource_name:" + resourceName, fmt.Sprintf("terraform_operation:%s", tfOperation)}
	metricName := "terraform.provider"
	metric := createNewCounterMetric(g.Prefix, metricName, tags)
	if err := g.submitMetric(metric, telemetryProviderConfiguration); err != nil {
		return err
	}
	return nil
}

// GetTelemetryProviderConfiguration returns a telemetryProviderConfigurationHTTPEndpoint loaded with headers mapping to
// the plugin configuration schema properties that match the ones specified in the TelemetryProviderHTTPEndpoint ProviderSchemaProperties values
func (g TelemetryProviderHTTPEndpoint) GetTelemetryProviderConfiguration(data *schema.ResourceData) TelemetryProviderConfiguration {
	tpConfig := telemetryProviderConfigurationHTTPEndpoint{
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
	var telemetryConfiguration telemetryProviderConfigurationHTTPEndpoint
	if telemetryProviderConfiguration != nil {
		var ok bool
		telemetryConfiguration, ok = telemetryProviderConfiguration.(telemetryProviderConfigurationHTTPEndpoint)
		if !ok {
			return fmt.Errorf("telemetryProviderConfiguration object not the expected one: telemetryProviderConfigurationHTTPEndpoint")
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

func (g TelemetryProviderHTTPEndpoint) createNewRequest(metric telemetryMetric, telemetryProviderConfiguration *telemetryProviderConfigurationHTTPEndpoint) (*http.Request, error) {
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
