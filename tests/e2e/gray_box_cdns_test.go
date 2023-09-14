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
	swagger := getFileContents(t, "data/gray_box_test_data/ignore_order/openapi.yaml")

	resourceStateRemote := make([]byte, 0)
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			resourceStateRemote = make([]byte, 0)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)
			bodyJSON := map[string]interface{}{}
			err = json.Unmarshal(body, &bodyJSON)
			assert.Nil(t, err)
			bodyJSON["id"] = "someid"
			if len(resourceStateRemote) > 0 {
				assert.Fail(t, "POST request triggered more than once where the resource is only expected to be created once.")
			}
			resourceStateRemote, err = json.Marshal(bodyJSON)
			assert.Nil(t, err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resourceStateRemote)
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	tfFileContentsStage1 := getFileContents(t, "data/gray_box_test_data/ignore_order/test_stage_1.tf")
	tfFileContentsStage2 := getFileContents(t, "data/gray_box_test_data/ignore_order/test_stage_2.tf")

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		ProviderFactories:         testAccProviders(provider),
		PreCheck:                  func() { testAccPreCheck(t, swaggerServer.URL) },
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: tfFileContentsStage1,
				Check: resource.ComposeTestCheckFunc(
					// check resource
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", "some label"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.0", "value1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.2", "value3"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "integer_list_prop.0", "1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "integer_list_prop.2", "3"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "nested_list_prop.0.some_property", "value1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "nested_list_prop.2.some_property", "value3"),
				),
			},
			{
				Config: tfFileContentsStage2,
				Check: resource.ComposeTestCheckFunc(
					// check resource
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", "some label"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.0", "value1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.2", "value2"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "integer_list_prop.0", "1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "integer_list_prop.2", "2"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "nested_list_prop.0.some_property", "value1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "nested_list_prop.2.some_property", "value2"),
				),
			},
			{
				Config:             tfFileContentsStage2,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

func TestAcc_ArrayBasedComputedPropertiesUpdateCorrectly(t *testing.T) {
	swagger := getFileContents(t, "data/gray_box_test_data/updatable-computed-properties-test/openapi.yaml")
	expectedRequestBodiesRaw := getFileContentsBytes(t, "data/gray_box_test_data/updatable-computed-properties-test/expected_request_bodies.json")
	var expectedRequestBodiesJSON []map[string]interface{}
	err := json.Unmarshal(expectedRequestBodiesRaw, &expectedRequestBodiesJSON)
	assert.Nil(t, err)

	resourceStateRemote := make([]byte, 0)
	requestWithBodyIdx := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			resourceStateRemote = make([]byte, 0)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			body, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)

			expectedRequestBody, _ := json.Marshal(expectedRequestBodiesJSON[requestWithBodyIdx])
			assert.Equal(t, string(expectedRequestBody), string(body))
			requestWithBodyIdx++

			bodyJSON := map[string]interface{}{}
			err = json.Unmarshal(body, &bodyJSON)
			assert.Nil(t, err)
			bodyJSON["id"] = "someid"
			resourceStateRemote, err = json.Marshal(bodyJSON)
			assert.Nil(t, err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resourceStateRemote)
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	tfFileContentsStage1 := getFileContents(t, "data/gray_box_test_data/updatable-computed-properties-test/test_stage_1.tf")
	tfFileContentsStage2 := getFileContents(t, "data/gray_box_test_data/updatable-computed-properties-test/test_stage_2.tf")

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		ProviderFactories:         testAccProviders(provider),
		PreCheck:                  func() { testAccPreCheck(t, swaggerServer.URL) },
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: tfFileContentsStage1,
			},
			{
				Config: tfFileContentsStage2,
			},
		},
	})
}

func TestAcc_OrderIgnoredPlanIsStableWithReadOnlyProperties(t *testing.T) {
	swagger := getFileContents(t, "data/gray_box_test_data/ignore-order-plan-stability-test/openapi.yaml")

	testResponseJSONStr := getFileContents(t, "data/gray_box_test_data/ignore-order-plan-stability-test/test_response.json")

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testResponseJSONStr))
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	tfFileContentsStage1 := getFileContents(t, "data/gray_box_test_data/ignore-order-plan-stability-test/test_stage_1.tf")
	tfFileContentsStage2 := getFileContents(t, "data/gray_box_test_data/ignore-order-plan-stability-test/test_stage_2.tf")

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		ProviderFactories:         testAccProviders(provider),
		PreCheck:                  func() { testAccPreCheck(t, swaggerServer.URL) },
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: tfFileContentsStage1,
			},
			{
				Config: tfFileContentsStage2,
			},
		},
	})
}

func TestAcc_ErrorOnUpdateDoesNotUpdateState(t *testing.T) {
	swagger := getFileContents(t, "data/gray_box_test_data/update-error-no-state-change/openapi.yaml")

	testResponseJSONStr := getFileContents(t, "data/gray_box_test_data/update-error-no-state-change/test_response.json")

	returnFailureResponseOnPut := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && returnFailureResponseOnPut {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testResponseJSONStr))
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	tfFileContentsStage1 := getFileContents(t, "data/gray_box_test_data/update-error-no-state-change/test_stage_1.tf")
	tfFileContentsStage2 := getFileContents(t, "data/gray_box_test_data/update-error-no-state-change/test_stage_2.tf")

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	expectedPutError, _ := regexp.Compile("Error running apply.*")

	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		ProviderFactories:         testAccProviders(provider),
		PreCheck:                  func() { testAccPreCheck(t, swaggerServer.URL) },
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: tfFileContentsStage1,
			},
			{
				Config: tfFileContentsStage2,
				PreConfig: func() {
					returnFailureResponseOnPut = true
				},
				ExpectError: expectedPutError,
			},
			{
				Config: tfFileContentsStage2,
				PreConfig: func() {
					returnFailureResponseOnPut = false
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// terraform refresh must be able to empty resource in state when getting 404 Not Found response
// without this, terraform is unable to detect deleted resources and restore them
func TestAcc_NotFoundErrorOnResourceReadShouldRemoveResourceState(t *testing.T) {
	// re-use data files of other test as this test only requires a single stage
	swagger := getFileContents(t, "data/gray_box_test_data/update-error-no-state-change/openapi.yaml")
	testResponseJSONStr := getFileContents(t, "data/gray_box_test_data/update-error-no-state-change/test_response.json")
	tfConfig := getFileContents(t, "data/gray_box_test_data/update-error-no-state-change/test_stage_1.tf")

	// instructs returning 404 response for GET request
	returnResourceNotFound := false

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testResponseJSONStr))
			return
		}

		if returnResourceNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"404 Not Found"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testResponseJSONStr))
		}
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
		IsUnitTest:                true,
		ProviderFactories:         testAccProviders(provider),
		PreCheck:                  func() { testAccPreCheck(t, swaggerServer.URL) },
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// regular tf flow and verifies resource in state
				Config: tfConfig,
				Check:
				// check resource attributes
				resource.TestCheckResourceAttr(openAPIResourceStateCDN, "id", "someid"),
			},
			{
				Config: tfConfig,
				PreConfig: func() {
					returnResourceNotFound = true
				},
				// test tf plan only which does a state refresh
				// expect tf detects state change and produce plan to restore manually deleted resource
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check:
				// check resource absence in state
				resource.TestCheckNoResourceAttr(openAPIResourceStateCDN, "id"),
			},
		},
	})
}

// Optional properties should be saved to state on state update
func TestAcc_OptionalPropertiesReflectedOnStateUpdate(t *testing.T) {
	swagger := getFileContents(t, "data/gray_box_test_data/update-state-containing-optional-properties/openapi.yaml")

	// stage 1: 1 resource creation containing optional properties, including nested ones under a list property
	// stage 2: modify the optional property in tf config, extend the list property
	responseStage1 := getFileContents(t, "data/gray_box_test_data/update-state-containing-optional-properties/stage1_response.json")
	responseStage2 := getFileContents(t, "data/gray_box_test_data/update-state-containing-optional-properties/stage2_response.json")

	resourceUpdated := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPut {
			resourceUpdated = true
		}
		w.WriteHeader(http.StatusOK)
		if resourceUpdated {
			w.Write([]byte(responseStage2))
		} else {
			w.Write([]byte(responseStage1))
		}
	}))
	apiHost := apiServer.URL[7:]
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(swagger, apiHost)
		w.Write([]byte(swaggerReturned))
	}))

	tfFileContentsStage1 := getFileContents(t, "data/gray_box_test_data/update-state-containing-optional-properties/test_stage_1.tf")
	tfFileContentsStage2 := getFileContents(t, "data/gray_box_test_data/update-state-containing-optional-properties/test_stage_2.tf")

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: swaggerServer.URL})
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		ProviderFactories:         testAccProviders(provider),
		PreCheck:                  func() { testAccPreCheck(t, swaggerServer.URL) },
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: tfFileContentsStage1,
			},
			{
				Config: tfFileContentsStage2,
				Check: resource.ComposeTestCheckFunc(
					// check resource attributes
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "id", "id_value"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "main_optional_prop", "main_optional_value_modified"),
					// order of elements in list property must follow order in tf config
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.0.sub_readonly_prop", "sub_readonly_value_2"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.0.sub_optional_prop", "sub_optional_value_2"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.1.sub_readonly_prop", "sub_readonly_value_1"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "list_prop.1.sub_optional_prop", "sub_optional_value_1"),
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
      - writeOnlyProperty
      - objectWriteOnlyProp
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
      writeOnlyProperty:
        type: "string"
        x-terraform-write-only: true
      listProp:
        type: "array"
        x-terraform-write-only: true
        items:
          type: "string"
      objectWriteOnlyProp:
        type: "object"
        x-terraform-write-only: true
        required:
          - nestedProp
        properties:
          nestedProp:
            type: "string"
            x-terraform-write-only: true
          nestedOptionalProp:
            type: "string"
            x-terraform-write-only: true
      objectProp:
        type: "object"
        required:
          - nestedProp
        properties:
          nestedProp:
            type: "string"
            x-terraform-write-only: true`
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := `{"id": "someid", "label": "some label", "objectProp":{}, "objectWriteOnlyProp": null}`
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
    nested_optional_prop = "optional val"
  }
  object_prop {
    nested_prop = "some other value"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "label", "some label"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "list_prop.#", "2"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "list_prop.0", "value1"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "list_prop.1", "value2"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "write_only_property", "some property value"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "object_write_only_prop.0.nested_prop", "some value"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "object_write_only_prop.0.nested_optional_prop", "optional val"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "object_prop.0.nested_prop", "some other value"),
				),
			},
			{
				ExpectNonEmptyPlan: false,
				Config: `# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  label = "some label"
  write_only_property = "some new property"
  list_prop = ["value3", "value4"]
  object_write_only_prop {
    nested_prop = "some new value"
    nested_optional_prop = "optional new val"
  }
  object_prop {
    nested_prop = "some other new value"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "write_only_property", "some new property"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "object_write_only_prop.0.nested_prop", "some new value"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "object_write_only_prop.0.nested_optional_prop", "optional new val"),
					resource.TestCheckResourceAttr(openAPIResourceStateCDN, "object_prop.0.nested_prop", "some other new value"),
				),
			},
		},
	})
}
