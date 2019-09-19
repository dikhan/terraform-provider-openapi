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

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/dikhan/terraform-provider-openapi/openapi"
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
    get:
      summary: "Get all cdns"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkCollectionV1"
    post:
      x-terraform-resource-name: "cdn"
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
      object_property_argument: # option 1 (handling objects with complex property types and configurations - eg: computed)
        type: "object"
        properties:
          account:
            type: string
          object_read_only_property:
            type: string
            readOnly: true
      object_property_block: # option 2 (handling objects with complex property types and configurations - eg: computed)
        type: "object"
        x-terraform-complex-object-legacy-config: true
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
"object_property_argument": {"account":"my_account", "object_read_only_property": "some computed value for object read only"},
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
			"id":                       expectedCDNID,
			"label":                    expectedCDNLabel,
			"computed_property":        "some auto-generated value",
			"object_property_argument": map[string]string{"account": "my_account", "object_read_only_property": "some computed value for object read only"},
			"object_property_block":    map[string]string{"account": "my_account", "object_read_only_property": "some computed value for object read only"},
		},
		{
			"id":                       "some other id",
			"label":                    "some other label",
			"computed_property":        "some auto-generated value",
			"object_property_argument": map[string]string{"account": "my_account", "object_read_only_property": "some computed value for object read only"},
			"object_property_block":    map[string]string{"account": "my_account", "object_read_only_property": "some computed value for object read only"},
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

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
	resource.Test(t, resource.TestCase{
		IsUnitTest:   true,
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t, api.swaggerURL) },
		CheckDestroy: testAccCheckDestroy(resourceInstancesToCheck),
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

					// option 1
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_argument.%", "2"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_argument.account", "my_account"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_argument.object_read_only_property", "some computed value for object read only"),

					// option 2 (chosen)
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

					// option 1
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_argument.%", "2"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_argument.account", "my_account"),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_property_argument.object_read_only_property", "some computed value for object read only"),

					// option 2
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

func TestAcc_Create_MissingRequiredParentPropertyInTFConfigurationFile(t *testing.T) {
	api := initAPI(t, cdnSwaggerYAMLTemplate)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: api.swaggerURL})
	assert.NoError(t, err)
	assertProviderSchema(t, provider)

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
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

	expectedValidationError, _ := regexp.Compile(".*config is invalid: Missing required argument: The argument \"cdn_v1_id\" is required, but no definition was found.*")
	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		Providers:                 testAccProviders,
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

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
	resource.Test(t, resource.TestCase{
		IsUnitTest:   true,
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t, api.swaggerURL) },
		CheckDestroy: testAccCheckDestroy(resourceInstancesToCheck),
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

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   func() { testAccPreCheck(t, api.swaggerURL) },
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

func assertProviderSchema(t *testing.T, provider *schema.Provider) {
	assert.Nil(t, provider.ResourcesMap[openAPIResourceNameCDN].Schema["id"])
	assert.NotNil(t, provider.ResourcesMap[openAPIResourceNameCDN].Schema["label"])
	assert.Nil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["id"])
	assert.NotNil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["label"])
	assert.NotNil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["cdn_v1_id"])
	assert.Nil(t, provider.ResourcesMap[openAPIResourceNameCDNFirewall].Schema["cdns_v1_id"])
}

func createTerraformFile(expectedCDNLabel, expectedFirewallLabel string) string {
	return fmt.Sprintf(`# URI /v1/cdns/
		resource "%s" "%s" {
		  label = "%s"
		  object_property_block {
		   account = "my_account"
		  }

		  object_property_argument = {
		   account = "my_account"
           object_read_only_property = "some computed value for object read only" // This is the workaround users will have to do in order to fix the diff issues with objects that contain readOnly properties
		  }

		}
		# URI /v1/cdns/{parent_id}/v1/firewalls/
        resource "%s" "%s" {
           cdn_v1_id = %s.id
           label = "%s"
        }`, openAPIResourceNameCDN, openAPIResourceInstanceNameCDN, expectedCDNLabel, openAPIResourceNameCDNFirewall, openAPIResourceInstanceNameCDNFirewall, openAPIResourceStateCDN, expectedFirewallLabel)
}
