package openapi

import (
	"errors"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"strings"
)

// TelemetryProviderGraphite defines the configuration for Graphite. This struct also implements the TelemetryProvider interface
// and ships metrics to the following namespace by default statsd.<prefix>.terraform.* where '<prefix>' can be configured.
type TelemetryProviderGraphite struct {
	// Host describes the graphite host (fqdn)
	Host string `yaml:"host"`
	// Port describes the port to where metrics will be pushed in Graphite
	Port int `yaml:"port"`
	// Prefix enables to append a prefix to the metrics pushed to graphite
	Prefix string `yaml:"prefix,omitempty"`
}

// Validate checks whether the provider is configured correctly. This validation is performed upon telemetry provider registration. If this
// method returns an error the error will be logged but the telemetry will be disabled. Otherwise, the telemetry will be enabled
// and the corresponding metrics will be shipped to Graphite
func (g TelemetryProviderGraphite) Validate() error {
	if g.Host == "" {
		return errors.New("graphite telemetry configuration is missing a value for the 'host property'")
	}
	if g.Port <= 0 {
		return errors.New("graphite telemetry configuration is missing a valid value (>0) for the 'port' property'")
	}
	return nil
}

// IncOpenAPIPluginVersionTotalRunsCounter will increment the counter 'statsd.<prefix>.terraform.openapi_plugin_version.%s.total_runs' metric to 1. The
// %s will be replaced by the OpenAPI plugin version used at runtime
func (g TelemetryProviderGraphite) IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string) error {
	version := strings.Replace(openAPIPluginVersion, ".", "_", -1)
	metric := fmt.Sprintf("terraform.openapi_plugin_version.%s.total_runs", version)
	log.Printf("[INFO] graphite metric to be submitted: %s", metric)
	if err := g.submitMetric(metric); err != nil {
		return err
	}
	log.Printf("[INFO] graphite metric successfully submitted: %s", metric)
	return nil
}

// IncServiceProviderTotalRunsCounter will increment the counter for a given provider 'statsd.<prefix>.terraform.providers.%s.total_runs' metric to 1. The
// %s will be replaced by the provider name used at runtime
func (g TelemetryProviderGraphite) IncServiceProviderTotalRunsCounter(providerName string) error {
	metric := fmt.Sprintf("terraform.providers.%s.total_runs", providerName)
	log.Printf("[INFO] graphite metric to be submitted: %s", metric)
	if err := g.submitMetric(metric); err != nil {
		return err
	}
	log.Printf("[INFO] graphite metric successfully submitted: %s", metric)
	return nil
}

func (g TelemetryProviderGraphite) GetTelemetryProviderConfiguration(data *schema.ResourceData) TelemetryProviderConfiguration {
	return nil
}

func (g TelemetryProviderGraphite) submitMetric(name string) error {
	c, err := g.getGraphiteClient()
	if err != nil {
		return err
	}
	nameWithPrefix := g.buildMetricName(name)
	return c.Incr(nameWithPrefix, nil, 1.0)
}

func (g TelemetryProviderGraphite) buildMetricName(name string) string {
	if g.Prefix != "" {
		return fmt.Sprintf("%s.%s", g.Prefix, name)
	}
	return name
}

func (g TelemetryProviderGraphite) getGraphiteClient() (*statsd.Client, error) {
	client, err := statsd.New(fmt.Sprintf("%s:%d", g.Host, g.Port))
	if err != nil {
		return nil, err
	}
	return client, nil
}
