package openapi

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testSwagger = `swagger: "2.0"
x-terraform-provider-multiregion-fqdn: "service.api.${region}.hostname.com"
x-terraform-provider-regions: "region1, region2, region3"
host: api.my-service.com 
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
  /v1/cdns/{id}:
    get:
      summary: "Get cdn by id"
      description: "description of cdn get operation"
      operationId: "ContentDeliveryNetworkGetV1"
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
  ContentDeliveryNetworkCollectionV1:
    type: "array"
    items:
      $ref: "#/definitions/ContentDeliveryNetworkV1"
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - required_prop
    properties:
      id:
        type: "string"
        readOnly: true
      required_prop:
        type: "string"
      list_prop:
        type: "array"
        items:
          type: "string"
      obj_prop:
        type: "object"
        properties:
          label:
            type: "string"`

// resources, data source, list endpoint so we get data source filters
func TestGenerateDocumentation(t *testing.T) {
	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(testSwagger))
		assert.Nil(t, err)
	}))
	defer swaggerServer.Close()

	providerName := "openapi"
	dg := TerraformProviderDocGenerator{
		ProviderName:  providerName,
		OpenAPIDocURL: swaggerServer.URL,
	}
	d, err := dg.GenerateDocumentation()
	assert.Nil(t, err)
	assert.Equal(t, providerName, d.ProviderName)

	// ProviderInstallation assertions
	expectedProviderInstallExample := fmt.Sprintf("$ export PROVIDER_NAME=%s && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME<br>"+
		"[INFO] Downloading https://github.com/dikhan/terraform-provider-openapi/releases/download/v0.29.4/terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz in temporally folder /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh...<br>"+
		"[INFO] Extracting terraform-provider-openapi from terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz...<br>"+
		"[INFO] Cleaning up tmp dir created for installation purposes: /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh<br>"+
		"[INFO] Terraform provider 'terraform-provider-%s' successfully installed at: '~/.terraform.d/plugins'!", providerName, providerName)
	assert.Equal(t, expectedProviderInstallExample, d.ProviderInstallation.Example)
	assert.Equal(t, "You can then start running the Terraform provider:", d.ProviderInstallation.Other)
	assert.Equal(t, fmt.Sprintf("$ export OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE=\"https://api.service.com/openapi.yaml\"<br>", providerName), d.ProviderInstallation.OtherCommand)

	// ProviderConfiguration assertions
	assert.Equal(t, []string{"region1", "region2", "region3"}, d.ProviderConfiguration.Regions)
	//assert.Equal(t, []Property{}, d.ProviderConfiguration.ConfigProperties)
	//assert.Equal(t, []ExampleUsage{}, d.ProviderConfiguration.ExampleUsage)
	//assert.Equal(t, ArgumentsReference{Notes: nil}, d.ProviderConfiguration.ArgumentsReference)

	// ProviderResources assertions
	assert.Len(t, d.ProviderResources.Resources, 1)
	cdnResource := d.ProviderResources.Resources[0]
	assert.Equal(t, "cdn_v1", cdnResource.Name)
	assert.Equal(t, "", cdnResource.Description)
	assert.Equal(t, ArgumentsReference{Notes: []string{}}, cdnResource.ArgumentsReference)
	cdnResourceProps := cdnResource.Properties
	assert.Len(t, cdnResourceProps, 4)
	assertProperty(t, cdnResourceProps[0], "required_prop", "string", "", "", true, false, nil)
	assertProperty(t, cdnResourceProps[1], "id", "string", "", "", false, true, nil)
	assertProperty(t, cdnResourceProps[2], "obj_prop", "object", "", "", false, false,
		[]Property{{Name: "label", Type: "string", ArrayItemsType: "", Required: false, Computed: false, Description: "", Schema: nil}})
	assertProperty(t, cdnResourceProps[3], "list_prop", "list", "string", "", false, false, nil)

	// DataSource assertions
	assert.Len(t, d.DataSources.DataSources, 1)
	cdnDataSource := d.DataSources.DataSources[0]
	assert.Equal(t, "cdn_v1", cdnDataSource.Name)
	assert.Equal(t, "", cdnDataSource.OtherExample)
	cdnDataSourceProps := cdnDataSource.Properties
	assert.Len(t, cdnResourceProps, 4)
	assertDataSourceProperty(t, cdnDataSourceProps[0], "list_prop", "list", "string", "", nil)
	assertDataSourceProperty(t, cdnDataSourceProps[1], "required_prop", "string", "", "", nil)
	assertDataSourceProperty(t, cdnDataSourceProps[2], "id", "string", "", "", nil)
	assertDataSourceProperty(t, cdnDataSourceProps[3], "obj_prop", "object", "", "",
		[]Property{{Name: "label", Type: "string", ArrayItemsType: "", Required: false, Computed: true, Description: "", Schema: nil}})

	// DataSourceInstance assertions
	assert.Len(t, d.DataSources.DataSourceInstances, 1)
	cdnDataSourceInstance := d.DataSources.DataSourceInstances[0]
	assert.Equal(t, "cdn_v1_instance", cdnDataSourceInstance.Name)
	assert.Equal(t, "", cdnDataSourceInstance.OtherExample)
	cdnDataSourceInstanceProps := cdnDataSourceInstance.Properties
	assert.Len(t, cdnDataSourceInstanceProps, 4)
	assertDataSourceProperty(t, cdnDataSourceInstanceProps[0], "list_prop", "list", "string", "", nil)
	assertDataSourceProperty(t, cdnDataSourceInstanceProps[1], "required_prop", "string", "", "", nil)
	assertDataSourceProperty(t, cdnDataSourceInstanceProps[2], "id", "string", "", "", nil)
	assertDataSourceProperty(t, cdnDataSourceInstanceProps[3], "obj_prop", "object", "", "",
		[]Property{{Name: "label", Type: "string", ArrayItemsType: "", Required: false, Computed: true, Description: "", Schema: nil}})
}

func assertDataSourceProperty(t *testing.T, actualProp Property, expectedName, expectedType, expectedArrayItemsType, expectedDescription string, expectedSchema []Property) {
	assertProperty(t, actualProp, expectedName, expectedType, expectedArrayItemsType, expectedDescription, false, true, expectedSchema)
}

func assertProperty(t *testing.T, actualProp Property, expectedName, expectedType, expectedArrayItemsType, expectedDescription string, expectedRequired, expectedComputed bool, expectedSchema []Property) {
	assert.Equal(t, expectedName, actualProp.Name)
	assert.Equal(t, expectedType, actualProp.Type)
	assert.Equal(t, expectedArrayItemsType, actualProp.ArrayItemsType)
	assert.Equal(t, expectedDescription, actualProp.Description)
	assert.Equal(t, expectedRequired, actualProp.Required)
	assert.Equal(t, expectedComputed, actualProp.Computed)
	if expectedSchema != nil {
		for i, expectedProp := range expectedSchema {
			assertProperty(t, actualProp.Schema[i], expectedProp.Name, expectedProp.Type, expectedProp.ArrayItemsType, expectedProp.Description, expectedProp.Required, expectedProp.Computed, expectedProp.Schema)
		}
	}
}

func TestOrderProps(t *testing.T) {
	inputProps := []Property{
		{Name: "prop3"},
		{Name: "prop1"},
		{Name: "prop2"},
	}
	orderedProps := orderProps(inputProps)
	expectedProps := []Property{
		{Name: "prop2"},
		{Name: "prop1"},
		{Name: "prop3"},
	}
	assert.Equal(t, expectedProps, orderedProps)
}
