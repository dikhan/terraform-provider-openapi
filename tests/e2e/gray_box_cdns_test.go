package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/dikhan/terraform-provider-openapi/v3/openapi"
	"github.com/stretchr/testify/assert"
)

const providerName = "openapi"

const resourceCDNName = "cdn_v1"

var openAPIResourceNameCDN = fmt.Sprintf("%s_%s", providerName, resourceCDNName)
var openAPIResourceInstanceNameCDN = "my_cdn"
var openAPIResourceStateCDN = fmt.Sprintf("%s.%s", openAPIResourceNameCDN, openAPIResourceInstanceNameCDN)
var openAPIDataSourceNameCDN = "my_data_source"
var openAPIDataSourceStateCDN = fmt.Sprintf("data.%s.%s", openAPIResourceNameCDN, openAPIDataSourceNameCDN)

const resourceCDNFirewallName = "cdn_v1_firewalls_v1"

var openAPIResourceNameCDNFirewall = fmt.Sprintf("%s_%s", providerName, resourceCDNFirewallName)
var openAPIResourceInstanceNameCDNFirewall = "my_cdn_firewall_v1"
var openAPIResourceStateCDNFirewall = fmt.Sprintf("%s.%s", openAPIResourceNameCDNFirewall, openAPIResourceInstanceNameCDNFirewall)

const cdnSwaggerYAMLTemplate = `swagger: "2.0"
host: %s 
schemes:
- "http"

paths:
  ######################
  #### CDN Resource ####
  ######################

  /v1/cdns:
    x-terraform-resource-name: "cdn"
    get:
      summary: "Get all cdns"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkCollectionV1"
    post:
      # x-terraform-resource-name: "cdn" (this extension has been deprecated and should be used on the root level path)
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV1"
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
      summary: "Get cdn by id"
      description: ""
      operationId: "ContentDeliveryNetworkGetV1"
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

    put:
      summary: "Updated cdn"
      operationId: "ContentDeliveryNetworkUpdateV1"
      parameters:
      - name: "id"
        in: "path"
        description: "cdn that needs to be updated"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Updated cdn object"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
    delete:
      summary: "Delete cdn"
      operationId: "ContentDeliveryNetworkDeleteV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn that needs to be deleted"
        required: true
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"

  ######################
  ## CDN sub-resource
  ######################

  /v1/cdns/{cdn_id}/v1/firewalls:
    post:
      summary: "Create cdn firewall"
      operationId: "ContentDeliveryNetworkFirewallCreateV1"
      parameters:
      - name: "cdn_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Created CDN firewall"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

  /v1/cdns/{cdn_id}/v1/firewalls/{id}:
    get:
      summary: "Get cdn firewall by id"
      description: ""
      operationId: "ContentDeliveryNetworkFirewallGetV1"
      parameters:
      - name: "cdn_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - name: "id"
        in: "path"
        description: "The cdn firewall id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
    delete: 
      operationId: ContentDeliveryNetworkFirewallDeleteV1
      parameters: 
        - description: "The cdn id that contains the firewall to be fetched."
          in: path
          name: parent_id
          required: true
          type: string
        - description: "The cdn firewall id that needs to be fetched."
          in: path
          name: id
          required: true
          type: string
      responses: 
        204: 
          description: "successful operation, no content is returned"
      summary: "Delete firewall"
    put:
      summary: "Updated firewall"
      operationId: "ContentDeliveryNetworkFirewallUpdatedV1"
      parameters:
      - name: "id"
        in: "path"
        description: "firewall that needs to be updated"
        required: true
        type: "string"
      - name: "parent_id"
        in: "path"
        description: "cdn which this firewall belongs to"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Updated firewall object"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

definitions:
  ContentDeliveryNetworkCollectionV1:
    type: "array"
    items:
      $ref: "#/definitions/ContentDeliveryNetworkV1"
  ContentDeliveryNetworkFirewallV1:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
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
      computed_property:
        type: "string"
        readOnly: true
      object_property_block: # Due to lack of support in Terraform for TypeMap with Elem *Resource; the only option available at the moment is to treat objects as TypeList with Elem *Resource and MaxItems 1. This will handle both simple objects as well as objects with complex property types (mix of types and even nested objects) and configurations - eg: computed)
        type: "object"
        properties:
          account:
            type: string
          object_read_only_property:
            type: string
            readOnly: true`

const expectedCDNID = "42"
const expectedCDNFirewallID = "1337"

var expectedCDNLabel = fmt.Sprintf("CDN #%s", expectedCDNID)
var expectedCDNFirewallLabel = fmt.Sprintf("FW #%s", expectedCDNFirewallID)

type api struct {
	swaggerURL string
	apiHost    string
	// cachePayloads holds the info posted to the different APIs. If a post has been called then the corresponding
	// payload response will be cached here so subsequent GET requests will return the same response mimicking the
	// same behaviour expected form a real API
	cachePayloads    map[string]interface{}
	requestsReceived []*http.Request
}

func initAPI(t *testing.T, swaggerYAMLTemplate string) *api {
	a := &api{
		cachePayloads: map[string]interface{}{},
	}
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.handleRequest(t, w, r)
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swaggerYAMLTemplate, apiHost)
		w.Write([]byte(swaggerReturned))
	}))
	a.swaggerURL = swaggerServer.URL
	a.apiHost = apiHost
	return a
}

func (a *api) handleRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	a.requestsReceived = append(a.requestsReceived, r)
	var cdnEndpoint = regexp.MustCompile(`^/v1/cdns`)
	var firewallEndpoint = regexp.MustCompile(`^/v1/cdns/[\d]*/v1/firewalls`)
	switch {
	case firewallEndpoint.MatchString(r.RequestURI):
		a.handleCDNFirewallRequest(t)[r.Method](w, r)
	case cdnEndpoint.MatchString(r.RequestURI):
		a.handleCDNRequest(t)[r.Method](w, r)
	}
}

func (a *api) handleCDNRequest(t *testing.T) map[string]http.HandlerFunc {
	apiServerBehaviors := map[string]http.HandlerFunc{}
	expectedRequestInstanceURI := fmt.Sprintf("/v1/cdns/%s", expectedCDNID)
	responseBody := fmt.Sprintf(
		`{
"id":%s,
"label":"%s",
"computed_property": "some auto-generated value",
"object_property_block": {"account":"my_account", "object_read_only_property": "some computed value for object read only"}
}`, expectedCDNID, expectedCDNLabel)

	apiServerBehaviors[http.MethodPost] = func(w http.ResponseWriter, r *http.Request) {
		assertExpectedRequestURI(t, "/v1/cdns", r)
		a.apiPostResponse(t, expectedRequestInstanceURI, responseBody, w, r)
	}
	apiServerBehaviors[http.MethodGet] = func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/v1/cdns" {
			expectedRequestInstanceURI = "/v1/cdns"
			assertExpectedRequestURI(t, expectedRequestInstanceURI, r)
			a.apiListResponse(t, w, r)
		} else {
			assertExpectedRequestURI(t, expectedRequestInstanceURI, r)
			a.apiGetResponse(t, w, r)
		}
	}
	apiServerBehaviors[http.MethodDelete] = func(w http.ResponseWriter, r *http.Request) {
		assertExpectedRequestURI(t, expectedRequestInstanceURI, r)
		a.apiDeleteResponse(t, w, r)
	}
	apiServerBehaviors[http.MethodPut] = func(w http.ResponseWriter, r *http.Request) {
		assertExpectedRequestURI(t, expectedRequestInstanceURI, r)
		a.apiPutResponse(t, w, r)
	}
	return apiServerBehaviors
}

func (a *api) handleCDNFirewallRequest(t *testing.T) map[string]http.HandlerFunc {
	apiServerBehaviors := map[string]http.HandlerFunc{}
	expectedRequestInstanceURI := fmt.Sprintf("/v1/cdns/%s/v1/firewalls/%s", expectedCDNID, expectedCDNFirewallID)
	responseBody := fmt.Sprintf(`{"id":%s,"label":"%s"}`, expectedCDNFirewallID, expectedCDNFirewallLabel)
	apiServerBehaviors[http.MethodPost] = func(w http.ResponseWriter, r *http.Request) {
		assertExpectedRequestURI(t, fmt.Sprintf("/v1/cdns/%s/v1/firewalls", expectedCDNID), r)
		a.apiPostResponse(t, expectedRequestInstanceURI, responseBody, w, r)
	}
	apiServerBehaviors[http.MethodGet] = func(w http.ResponseWriter, r *http.Request) {
		assertExpectedRequestURI(t, expectedRequestInstanceURI, r)
		a.apiGetResponse(t, w, r)
	}
	apiServerBehaviors[http.MethodDelete] = func(w http.ResponseWriter, r *http.Request) {
		assertExpectedRequestURI(t, expectedRequestInstanceURI, r)
		a.apiDeleteResponse(t, w, r)
	}
	apiServerBehaviors[http.MethodPut] = func(w http.ResponseWriter, r *http.Request) {
		assertExpectedRequestURI(t, expectedRequestInstanceURI, r)
		a.apiPutResponse(t, w, r)
	}
	return apiServerBehaviors
}

func (a *api) apiPostResponse(t *testing.T, cacheID string, responseBody string, w http.ResponseWriter, r *http.Request) {
	a.cachePayloads[cacheID] = responseBody
	a.apiResponse(t, responseBody, http.StatusCreated, w, r)
}

func (a *api) apiGetResponse(t *testing.T, w http.ResponseWriter, r *http.Request) {
	cachedBody := a.cachePayloads[r.RequestURI]
	if cachedBody == nil {
		a.apiResponse(t, "", http.StatusNotFound, w, r)
		return
	}
	a.apiResponse(t, cachedBody.(string), http.StatusOK, w, r)
}

func (a *api) apiListResponse(t *testing.T, w http.ResponseWriter, r *http.Request) {
	cdnList := []map[string]interface{}{
		{
			"id":                    expectedCDNID,
			"label":                 expectedCDNLabel,
			"computed_property":     "some auto-generated value",
			"object_property_block": map[string]string{"account": "my_account", "object_read_only_property": "some computed value for object read only"},
		},
		{
			"id":                    "some other id",
			"label":                 "some other label",
			"computed_property":     "some auto-generated value",
			"object_property_block": map[string]string{"account": "my_account", "object_read_only_property": "some computed value for object read only"},
		},
	}
	response, err := json.Marshal(cdnList)
	assert.Nil(t, err)
	a.apiResponse(t, string(response), http.StatusOK, w, r)
}

func (a *api) apiDeleteResponse(t *testing.T, w http.ResponseWriter, r *http.Request) {
	cachedBody := a.cachePayloads[r.RequestURI]
	if cachedBody == nil {
		a.apiResponse(t, "", http.StatusNotFound, w, r)
		return
	}
	a.cachePayloads[r.RequestURI] = nil
	a.apiResponse(t, "", http.StatusNoContent, w, r)
}

func (a *api) apiPutResponse(t *testing.T, w http.ResponseWriter, r *http.Request) {
	cachedBody := a.cachePayloads[r.RequestURI]
	if cachedBody == nil {
		a.apiResponse(t, "", http.StatusNotFound, w, r)
		return
	}
	cachedBodyStr := cachedBody.(string)
	if strings.Contains(cachedBodyStr, `"id":42`) {
		a.cachePayloads[r.RequestURI] = `{"id":42, "label":"updatedCDNLabel"}`
		a.apiResponse(t, `{"label":"updatedCDNLabel"}`, http.StatusOK, w, r)
	} else if strings.Contains(cachedBodyStr, `"id":1337`) {
		a.cachePayloads[r.RequestURI] = `{"id":1337, "label":"updatedFWLabel"}`
		a.apiResponse(t, `{"label":"updatedFWLabel"}`, http.StatusOK, w, r)
	} else {
		assert.Fail(t, fmt.Sprintf("no PUT implementation in apiServer for %s", cachedBody))
	}

}

func (a *api) apiResponse(t *testing.T, responseBody string, httpResponseStatusCode int, w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		_, e := ioutil.ReadAll(r.Body)
		require.NoError(t, e)
		//fmt.Printf("%s request body >>> %s\n", r.Method, string(bs))
	}
	w.WriteHeader(httpResponseStatusCode)
	if responseBody != "" {
		w.Write([]byte(responseBody))
	}
	//fmt.Printf("%d response body >>> %s\n", httpResponseStatusCode, responseBody)
}

func TestAccCDN_CreateResourceWithIgnoreListOrderExtension(t *testing.T) {
	swagger := `swagger: "2.0"
host: %s 
schemes:
- "http"

paths:
  ######################
  #### CDN Resource ####
  ######################

  /v1/cdns:
    x-terraform-resource-name: "cdn"
    post:
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV1"
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
      summary: "Get cdn by id"
      description: ""
      operationId: "ContentDeliveryNetworkGetV1"
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
      summary: "Delete cdn"
      operationId: "ContentDeliveryNetworkDeleteV1"
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
      list_prop:
        type: "array"
        x-terraform-ignore-order: true
        items:
          type: "string"`
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body := `{"id": "someid", "label":"some label", "list_prop": ["value1", "value2", "value3"]}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	tfFileContents := `# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  label = "some label"
  list_prop = ["value3", "value1", "value2"]
}`

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: false,
				Config:             tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					// check resource
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", "some label"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.#", "3"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.0", "value3"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.1", "value1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.2", "value2"),
				),
			},
		},
	})
}

func TestAccCDN_Create_and_UpdateSubResource(t *testing.T) {
	api := initAPI(t, cdnSwaggerYAMLTemplate)
	tfFileContents := createTerraformFile(expectedCDNLabel, expectedCDNFirewallLabel)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: api.swaggerURL})
	assert.NoError(t, err)
	assertProviderSchema(t, provider)

	resourceInstancesToCheck := map[string]string{
		openAPIResourceNameCDNFirewall: fmt.Sprintf("%s/v1/cdns/%s/v1/firewalls", api.apiHost, expectedCDNID),
		openAPIResourceNameCDN:         fmt.Sprintf("%s/v1/cdns", api.apiHost),
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, api.swaggerURL) },
		CheckDestroy:      testAccCheckDestroy(resourceInstancesToCheck),
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWhetherResourceExist(resourceInstancesToCheck),

					// check resource
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", expectedCDNLabel),

					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "computed_property", "some auto-generated value"),

					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_block.#", "1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_block.0.account", "my_account"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_block.0.object_read_only_property", "some computed value for object read only"),

					resource.TestCheckResourceAttr(
						openAPIResourceStateCDNFirewall, "cdn_v1_id", expectedCDNID),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDNFirewall, "label", expectedCDNFirewallLabel),
				),
			},
			{
				Config: createTerraformFile("updatedCDNLabel", "updatedFWLabel"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWhetherResourceExist(resourceInstancesToCheck),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", "updatedCDNLabel"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "computed_property", "some auto-generated value"),

					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_block.#", "1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_block.0.account", "my_account"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_block.0.object_read_only_property", "some computed value for object read only"),

					resource.TestCheckResourceAttr(
						openAPIResourceStateCDNFirewall, "cdn_v1_id", expectedCDNID),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDNFirewall, "label", "updatedFWLabel"),
				),
			},
		},
	})
}

func TestAccCDN_POSTRequestSchemaContainsInputsAndResponseSchemaContainsOutputs(t *testing.T) {
	expectedID := "some_id"
	expectedLabel := "my_label"
	swagger := `swagger: "2.0"
host: %s 
schemes:
- "http"

paths:
  ######################
  #### CDN Resource ####
  ######################

  /v1/cdns:
    x-terraform-resource-name: "cdn"
    post:
      summary: "Create cdn"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkInput"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkOutput"
  /v1/cdns/{cdn_id}:
    get:
      summary: "Get cdn by id"
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
            $ref: "#/definitions/ContentDeliveryNetworkOutput"
    delete:
      summary: "Delete cdn"
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
  ContentDeliveryNetworkInput:
    type: "object"
    required:
      - label
    properties:
      label:
        type: "string"
  ContentDeliveryNetworkOutput:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
        readOnly: true`
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responsePayload string
		switch r.Method {
		case http.MethodPost:
			body, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)
			bodyJSON := map[string]interface{}{}
			err = json.Unmarshal(body, &bodyJSON)
			assert.Nil(t, err)
			assert.Equal(t, expectedLabel, bodyJSON["label"])
			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		responsePayload = fmt.Sprintf(`{"id": "%s", "label":"%s"}`, expectedID, expectedLabel)
		w.Write([]byte(responsePayload))
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	tfFileContents := fmt.Sprintf(`# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  label = "%s"
}`, expectedLabel)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: false,
				Config:             tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					// check resource
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "id", expectedID),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", expectedLabel),
				),
			},
		},
	})
}

func TestAcc_Create_MissingRequiredParentPropertyInTFConfigurationFile(t *testing.T) {
	api := initAPI(t, cdnSwaggerYAMLTemplate)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: api.swaggerURL})
	assert.NoError(t, err)
	assertProviderSchema(t, provider)

	testCDNCreateMissingParentPropertyInFW := fmt.Sprintf(`
		# URI /v1/cdns/
		resource "%s" "%s" {
		  label = "%s"
		}
		# URI /v1/cdns/{parent_id}/v1/firewalls/
        resource "%s" "%s" {
           # cdn_v1_id = %s.id All parent properties must be specified in subresources
           label = "%s"
        }`, openAPIResourceNameCDN, openAPIResourceInstanceNameCDN, expectedCDNLabel, openAPIResourceNameCDNFirewall, openAPIResourceInstanceNameCDNFirewall, openAPIResourceStateCDN, expectedCDNFirewallLabel)

	expectedValidationError, _ := regexp.Compile(".*The argument \"cdn_v1_id\" is required, but no definition was found.*")
	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		ProviderFactories:         testAccProviders(provider),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config:      testCDNCreateMissingParentPropertyInFW,
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccCDN_ImportSubResource(t *testing.T) {
	api := initAPI(t, cdnSwaggerYAMLTemplate)

	api.cachePayloads["/v1/cdns/42/v1/firewalls/1337"] = `{"id":1337, "label":"importedFWLabel"}`

	tfFileContents := fmt.Sprintf(`
		# URI /v1/cdns/{parent_id}/v1/firewalls/
	   resource "%s" "%s" {
	   }`,
		openAPIResourceNameCDNFirewall,
		openAPIResourceInstanceNameCDNFirewall)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: api.swaggerURL})
	assert.NoError(t, err)
	assertProviderSchema(t, provider)

	resourceInstancesToCheck := map[string]string{
		openAPIResourceNameCDNFirewall: fmt.Sprintf("%s/v1/cdns/%s/v1/firewalls", api.apiHost, expectedCDNID),
		openAPIResourceNameCDN:         fmt.Sprintf("%s/v1/cdns", api.apiHost),
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, api.swaggerURL) },
		CheckDestroy:      testAccCheckDestroy(resourceInstancesToCheck),
		Steps: []resource.TestStep{
			{
				Config:        tfFileContents,
				ResourceName:  openAPIResourceStateCDNFirewall,
				ImportStateId: fmt.Sprintf("%s/%s", expectedCDNID, expectedCDNFirewallID),
				ImportState:   true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWhetherResourceExist(resourceInstancesToCheck),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", expectedCDNLabel),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDNFirewall, "cdn_v1_id", expectedCDNID),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDNFirewall, "label", expectedCDNFirewallLabel),
				),
			},
		},
	})
}

func TestAccCDN_DataSource(t *testing.T) {
	api := initAPI(t, cdnSwaggerYAMLTemplate)
	tfFileContents := fmt.Sprintf(`
		data "%s" "%s" {
		  filter {
		    name = "label"
		    values = ["%s"]
		  }
		}`, openAPIResourceNameCDN, openAPIDataSourceNameCDN, expectedCDNLabel)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: api.swaggerURL})
	assert.NoError(t, err)
	assertProviderSchema(t, provider)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, api.swaggerURL) },
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					// check data source
					resource.TestCheckResourceAttr(
						openAPIDataSourceStateCDN, "label", expectedCDNLabel),
					resource.TestCheckResourceAttr(
						openAPIDataSourceStateCDN, "id", expectedCDNID),
				),
			},
		},
	})
}

func TestAccCDN_DataSourceInstance(t *testing.T) {
	api := initAPI(t, cdnSwaggerYAMLTemplate)
	api.cachePayloads = map[string]interface{}{ // Pretending resource already exists remotely
		"/v1/cdns/" + expectedCDNID: fmt.Sprintf(
			`{
"id":%s,
"label":"%s",
"computed_property": "some auto-generated value",
"object_property_block": {"account":"my_account", "object_read_only_property": "some computed value for object read only"}
}`, expectedCDNID, expectedCDNLabel),
	}

	dataSourceInstanceName := fmt.Sprintf("%s_instance", openAPIResourceNameCDN)
	tfFileContents := fmt.Sprintf(`
		data "%s" "%s" {
		  id = "%s"
		}`, dataSourceInstanceName, openAPIDataSourceNameCDN, expectedCDNID)
	openAPIDataSourceInstanceStateCDN := fmt.Sprintf("data.%s.%s", dataSourceInstanceName, openAPIDataSourceNameCDN)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: api.swaggerURL})
	assert.NoError(t, err)
	assertProviderSchema(t, provider)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, api.swaggerURL) },
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						openAPIDataSourceInstanceStateCDN, "id", expectedCDNID),
					resource.TestCheckResourceAttr(
						openAPIDataSourceInstanceStateCDN, "label", expectedCDNLabel),
					resource.TestCheckResourceAttr(
						openAPIDataSourceInstanceStateCDN, "computed_property", "some auto-generated value"),
					resource.TestCheckResourceAttr(
						openAPIDataSourceInstanceStateCDN, "object_property_block.#", "1"),
					resource.TestCheckResourceAttr(
						openAPIDataSourceInstanceStateCDN, "object_property_block.0.account", "my_account"),
					resource.TestCheckResourceAttr(
						openAPIDataSourceInstanceStateCDN, "object_property_block.0.object_read_only_property", "some computed value for object read only"),
				),
			},
		},
	})
}

func assertProviderSchema(t *testing.T, provider *schema.Provider) {
	// Resource map check
	assert.Nil(t, provider.ResourcesMap[openAPIResourceNameCDN].Schema["id"])
	assert.NotNil(t, provider.ResourcesMap[openAPIResourceNameCDN].Schema["label"])
	assert.Nil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["id"])
	assert.NotNil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["label"])
	assert.NotNil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["cdn_v1_id"])
	assert.Nil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["cdns_v1_id"])

	// Data source map check
	assert.Nil(t, provider.DataSourcesMap[openAPIResourceNameCDN].Schema["id"])
	assert.NotNil(t, provider.DataSourcesMap[openAPIResourceNameCDN].Schema["label"])
	assert.NotNil(t, provider.DataSourcesMap[openAPIResourceNameCDN].Schema["computed_property"])
	assert.NotNil(t, provider.DataSourcesMap[openAPIResourceNameCDN].Schema["object_property_block"])

	openAPIDataSourceInstanceCDN := openAPIResourceNameCDN + "_instance"
	assert.NotNil(t, provider.DataSourcesMap[openAPIDataSourceInstanceCDN].Schema["id"]) // data source instance expects only one property from the user called 'id'. Hence, checking that is configured as expected
	assert.NotNil(t, provider.DataSourcesMap[openAPIDataSourceInstanceCDN].Schema["label"])
	assert.NotNil(t, provider.DataSourcesMap[openAPIDataSourceInstanceCDN].Schema["computed_property"])
	assert.NotNil(t, provider.DataSourcesMap[openAPIDataSourceInstanceCDN].Schema["object_property_block"])
}

func createTerraformFile(expectedCDNLabel, expectedFirewallLabel string) string {
	return fmt.Sprintf(`# URI /v1/cdns/
		resource "%s" "%s" {
		  label = "%s"
		  object_property_block {
		   account = "my_account"
		  }
		}
		# URI /v1/cdns/{parent_id}/v1/firewalls/
        resource "%s" "%s" {
           cdn_v1_id = %s.id
           label = "%s"
        }`, openAPIResourceNameCDN, openAPIResourceInstanceNameCDN, expectedCDNLabel, openAPIResourceNameCDNFirewall, openAPIResourceInstanceNameCDNFirewall, openAPIResourceStateCDN, expectedFirewallLabel)
}

func TestAccCDN_WriteOnlyProperties(t *testing.T) {
	swagger := `swagger: "2.0"
host: %s 
schemes:
- "http"

paths:
  ######################
  #### CDN Resource ####
  ######################

  /v1/cdns:
    x-terraform-resource-name: "cdn"
    post:
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV1"
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
      summary: "Get cdn by id"
      description: ""
      operationId: "ContentDeliveryNetworkGetV1"
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
    put:
      summary: "Updated cdn"
      operationId: "ContentDeliveryNetworkUpdateV1"
      parameters:
        - name: "id"
          in: "path"
          description: "cdn that needs to be updated"
          required: true
          type: "string"
        - in: "body"
          name: "body"
          description: "Updated cdn object"
          required: true
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
    delete:
      summary: "Delete cdn"
      operationId: "ContentDeliveryNetworkDeleteV1"
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
      - write_only_property
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
      write_only_property:
        type: "string"
        x-terraform-write-only: true
      list_prop:
        type: "array"
        x-terraform-write-only: true
        items:
          type: "string"
      object_write_only_prop:
        type: "object"
        x-terraform-write-only: true
        required:
          - nested_prop
        properties:
          nested_prop:
            type: "string"
      object_prop:
        type: "object"
        required:
          - nested_prop
        properties:
          nested_prop:
            type: "string"
            x-terraform-write-only: true`
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := `{"id": "someid", "label": "some label", "object_prop":{}}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: false,
				Config: `# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  label = "some label"
  write_only_property = "some property value"
  list_prop = ["value1", "value2"]
  object_write_only_prop {
    nested_prop = "some value"
  }
  object_prop {
    nested_prop = "some other value"
  }
}`,
			},
			{
				ExpectNonEmptyPlan: false,
				Config: `# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  label = "some label"
  write_only_property = "some property label"
  list_prop = ["value3", "value4"]
  object_write_only_prop {
    nested_prop = "some new value"
  }
  object_prop {
    nested_prop = "some other new value"
  }
}`,
			},
			{
				ExpectNonEmptyPlan: false,
				ImportStateVerify:  true,
				ImportStateId:      "someid",
				Config: `# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  label = "some label"
  write_only_property = "some property label"
  list_prop = ["value3", "value4"]
  object_write_only_prop {
    nested_prop = "some new value"
  }
  object_prop {
    nested_prop = "some other new value"
  }
}`,
			},
		},
	})
}
