package main

import (
	"bytes"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/printers"
	"log"
	"net/http"
	"net/http/httptest"
)

func main() {

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: my-api.com
schemes:
- "http"
security:
  - api_auth: []
securityDefinitions:
  api_auth:
    type: apiKey
    name: Authorization
    in: header
paths:
  /cdns:
    x-terraform-resource-name: cdn
    post:
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      - in: header
        type: string
        name: required_header_example
        required: true
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
  /cdns/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
        description: System generated identifier for the CDN
      label:
        type: "string"
        description: Label to use for the CDN`)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	var buf bytes.Buffer
	log.SetOutput(&buf)

	terraformDocGenerator := openapi.TerraformProviderDocGenerator{
		OpenAPIDocURL: swaggerServer.URL,
		Printer:       &printers.MarkdownPrinter{},
		ProviderName:  "cloudflare", // TODO: add support for extension x-terraform-docs-provider-name
	}

	err := terraformDocGenerator.GenerateDocumentation()
	if err != nil {
		log.Fatal(err)
	}
}
