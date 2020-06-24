package openapi

import (
	"errors"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/stretchr/testify/assert"
	"testing"
)

//const testSwagger = `swagger: "2.0"
//host: service.api.hostname.com
//schemes:
//- "https"
//x-terraform-provider-multiregion-fqdn: "service.api.${region}.hostname.com"
//x-terraform-provider-regions: "region1, region2, region3"
//security:
//  - authToken: []
//securityDefinitions:
//  authToken:
//    in: header
//    name: auth
//    type: apiKey
//    x-terraform-authentication-scheme-bearer: true
//
//paths:
//  ######################
//  #### CDN Resource ####
//  ######################
//
//  /v1/cdns:
//    x-terraform-resource-name: "cdn"
//    get:
//      summary: "Get all cdns"
//      responses:
//        200:
//          description: "successful operation"
//          schema:
//            $ref: "#/definitions/ContentDeliveryNetworkCollectionV1"
//    post:
//      summary: "Create cdn"
//      operationId: "ContentDeliveryNetworkCreateV1"
//      parameters:
//      - in: "body"
//        name: "body"
//        description: "Created CDN"
//        required: true
//        schema:
//          $ref: "#/definitions/ContentDeliveryNetworkV1"
//      responses:
//        201:
//          description: "successful operation"
//          schema:
//            $ref: "#/definitions/ContentDeliveryNetworkV1"
//  /v1/cdns/{id}:
//    get:
//      summary: "Get cdn by id"
//      description: "description of cdn get operation"
//      operationId: "ContentDeliveryNetworkGetV1"
//      parameters:
//      - name: "id"
//        in: "path"
//        description: "The cdn id that needs to be fetched."
//        required: true
//        type: "string"
//      responses:
//        200:
//          description: "successful operation"
//          schema:
//            $ref: "#/definitions/ContentDeliveryNetworkV1"
//definitions:
//  ContentDeliveryNetworkCollectionV1:
//    type: "array"
//    items:
//      $ref: "#/definitions/ContentDeliveryNetworkV1"
//  ContentDeliveryNetworkV1:
//    type: "object"
//    required:
//      - required_prop
//    properties:
//      id:
//        type: "string"
//        readOnly: true
//      required_prop:
//        type: "string"
//      list_prop:
//        type: "array"
//        items:
//          type: "string"
//      obj_prop:
//        type: "object"
//        properties:
//          label:
//            type: "string"`
//
//func TestGenerateDocumentation(t *testing.T) {
//	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		_, err := w.Write([]byte(testSwagger))
//		assert.Nil(t, err)
//	}))
//	defer swaggerServer.Close()
//
//	providerName := "openapi"
//	dg := TerraformProviderDocGenerator{
//		ProviderName:  providerName,
//		OpenAPIDocURL: swaggerServer.URL,
//	}
//	d, err := dg.GenerateDocumentation()
//	assert.Nil(t, err)
//	assert.Equal(t, providerName, d.ProviderName)
//
//	// ProviderInstallation assertions
//	expectedProviderInstallExample := fmt.Sprintf("$ export PROVIDER_NAME=%s && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME<br>"+
//		"[INFO] Downloading https://github.com/dikhan/terraform-provider-openapi/releases/download/v0.29.4/terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz in temporally folder /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh...<br>"+
//		"[INFO] Extracting terraform-provider-openapi from terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz...<br>"+
//		"[INFO] Cleaning up tmp dir created for installation purposes: /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh<br>"+
//		"[INFO] Terraform provider 'terraform-provider-%s' successfully installed at: '~/.terraform.d/plugins'!", providerName, providerName)
//	assert.Equal(t, expectedProviderInstallExample, d.ProviderInstallation.Example)
//	assert.Equal(t, "You can then start running the Terraform provider:", d.ProviderInstallation.Other)
//	assert.Equal(t, fmt.Sprintf("$ export OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE=\"https://api.service.com/openapi.yaml\"<br>", providerName), d.ProviderInstallation.OtherCommand)
//
//	// ProviderConfiguration assertions
//	assert.Equal(t, []string{"region1", "region2", "region3"}, d.ProviderConfiguration.Regions)
//	assert.Equal(t, []Property{{Name: "auth_token", Type: "string", ArrayItemsType: "", Required: true, Computed: false, Description: "", Schema: nil}}, d.ProviderConfiguration.ConfigProperties)
//	assert.Nil(t, d.ProviderConfiguration.ExampleUsage)
//	assert.Equal(t, ArgumentsReference{Notes: nil}, d.ProviderConfiguration.ArgumentsReference)
//
//	// ProviderResources assertions
//	assert.Len(t, d.ProviderResources.Resources, 1)
//	cdnResource := d.ProviderResources.Resources[0]
//	assert.Equal(t, "cdn_v1", cdnResource.Name)
//	assert.Equal(t, "", cdnResource.Description)
//	assert.Equal(t, ArgumentsReference{Notes: []string{}}, cdnResource.ArgumentsReference)
//	cdnResourceProps := cdnResource.Properties
//	assert.Len(t, cdnResourceProps, 4)
//	assertProperty(t, cdnResourceProps[0], "required_prop", "string", "", "", true, false, nil)
//	assertProperty(t, cdnResourceProps[1], "id", "string", "", "", false, true, nil)
//	assertProperty(t, cdnResourceProps[2], "obj_prop", "object", "", "", false, false,
//		[]Property{{Name: "label", Type: "string", ArrayItemsType: "", Required: false, Computed: false, Description: "", Schema: nil}})
//	assertProperty(t, cdnResourceProps[3], "list_prop", "list", "string", "", false, false, nil)
//
//	// DataSource assertions
//	assert.Len(t, d.DataSources.DataSources, 1)
//	cdnDataSource := d.DataSources.DataSources[0]
//	assert.Equal(t, "cdn_v1", cdnDataSource.Name)
//	assert.Equal(t, "", cdnDataSource.OtherExample)
//	cdnDataSourceProps := cdnDataSource.Properties
//	assert.Len(t, cdnResourceProps, 4)
//	assertDataSourceProperty(t, cdnDataSourceProps[0], "list_prop", "list", "string", "", nil)
//	assertDataSourceProperty(t, cdnDataSourceProps[1], "required_prop", "string", "", "", nil)
//	assertDataSourceProperty(t, cdnDataSourceProps[2], "id", "string", "", "", nil)
//	assertDataSourceProperty(t, cdnDataSourceProps[3], "obj_prop", "object", "", "",
//		[]Property{{Name: "label", Type: "string", ArrayItemsType: "", Required: false, Computed: true, Description: "", Schema: nil}})
//
//	// DataSourceInstance assertions
//	assert.Len(t, d.DataSources.DataSourceInstances, 1)
//	cdnDataSourceInstance := d.DataSources.DataSourceInstances[0]
//	assert.Equal(t, "cdn_v1_instance", cdnDataSourceInstance.Name)
//	assert.Equal(t, "", cdnDataSourceInstance.OtherExample)
//	cdnDataSourceInstanceProps := cdnDataSourceInstance.Properties
//	assert.Len(t, cdnDataSourceInstanceProps, 4)
//	assertDataSourceProperty(t, cdnDataSourceInstanceProps[0], "list_prop", "list", "string", "", nil)
//	assertDataSourceProperty(t, cdnDataSourceInstanceProps[1], "required_prop", "string", "", "", nil)
//	assertDataSourceProperty(t, cdnDataSourceInstanceProps[2], "id", "string", "", "", nil)
//	assertDataSourceProperty(t, cdnDataSourceInstanceProps[3], "obj_prop", "object", "", "",
//		[]Property{{Name: "label", Type: "string", ArrayItemsType: "", Required: false, Computed: true, Description: "", Schema: nil}})
//}
//
//func assertDataSourceProperty(t *testing.T, actualProp Property, expectedName, expectedType, expectedArrayItemsType, expectedDescription string, expectedSchema []Property) {
//	assertProperty(t, actualProp, expectedName, expectedType, expectedArrayItemsType, expectedDescription, false, true, expectedSchema)
//}
//
//func assertProperty(t *testing.T, actualProp Property, expectedName, expectedType, expectedArrayItemsType, expectedDescription string, expectedRequired, expectedComputed bool, expectedSchema []Property) {
//	assert.Equal(t, expectedName, actualProp.Name)
//	assert.Equal(t, expectedType, actualProp.Type)
//	assert.Equal(t, expectedArrayItemsType, actualProp.ArrayItemsType)
//	assert.Equal(t, expectedDescription, actualProp.Description)
//	assert.Equal(t, expectedRequired, actualProp.Required)
//	assert.Equal(t, expectedComputed, actualProp.Computed)
//	if expectedSchema != nil {
//		for i, expectedProp := range expectedSchema {
//			assertProperty(t, actualProp.Schema[i], expectedProp.Name, expectedProp.Type, expectedProp.ArrayItemsType, expectedProp.Description, expectedProp.Required, expectedProp.Computed, expectedProp.Schema)
//		}
//	}
//}

func TestGetRegions(t *testing.T) {
	sa := specAnalyserStub{
		backendConfiguration: &specStubBackendConfiguration{
			host:    "service.api.${region}.hostname.com",
			regions: []string{"region1", "region2"},
		},
	}
	regions, err := getRegions(&sa)
	assert.Nil(t, err)
	assert.Equal(t, sa.backendConfiguration.regions, regions)
}

func TestGetRegions_NotMultiRegion(t *testing.T) {
	sa := specAnalyserStub{}
	regions, err := getRegions(&sa)
	assert.Nil(t, regions)
	assert.Nil(t, err)
}

func TestGetRegions_BackendConfigError(t *testing.T) {
	sa := specAnalyserStub{backendConfiguration: &specStubBackendConfiguration{err: errors.New("specStubBackendConfiguration error")}}
	regions, err := getRegions(&sa)
	assert.Nil(t, regions)
	assert.EqualError(t, err, "specStubBackendConfiguration error")
}

func TestGetRegions_SpecAnalyserError(t *testing.T) {
	sa := specAnalyserStub{error: errors.New("specAnalyser error")}
	regions, err := getRegions(&sa)
	assert.Nil(t, regions)
	assert.EqualError(t, err, "specAnalyser error")
}

func TestGetSecurity(t *testing.T) {
	expectedSecuritySchemes := openapi.SpecSecuritySchemes{{Name: "required_token"}}
	expectedSecurityDefinitions := &openapi.SpecSecurityDefinitions{specStubSecurityDefinition{name: "required_token"}}
	sa := specAnalyserStub{
		security: &specSecurityStub{
			globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) { return expectedSecuritySchemes, nil },
			securityDefinitions:   func() (*openapi.SpecSecurityDefinitions, error) { return expectedSecurityDefinitions, nil },
		},
	}
	securitySchemes, securityDefinitions, err := getSecurity(&sa)
	assert.Nil(t, err)
	assert.Equal(t, expectedSecuritySchemes, securitySchemes)
	assert.Equal(t, expectedSecurityDefinitions, securityDefinitions)
}

func TestGetSecurity_NoSpecSecurity(t *testing.T) {
	sa := specAnalyserStub{}
	securitySchemes, securityDefinitions, err := getSecurity(&sa)
	assert.Nil(t, err)
	assert.Nil(t, securitySchemes)
	assert.Nil(t, securityDefinitions)
}

func TestGetSecurity_SecuritySchemesError(t *testing.T) {
	sa := specAnalyserStub{
		security: &specSecurityStub{
			globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) { return nil, errors.New("globalSecuritySchemes error") },
			securityDefinitions: func() (*openapi.SpecSecurityDefinitions, error) {
				return &openapi.SpecSecurityDefinitions{specStubSecurityDefinition{name: "required_token"}}, nil
			},
		},
	}
	securitySchemes, securityDefinitions, err := getSecurity(&sa)
	assert.Nil(t, securitySchemes)
	assert.Nil(t, securityDefinitions)
	assert.EqualError(t, err, "globalSecuritySchemes error")
}

func TestGetSecurity_SecurityDefinitionsError(t *testing.T) {
	sa := specAnalyserStub{
		security: &specSecurityStub{
			globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) {
				return openapi.SpecSecuritySchemes{{Name: "required_token"}}, nil
			},
			securityDefinitions: func() (*openapi.SpecSecurityDefinitions, error) { return nil, errors.New("securityDefinitions error") },
		},
	}
	securitySchemes, securityDefinitions, err := getSecurity(&sa)
	assert.Nil(t, securitySchemes)
	assert.Nil(t, securityDefinitions)
	assert.EqualError(t, err, "securityDefinitions error")
}

func TestGetDataSourceFilters(t *testing.T) {
	testCases := []struct {
		name          string
		openapiProps  openapi.SpecSchemaDefinitionProperties
		expectedProps []Property
	}{
		{
			name: "happy path - string prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "string_prop",
					Type: openapi.TypeString,
				},
			},
			expectedProps: []Property{{Name: "string_prop", Type: "string", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - int prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "int_prop",
					Type: openapi.TypeInt,
				},
			},
			expectedProps: []Property{{Name: "int_prop", Type: "integer", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - float prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "float_prop",
					Type: openapi.TypeFloat,
				},
			},
			expectedProps: []Property{{Name: "float_prop", Type: "number", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - list prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name:           "list_prop",
					Type:           openapi.TypeList,
					ArrayItemsType: openapi.TypeString,
				},
			},
			expectedProps: []Property{{Name: "list_prop", Type: "list", ArrayItemsType: "string", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - obj prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "obj_prop",
					Type: openapi.TypeObject,
					SpecSchemaDefinition: &openapi.SpecSchemaDefinition{
						Properties: openapi.SpecSchemaDefinitionProperties{
							{Name: "string_prop", Type: openapi.TypeString},
						},
					},
				},
			},
			expectedProps: []Property{
				{
					Name:               "obj_prop",
					Type:               "object",
					Required:           false,
					Computed:           true,
					IsOptionalComputed: true,
					Schema: []Property{
						{Name: "string_prop", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		openapiDataSources := []openapi.SpecResource{
			&specStubResource{
				name: "test_resource",
				schemaDefinition: &openapi.SpecSchemaDefinition{
					Properties: tc.openapiProps,
				},
			},
		}
		dg := TerraformProviderDocGenerator{}
		actualDataSources, err := dg.getDataSourceFilters(openapiDataSources)

		expectedDataSources := []DataSource{
			{
				Name:         "test_resource",
				Properties:   tc.expectedProps,
				OtherExample: "",
			},
		}
		assert.Nil(t, err, tc.name)
		assert.Equal(t, expectedDataSources, actualDataSources, tc.name)
	}
}

func TestGetDataSourceFilters_Error(t *testing.T) {
	openapiDataSources := []openapi.SpecResource{
		&specStubResource{
			name:  "test_resource",
			error: errors.New("specStubResource error"),
		},
	}
	dg := TerraformProviderDocGenerator{}
	dataSourceFilters, err := dg.getDataSourceFilters(openapiDataSources)

	assert.Nil(t, dataSourceFilters)
	assert.EqualError(t, err, "specStubResource error")
}

func TestGetDataSourceInstances(t *testing.T) {
	testCases := []struct {
		name          string
		openapiProps  openapi.SpecSchemaDefinitionProperties
		expectedProps []Property
	}{
		{
			name: "happy path - string prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "string_prop",
					Type: openapi.TypeString,
				},
			},
			expectedProps: []Property{{Name: "string_prop", Type: "string", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - int prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "int_prop",
					Type: openapi.TypeInt,
				},
			},
			expectedProps: []Property{{Name: "int_prop", Type: "integer", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - float prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "float_prop",
					Type: openapi.TypeFloat,
				},
			},
			expectedProps: []Property{{Name: "float_prop", Type: "number", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - list prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name:           "list_prop",
					Type:           openapi.TypeList,
					ArrayItemsType: openapi.TypeString,
				},
			},
			expectedProps: []Property{{Name: "list_prop", Type: "list", ArrayItemsType: "string", Required: false, Computed: true, IsOptionalComputed: true}},
		},
		{
			name: "happy path - obj prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "obj_prop",
					Type: openapi.TypeObject,
					SpecSchemaDefinition: &openapi.SpecSchemaDefinition{
						Properties: openapi.SpecSchemaDefinitionProperties{
							{Name: "string_prop", Type: openapi.TypeString},
						},
					},
				},
			},
			expectedProps: []Property{
				{
					Name:               "obj_prop",
					Type:               "object",
					Required:           false,
					Computed:           true,
					IsOptionalComputed: true,
					Schema: []Property{
						{Name: "string_prop", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		openapiResources := []openapi.SpecResource{
			&specStubResource{
				name: "test_resource",
				schemaDefinition: &openapi.SpecSchemaDefinition{
					Properties: tc.openapiProps,
				},
			},
		}
		dg := TerraformProviderDocGenerator{}
		dataSourceInstances, err := dg.getDataSourceInstances(openapiResources)

		expectedDataSourceInstances := []DataSource{
			{
				Name:         "test_resource_instance",
				Properties:   tc.expectedProps,
				OtherExample: "",
			},
		}
		assert.Nil(t, err, tc.name)
		assert.Equal(t, expectedDataSourceInstances, dataSourceInstances, tc.name)
	}
}

func TestGetDataSourceInstances_Error(t *testing.T) {
	openapiResources := []openapi.SpecResource{
		&specStubResource{
			name:  "test_resource",
			error: errors.New("specStubResource error"),
		},
	}
	dg := TerraformProviderDocGenerator{}
	dataSourceInstances, err := dg.getDataSourceInstances(openapiResources)

	assert.Nil(t, dataSourceInstances)
	assert.EqualError(t, err, "specStubResource error")
}

func TestGetProviderResources(t *testing.T) {
	testCases := []struct {
		name          string
		openapiProps  openapi.SpecSchemaDefinitionProperties
		expectedProps []Property
	}{
		{
			name: "happy path - string prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "string_prop",
					Type: openapi.TypeString,
				},
			},
			expectedProps: []Property{{Name: "string_prop", Type: "string", Required: false, Computed: false}},
		},
		{
			name: "happy path - int prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "int_prop",
					Type: openapi.TypeInt,
				},
			},
			expectedProps: []Property{{Name: "int_prop", Type: "integer", Required: false, Computed: false}},
		},
		{
			name: "happy path - float prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "float_prop",
					Type: openapi.TypeFloat,
				},
			},
			expectedProps: []Property{{Name: "float_prop", Type: "number", Required: false, Computed: false}},
		},
		{
			name: "happy path - list prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name:           "list_prop",
					Type:           openapi.TypeList,
					ArrayItemsType: openapi.TypeString,
				},
			},
			expectedProps: []Property{{Name: "list_prop", Type: "list", ArrayItemsType: "string", Required: false, Computed: false}},
		},
		{
			name: "happy path - obj prop",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "obj_prop",
					Type: openapi.TypeObject,
					SpecSchemaDefinition: &openapi.SpecSchemaDefinition{
						Properties: openapi.SpecSchemaDefinitionProperties{
							{Name: "string_prop", Type: openapi.TypeString},
						},
					},
				},
			},
			expectedProps: []Property{
				{
					Name:     "obj_prop",
					Type:     "object",
					Required: false,
					Computed: false,
					Schema: []Property{
						{Name: "string_prop", Type: "string", Required: false, Computed: false},
					},
				},
			},
		},
		{
			name: "happy path - required props should be listed before optional props",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name:     "optional_prop",
					Type:     openapi.TypeString,
					Required: false,
				},
				&openapi.SpecSchemaDefinitionProperty{
					Name:     "required_prop",
					Type:     openapi.TypeString,
					Required: true,
				},
			},
			expectedProps: []Property{
				{Name: "required_prop", Type: "string", Required: true, Computed: false},
				{Name: "optional_prop", Type: "string", Required: false, Computed: false},
			},
		},
	}
	for _, tc := range testCases {
		openapiResources := []openapi.SpecResource{
			&specStubResource{
				name: "test_resource",
				schemaDefinition: &openapi.SpecSchemaDefinition{
					Properties: tc.openapiProps,
				},
			},
		}
		dg := TerraformProviderDocGenerator{}
		actualResources, err := dg.getProviderResources(openapiResources)

		expectedResources := []Resource{
			{
				Name:               "test_resource",
				Description:        "",
				Properties:         tc.expectedProps,
				ExampleUsage:       nil,
				ArgumentsReference: ArgumentsReference{Notes: []string{}},
			},
		}
		assert.Nil(t, err, tc.name)
		assert.Equal(t, expectedResources, actualResources, tc.name)
	}
}

func TestGetProviderResources_IgnoreResource(t *testing.T) {
	openapiResources := []openapi.SpecResource{
		&specStubResource{
			name:         "ignore_resource",
			shouldIgnore: true,
		},
	}
	dg := TerraformProviderDocGenerator{}
	actualResources, err := dg.getProviderResources(openapiResources)

	assert.Nil(t, err)
	assert.Len(t, actualResources, 0)
}

func TestGetProviderResources_Error(t *testing.T) {
	openapiResources := []openapi.SpecResource{
		&specStubResource{
			name:  "test_resource",
			error: errors.New("specStubResource error"),
		},
	}
	dg := TerraformProviderDocGenerator{}
	actualResources, err := dg.getProviderResources(openapiResources)
	assert.Nil(t, actualResources)
	assert.EqualError(t, err, "specStubResource error")
}

func TestGetRequiredProviderConfigurationProperties(t *testing.T) {
	testCases := []struct {
		name                  string
		regions               []string
		globalSecuritySchemes openapi.SpecSecuritySchemes
		securityDefinitions   *openapi.SpecSecurityDefinitions
		headers               openapi.SpecHeaderParameters
		expectedRegions       []string
		expectedConfigProps   []Property
		expectedErr           error
	}{
		{
			name: "happy path - required security scheme property",
			securityDefinitions: &openapi.SpecSecurityDefinitions{
				specStubSecurityDefinition{name: "required_token"},
			},
			globalSecuritySchemes: []openapi.SpecSecurityScheme{
				{Name: "required_token"},
			},
			expectedConfigProps: []Property{
				{
					Name:           "required_token",
					Type:           "string",
					ArrayItemsType: "",
					Required:       true,
					Computed:       false,
					Description:    "",
					Schema:         nil,
				},
			},
		},
		{
			name: "happy path - optional security scheme property",
			securityDefinitions: &openapi.SpecSecurityDefinitions{
				specStubSecurityDefinition{name: "optional_token"},
			},
			expectedConfigProps: []Property{
				{
					Name:           "optional_token",
					Type:           "string",
					ArrayItemsType: "",
					Required:       false,
					Computed:       false,
					Description:    "",
					Schema:         nil,
				},
			},
		},
		{
			name:            "happy path - multi region",
			regions:         []string{"region1", "region2", "region3"},
			expectedRegions: []string{"region1", "region2", "region3"},
		},
		{
			name: "happy path - with optional header",
			headers: openapi.SpecHeaderParameters{
				{
					Name:       "optional_header",
					IsRequired: false,
				},
			},
			expectedConfigProps: []Property{
				{
					Name:           "optional_header",
					Type:           "string",
					ArrayItemsType: "",
					Required:       false,
					Computed:       false,
					Description:    "",
					Schema:         nil,
				},
			},
		},
		{
			name: "happy path - with required header",
			headers: openapi.SpecHeaderParameters{
				{
					Name:       "required_header",
					IsRequired: true,
				},
			},
			expectedConfigProps: []Property{
				{
					Name:           "required_header",
					Type:           "string",
					ArrayItemsType: "",
					Required:       true,
					Computed:       false,
					Description:    "",
					Schema:         nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		dg := TerraformProviderDocGenerator{}
		regions, configProps := dg.getRequiredProviderConfigurationProperties(tc.regions, tc.globalSecuritySchemes, tc.securityDefinitions, tc.headers)
		assert.Equal(t, tc.expectedRegions, regions, tc.name)
		assert.Equal(t, tc.expectedConfigProps, configProps, tc.name)
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
		{Name: "prop1"},
		{Name: "prop2"},
		{Name: "prop3"},
	}
	assert.Equal(t, expectedProps, orderedProps)
}
