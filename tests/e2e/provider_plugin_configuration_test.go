package e2e

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

// TestAcc_ProviderConfiguration_PluginExternalFile confirms regressions introduced in the logic related to the plugin
// external configuration. This test confirms that the plugin is able to start up properly and functions as expected even
// when the plugin uses the external configuration containing:
// - Telemetry configuration
// - Service configurations
func TestAcc_ProviderConfiguration_PluginExternalFile(t *testing.T) {
	httpEndpointTelemetryCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/metrics":
			httpEndpointTelemetryCalled = true
			w.WriteHeader(http.StatusOK)
			break
		}
		w.Write([]byte(`{"id":"someID", "label": "some_label"}`))
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
	pc, telemetryHost, telemetryPort := graphiteServer(metricChannel)
	defer pc.Close()

	testPluginConfig := fmt.Sprintf(`version: '1'
telemetry:
  graphite:
    host: %s
    port: %s
  http_endpoint:
    url: http://%s/v1/metrics
services:
  openapi:
    swagger-url: %s
    insecure_skip_verify: true`, telemetryHost, telemetryPort, apiHost, swaggerServer.URL)

	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())

	otfVarPluginConfigEnvVariableName := fmt.Sprintf("OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE", providerName)
	os.Setenv(otfVarPluginConfigEnvVariableName, file.Name())

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProvider()
	assert.NoError(t, err)

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "openapi_cdns_v1" "my_cdn" { label = "some_label"}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openapi_cdns_v1.my_cdn", "label", "some_label"),
					func(s *terraform.State) error { // asserting that the httpendpoint server received the expected metrics counter
						if !httpEndpointTelemetryCalled {
							return fmt.Errorf("http endpoint telemetry not called")
						}
						return nil
					},
					func(s *terraform.State) error { // asserting that the graphite server received the expected metrics counter
						assertExpectedMetric(t, metricChannel, "terraform.providers.openapi.total_runs:1|c")
						assertExpectedMetric(t, metricChannel, "terraform.openapi_plugin_version.dev.total_runs:1|c")
						return nil
					},
				),
			},
		},
	})
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func assertExpectedMetric(t *testing.T, metricChannel chan string, expectedMetric string) {
	select {
	case metricReceived := <-metricChannel:
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

func graphiteServer(metricChannel chan string) (net.PacketConn, string, string) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}
	telemetryServer := pc.LocalAddr().String()
	telemetryHost := strings.Split(telemetryServer, ":")[0]
	telemetryPort := strings.Split(telemetryServer, ":")[1]
	go func() {
		for {
			buf := make([]byte, 1024)
			n, _, err := pc.ReadFrom(buf)
			if err != nil {
				continue
			}
			body := string(buf[:n])
			metricChannel <- body
		}
	}()
	return pc, telemetryHost, telemetryPort
}
