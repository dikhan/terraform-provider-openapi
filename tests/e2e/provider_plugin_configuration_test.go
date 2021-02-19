package e2e

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestAcc_ProviderConfiguration_PluginExternalFile_HTTPEndpointTelemetry confirms regressions introduced in the logic related to the plugin
// external configuration. This test confirms that the plugin is able to start up properly and functions as expected even
// when the plugin uses the external configuration containing:
// - HTTPEndpoint telemetry configuration
// - Service configurations
func TestAcc_ProviderConfiguration_PluginExternalFile_HTTPEndpointTelemetry(t *testing.T) {
	httpEndpointTelemetryCalled := false
	var headersReceived http.Header
	var metricsReceived []string
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/metrics":
			httpEndpointTelemetryCalled = true
			headersReceived = r.Header
			body, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)
			metricsReceived = append(metricsReceived, string(body))
			w.WriteHeader(http.StatusOK)
			break
		case "/v1/cdns", "/v1/cdns/someID":
			httpEndpointTelemetryCalled = true
			if r.Method == http.MethodGet && r.URL.Path == "/v1/cdns" { // When the data source (with filters support) is calling the GET endpoint
				w.Write([]byte(`[{"id":"someID", "label": "some_label"}]`))
				break
			}
			w.Write([]byte(`{"id":"someID", "label": "some_label"}`))
			w.WriteHeader(http.StatusOK)
			break
		}
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: "%s"

schemes:
- "http"

paths:
  /v1/cdns:
    get:
     responses:
       200:
         schema:
           $ref: "#/definitions/ContentDeliveryNetworkV1Collection"
    post:
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      - type: string
        x-terraform-header: some_header
        name: some_header
        in: header
        required: true
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"

  /v1/cdns/{cdn_id}:
    get:
      parameters:
      - name: "cdn_id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
    delete:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn that needs to be deleted"
        required: true
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
  ContentDeliveryNetworkV1Collection:
    type: array
    items:
      $ref: "#/definitions/ContentDeliveryNetworkV1"
securityDefinitions:
  some_token:
    in: header
    name: Token
    type: apiKey`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    telemetry:
      http_endpoint:
        url: http://%s/v1/metrics
        provider_schema_properties: ["some_token", "some_header"]
    swagger-url: %s
    insecure_skip_verify: true`, apiHost, swaggerServer.URL)

	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())

	otfVarPluginConfigEnvVariableName := fmt.Sprintf("OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE", providerName)
	os.Setenv(otfVarPluginConfigEnvVariableName, file.Name())

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProvider()
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "openapi" {
  some_token = "token_value"
  some_header = "header_value" 
}
resource "openapi_cdns_v1" "my_cdn" { 
   label = "some_label"
}

data "openapi_cdns_v1" "my_data_cdn" { 
  filter {
	name = "id"
	values = ["someID"]
  }
}

data "openapi_cdns_v1_instance" "my_cdn" {
  id = "someID"
}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openapi_cdns_v1.my_cdn", "label", "some_label"),
					func(s *terraform.State) error { // asserting that the httpendpoint server received the expected metrics counter
						if !httpEndpointTelemetryCalled {
							return fmt.Errorf("http endpoint telemetry not called")
						}
						if headersReceived.Get("some_token") != "token_value" {
							return fmt.Errorf("expected header `some_token` in the metric API not received or not expected value received: %s", headersReceived.Get("some_token"))
						}
						if headersReceived.Get("some_header") != "header_value" {
							return fmt.Errorf("expected header `some_header` in the metric API not received or not expected value received: %s", headersReceived.Get("some_header"))
						}
						expectedPluginVersionMetric := `{"metric_type":"IncCounter","metric_name":"terraform.openapi_plugin_version.total_runs","tags":["openapi_plugin_version:dev"]}`
						if err := assertMetricExists(expectedPluginVersionMetric, metricsReceived); err != nil {
							return err
						}
						expectedDataSourceInstanceMetric := `{"metric_type":"IncCounter","metric_name":"terraform.provider","tags":["provider_name:openapi","resource_name:data_cdns_v1_instance","terraform_operation:read"]}`
						if err := assertMetricExists(expectedDataSourceInstanceMetric, metricsReceived); err != nil {
							return err
						}
						expectedResourceMetrics := `{"metric_type":"IncCounter","metric_name":"terraform.provider","tags":["provider_name:openapi","resource_name:cdns_v1","terraform_operation:create"]}`
						if err := assertMetricExists(expectedResourceMetrics, metricsReceived); err != nil {
							return err
						}
						return nil
					},
				),
			},
		},
	})
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func assertMetricExists(expectedMetric string, metrics []string) error {
	for _, metric := range metrics {
		if metric == expectedMetric {
			return nil
		}
	}
	return fmt.Errorf("metrics received [%s] don't contain the expected one [%s]", metrics, expectedMetric)
}

// TestAcc_ProviderConfiguration_PluginExternalFile_GraphiteTelemetry confirms regressions introduced in the logic related to the plugin
// external configuration. This test confirms that the plugin is able to start up properly and functions as expected even
// when the plugin uses the external configuration containing:
// - Graphite telemetry configuration
// - Service configurations
func TestAcc_ProviderConfiguration_PluginExternalFile_GraphiteTelemetry(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"someID", "label": "some_label"}`))
		w.WriteHeader(http.StatusOK)
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: "%s"

schemes:
- "http"

paths:
  /v1/cdns:
    post:
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"

  /v1/cdns/{cdn_id}:
    get:
      parameters:
      - name: "cdn_id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
    delete:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn that needs to be deleted"
        required: true
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	metricChannel := make(chan string)
	pc, telemetryHost, telemetryPort := graphiteServer(&metricChannel)
	defer pc.Close()

	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    telemetry:
      graphite:
        host: %s
        port: %s
    swagger-url: %s
    insecure_skip_verify: true`, telemetryHost, telemetryPort, swaggerServer.URL)

	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())

	otfVarPluginConfigEnvVariableName := fmt.Sprintf("OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE", providerName)
	os.Setenv(otfVarPluginConfigEnvVariableName, file.Name())

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProvider()
	assert.NoError(t, err)

	// Locking test to a fixed TF version to be able to assert the output more predictably since due to the channels the data received
	// is consumed sequentially, otherwise the TF testing framework will download the latest version of TF CLI which is not assured will
	// match the behaviour asserted below. Considering the assertions below are meant to test that the UDP server actually got the expected
	// metrics based on the graphite telemetry config; it's not important how Terraform handled the calls (which can vary depending on the
	// version and it's an implementation detail that should not affect the outcome of the test).
	// Note: Making a conscious decision here to lock this particular test to a specific version instead of having a more robust solution
	// like using Docker for the builds (configured with a locked go and terraform version) for the sake of more agile software development AND
	// also for documentation purposes so it's clear in the test itself the rational. This decision might be revisited if more
	// tests need a locked TF version; in which case other solutions might be preferable.
	os.Setenv("TF_ACC_TERRAFORM_VERSION", "0.13.5")
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "openapi_cdns_v1" "my_cdn" { label = "some_label"}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openapi_cdns_v1.my_cdn", "label", "some_label"),
					func(s *terraform.State) error { // asserting that the graphite server received the expected metrics counter
						assertExpectedMetric(t, &metricChannel, "terraform.openapi_plugin_version.total_runs:1|c|#openapi_plugin_version:dev") //Plan
						assertExpectedMetric(t, &metricChannel, "terraform.openapi_plugin_version.total_runs:1|c|#openapi_plugin_version:dev") //Apply
						assertExpectedMetric(t, &metricChannel, "terraform.provider:1|c|#provider_name:openapi,resource_name:cdns_v1,terraform_operation:create")
						return nil
					},
				),
			},
		},
	})
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
	os.Unsetenv("TF_ACC_TERRAFORM_VERSION")
}

func assertExpectedMetric(t *testing.T, metricChannel *chan string, expectedMetric string) {
	select {
	case metricReceived := <-*metricChannel:
		assert.Contains(t, metricReceived, expectedMetric)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("[FAIL] '%s' not received within the expected timeframe (timed out)", expectedMetric)
	}
}

func createPluginConfigFile(content string) *os.File {
	file, err := ioutil.TempFile("", "terraform-provider-openapi.yaml")
	if err != nil {
		log.Fatal(err)
	}
	file.Write([]byte(content))
	return file
}

func graphiteServer(metricChannel *chan string) (net.PacketConn, string, string) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}
	telemetryServer := pc.LocalAddr().String()
	telemetryHost := strings.Split(telemetryServer, ":")[0]
	telemetryPort := strings.Split(telemetryServer, ":")[1]
	go func() {
		for {
			buf := make([]byte, 2048)
			n, _, err := pc.ReadFrom(buf)
			if err != nil {
				continue
			}
			body := string(buf[:n])
			*metricChannel <- body
		}
	}()
	return pc, telemetryHost, telemetryPort
}
