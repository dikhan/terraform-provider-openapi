package openapi

import (
	"errors"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"log"
	"strings"
)

type TelemetryProvider interface {
	// Validate performs a check to confirm that the telemetry configuration is valid
	Validate() error
	// IncOpenAPIPluginVersionTotalRunsCounter is the method responsible for submitting to the corresponding telemetry platform the counter increase for the OpenAPI plugin Version used
	IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string) error
	// IncServiceProviderTotalRunsCounter is the method responsible for submitting to the corresponding telemetry platform the counter increase for the service provider used
	IncServiceProviderTotalRunsCounter(providerName string) error
}

type TelemetryConfig struct {
	Graphite *TelemetryProviderGraphite `yaml:"graphite,omitempty"`
}

type TelemetryProviderGraphite struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Prefix string `yaml:"prefix,omitempty"`
}

func (g TelemetryProviderGraphite) Validate() error {
	if g.Host == "" {
		return errors.New("graphite telemetry configuration is missing a value for the 'host property'")
	}
	if g.Port <= 0 {
		return errors.New("graphite telemetry configuration is missing a valid value (>0) for the 'port' property'")
	}
	return nil
}

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

func (g TelemetryProviderGraphite) IncServiceProviderTotalRunsCounter(providerName string) error {
	metric := fmt.Sprintf("terraform.providers.%s.total_runs", providerName)
	log.Printf("[INFO] graphite metric to be submitted: %s", metric)
	if err := g.submitMetric(metric); err != nil {
		return err
	}
	log.Printf("[INFO] graphite metric successfully submitted: %s", metric)
	return nil
}

func (g TelemetryProviderGraphite) submitMetric(name string) error {
	c, err := g.getGraphiteClient()
	if err != nil {
		return err
	}
	nameWithPrefix := g.buildMetricName(name)
	err = c.Incr(nameWithPrefix, nil, 1.0)
	if err != nil {
		return err
	}
	return nil
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
