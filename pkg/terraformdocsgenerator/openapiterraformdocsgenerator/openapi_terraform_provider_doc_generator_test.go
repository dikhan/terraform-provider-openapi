package openapiterraformdocsgenerator

import (
	"errors"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSwaggerServer(t *testing.T, swagger string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(swagger))
		assert.NoError(t, err)
	}))
}

func TestNewTerraformProviderDocGenerator(t *testing.T) {
	testSwagger := `swagger: "2.0"`
	swaggerServer := newSwaggerServer(t, testSwagger)
	defer swaggerServer.Close()

	providerName := "openapi"
	dg, err := NewTerraformProviderDocGenerator(providerName, swaggerServer.URL)
	assert.Equal(t, providerName, dg.ProviderName)
	assert.NotNil(t, dg.SpecAnalyser)
	assert.NoError(t, err)
}

func TestNewTerraformProviderDocGenerator_ErrorLoadingSpecAnalyser(t *testing.T) {
	dg, err := NewTerraformProviderDocGenerator("openapi", "badURL")
	assert.Empty(t, dg)
	assert.EqualError(t, err, "failed to retrieve the OpenAPI document from 'badURL' - error = open badURL: no such file or directory")
}

func TestGenerateDocumentation(t *testing.T) {
	providerName := "openapi"
	resources := []openapi.SpecResource{
		&specStubResource{
			name:         "cdn_v1",
			shouldIgnore: false,
			schemaDefinition: &openapi.SpecSchemaDefinition{
				Properties: openapi.SpecSchemaDefinitionProperties{
					&openapi.SpecSchemaDefinitionProperty{
						Name:     "id",
						Type:     openapi.TypeString,
						Required: false,
						Computed: true,
					},
				},
			},
		},
		&specStubResource{
			name:         "lb_v1",
			shouldIgnore: false,
			schemaDefinition: &openapi.SpecSchemaDefinition{
				Properties: openapi.SpecSchemaDefinitionProperties{
					&openapi.SpecSchemaDefinitionProperty{
						Name:     "id",
						Type:     openapi.TypeString,
						Required: false,
						Computed: true,
					},
				},
			},
		},
	}
	dg := TerraformProviderDocGenerator{
		ProviderName: providerName,
		SpecAnalyser: &specAnalyserStub{
			resources:   func() ([]openapi.SpecResource, error) { return resources, nil },
			dataSources: func() []openapi.SpecResource { return resources },
			security: &specSecurityStub{
				globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) {
					return openapi.SpecSecuritySchemes{{Name: "required_token"}}, nil
				},
				securityDefinitions: func() (*openapi.SpecSecurityDefinitions, error) {
					return &openapi.SpecSecurityDefinitions{specStubSecurityDefinition{name: "required_token"}}, nil
				},
			},
			headers: openapi.SpecHeaderParameters{},
			backendConfiguration: func() (*specStubBackendConfiguration, error) {
				return &specStubBackendConfiguration{
					host:    "service.api.${region}.hostname.com",
					regions: []string{"region1", "region2", "region3"},
				}, nil
			},
		},
	}
	d, err := dg.GenerateDocumentation()
	assert.NoError(t, err)
	assert.Equal(t, providerName, d.ProviderName)

	// ProviderInstallation assertions
	assert.Equal(t, providerName, d.ProviderInstallation.ProviderName)
	expectedProviderInstallExample := fmt.Sprintf("$ export PROVIDER_NAME=%s && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME<br>"+
		"[INFO] Downloading https://github.com/dikhan/terraform-provider-openapi/v2/releases/download/v0.29.4/terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz in temporally folder /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh...<br>"+
		"[INFO] Extracting terraform-provider-openapi from terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz...<br>"+
		"[INFO] Cleaning up tmp dir created for installation purposes: /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh<br>"+
		"[INFO] Terraform provider 'terraform-provider-%s' successfully installed at: '~/.terraform.d/plugins'!", providerName, providerName)
	assert.Equal(t, expectedProviderInstallExample, d.ProviderInstallation.Example)
	assert.Equal(t, "You can then start running the Terraform provider:", d.ProviderInstallation.Other)
	assert.Equal(t, fmt.Sprintf("$ export OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE=\"https://api.service.com/openapi.yaml\"<br>", providerName), d.ProviderInstallation.OtherCommand)

	// ProviderConfiguration assertions
	assert.Equal(t, providerName, d.ProviderConfiguration.ProviderName)
	assert.Equal(t, []string{"region1", "region2", "region3"}, d.ProviderConfiguration.Regions)
	assert.Equal(t, []Property{{Name: "required_token", Type: "string", ArrayItemsType: "", Required: true, Computed: false, Description: "", Schema: nil}}, d.ProviderConfiguration.ConfigProperties)
	assert.Nil(t, d.ProviderConfiguration.ExampleUsage)
	assert.Equal(t, ArgumentsReference{Notes: nil}, d.ProviderConfiguration.ArgumentsReference)

	// ProviderResources assertions
	assert.Equal(t, providerName, d.ProviderResources.ProviderName)
	assert.Len(t, d.ProviderResources.Resources, 2)

	cdnResource := d.ProviderResources.Resources[0]
	assert.Equal(t, "cdn_v1", cdnResource.Name)
	assert.Equal(t, "", cdnResource.Description)
	assert.Equal(t, ArgumentsReference{Notes: []string{}}, cdnResource.ArgumentsReference)
	cdnResourceProps := cdnResource.Properties
	assert.Len(t, cdnResourceProps, 1)
	assertProperty(t, cdnResourceProps[0], "id", "string", "", "", false, true, nil)

	assert.Equal(t, providerName, d.ProviderResources.ProviderName)
	lbResource := d.ProviderResources.Resources[1]
	assert.Equal(t, "lb_v1", lbResource.Name)
	assert.Equal(t, "", lbResource.Description)
	assert.Equal(t, ArgumentsReference{Notes: []string{}}, lbResource.ArgumentsReference)
	lbResourceProps := lbResource.Properties
	assert.Len(t, lbResourceProps, 1)
	assertProperty(t, lbResourceProps[0], "id", "string", "", "", false, true, nil)

	// DataSources assertions
	assert.Equal(t, providerName, d.DataSources.ProviderName)

	// DataSource (filters) assertions
	assert.Len(t, d.DataSources.DataSources, 2)
	cdnDataSource := d.DataSources.DataSources[0]
	assert.Equal(t, "cdn_v1", cdnDataSource.Name)
	assert.Equal(t, "", cdnDataSource.OtherExample)
	cdnDataSourceProps := cdnDataSource.Properties
	assert.Len(t, cdnResourceProps, 1)
	assertDataSourceProperty(t, cdnDataSourceProps[0], "id", "string", "", "", nil)

	lbDataSource := d.DataSources.DataSources[1]
	assert.Equal(t, "lb_v1", lbDataSource.Name)
	assert.Equal(t, "", lbDataSource.OtherExample)
	lbDataSourceProps := lbDataSource.Properties
	assert.Len(t, cdnResourceProps, 1)
	assertDataSourceProperty(t, lbDataSourceProps[0], "id", "string", "", "", nil)

	// DataSourceInstance assertions
	assert.Len(t, d.DataSources.DataSourceInstances, 2)
	cdnDataSourceInstance := d.DataSources.DataSourceInstances[0]
	assert.Equal(t, "cdn_v1_instance", cdnDataSourceInstance.Name)
	assert.Equal(t, "", cdnDataSourceInstance.OtherExample)
	cdnDataSourceInstanceProps := cdnDataSourceInstance.Properties
	assert.Len(t, cdnDataSourceInstanceProps, 1)
	assertDataSourceProperty(t, cdnDataSourceInstanceProps[0], "id", "string", "", "", nil)

	lbDataSourceInstance := d.DataSources.DataSourceInstances[1]
	assert.Equal(t, "lb_v1_instance", lbDataSourceInstance.Name)
	assert.Equal(t, "", lbDataSourceInstance.OtherExample)
	lbDataSourceInstanceProps := lbDataSourceInstance.Properties
	assert.Len(t, lbDataSourceInstanceProps, 1)
	assertDataSourceProperty(t, lbDataSourceInstanceProps[0], "id", "string", "", "", nil)
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

func TestGenerateDocumentation_ErrorCases(t *testing.T) {
	testCases := []struct {
		name         string
		specAnalyser *specAnalyserStub
		expectedErr  error
	}{
		{
			name: "getRegions error",
			specAnalyser: &specAnalyserStub{
				backendConfiguration: func() (*specStubBackendConfiguration, error) {
					return nil, errors.New("getRegions error")
				},
			},
			expectedErr: errors.New("getRegions error"),
		},
		{
			name: "getSecurity error",
			specAnalyser: &specAnalyserStub{
				security: &specSecurityStub{
					globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) { return nil, errors.New("getSecurity error") },
				},
			},
			expectedErr: errors.New("getSecurity error"),
		},
		{
			name: "GetTerraformCompliantResources error",
			specAnalyser: &specAnalyserStub{
				resources: func() ([]openapi.SpecResource, error) {
					return nil, errors.New("GetTerraformCompliantResources error")
				},
			},
			expectedErr: errors.New("GetTerraformCompliantResources error"),
		},
		{
			name: "getProviderResources error",
			specAnalyser: &specAnalyserStub{
				resources: func() ([]openapi.SpecResource, error) {
					return []openapi.SpecResource{
						&specStubResource{
							name:  "test_resource",
							error: errors.New("getProviderResources error"),
						},
					}, nil
				},
			},
			expectedErr: errors.New("getProviderResources error"),
		},
		{
			name: "getDataSourceFilters error",
			specAnalyser: &specAnalyserStub{
				dataSources: func() []openapi.SpecResource {
					return []openapi.SpecResource{
						&specStubResource{
							name:  "test_datasource",
							error: errors.New("getDataSourceFilters error"),
						},
					}
				},
			},
			expectedErr: errors.New("getDataSourceFilters error"),
		},
	}

	for _, tc := range testCases {
		dg := TerraformProviderDocGenerator{SpecAnalyser: tc.specAnalyser}
		d, err := dg.GenerateDocumentation()
		assert.Empty(t, d, tc.name)
		assert.EqualError(t, err, tc.expectedErr.Error(), tc.name)
	}
}

func TestGetRegions(t *testing.T) {
	testCases := []struct {
		name            string
		specAnalyser    *specAnalyserStub
		expectedRegions []string
		expectedErr     error
	}{
		{
			name: "happy path",
			specAnalyser: &specAnalyserStub{
				backendConfiguration: func() (*specStubBackendConfiguration, error) {
					return &specStubBackendConfiguration{
						host:    "service.api.${region}.hostname.com",
						regions: []string{"region1", "region2"},
					}, nil
				},
			},
			expectedRegions: []string{"region1", "region2"},
			expectedErr:     nil,
		},
		{
			name:            "happy path - not multi region",
			specAnalyser:    &specAnalyserStub{},
			expectedRegions: nil,
			expectedErr:     nil,
		},
		{
			name: "crappy path - IsMultiRegion error",
			specAnalyser: &specAnalyserStub{
				backendConfiguration: func() (*specStubBackendConfiguration, error) {
					return &specStubBackendConfiguration{
						err: errors.New("IsMultiRegion error"),
					}, nil
				},
			},
			expectedRegions: nil,
			expectedErr:     errors.New("IsMultiRegion error"),
		},
		{
			name: "crappy path - GetAPIBackendConfiguration error",
			specAnalyser: &specAnalyserStub{
				backendConfiguration: func() (*specStubBackendConfiguration, error) {
					return nil, errors.New("GetAPIBackendConfiguration error")
				},
			},
			expectedRegions: nil,
			expectedErr:     errors.New("GetAPIBackendConfiguration error"),
		},
	}

	for _, tc := range testCases {
		regions, err := getRegions(tc.specAnalyser)
		if tc.expectedErr != nil {
			assert.Nil(t, regions, tc.name)
			assert.EqualError(t, err, tc.expectedErr.Error(), tc.name)
		} else {
			assert.NoError(t, err, tc.name)
			assert.Equal(t, tc.expectedRegions, regions, tc.name)
		}
	}
}

func TestGetSecurity(t *testing.T) {
	testCases := []struct {
		name                        string
		specAnalyser                *specAnalyserStub
		expectedSecuritySchemes     openapi.SpecSecuritySchemes
		expectedSecurityDefinitions *openapi.SpecSecurityDefinitions
		expectedErr                 error
	}{
		{
			name: "happy path",
			specAnalyser: &specAnalyserStub{
				security: &specSecurityStub{
					globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) {
						return openapi.SpecSecuritySchemes{{Name: "required_token"}}, nil
					},
					securityDefinitions: func() (*openapi.SpecSecurityDefinitions, error) {
						return &openapi.SpecSecurityDefinitions{specStubSecurityDefinition{name: "required_token"}}, nil
					},
				},
			},
			expectedSecuritySchemes:     openapi.SpecSecuritySchemes{{Name: "required_token"}},
			expectedSecurityDefinitions: &openapi.SpecSecurityDefinitions{specStubSecurityDefinition{name: "required_token"}},
			expectedErr:                 nil,
		},
		{
			name:                        "happy path - no spec security",
			specAnalyser:                &specAnalyserStub{},
			expectedSecuritySchemes:     nil,
			expectedSecurityDefinitions: nil,
			expectedErr:                 nil,
		},
		{
			name: "crappy path - security schemes error",
			specAnalyser: &specAnalyserStub{
				security: &specSecurityStub{
					globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) { return nil, errors.New("globalSecuritySchemes error") },
					securityDefinitions: func() (*openapi.SpecSecurityDefinitions, error) {
						return &openapi.SpecSecurityDefinitions{specStubSecurityDefinition{name: "required_token"}}, nil
					},
				},
			},
			expectedSecuritySchemes:     nil,
			expectedSecurityDefinitions: nil,
			expectedErr:                 errors.New("globalSecuritySchemes error"),
		},
		{
			name: "crappy path - security definitions error",
			specAnalyser: &specAnalyserStub{
				security: &specSecurityStub{
					globalSecuritySchemes: func() (openapi.SpecSecuritySchemes, error) {
						return openapi.SpecSecuritySchemes{{Name: "required_token"}}, nil
					},
					securityDefinitions: func() (*openapi.SpecSecurityDefinitions, error) { return nil, errors.New("securityDefinitions error") }},
			},
			expectedSecuritySchemes:     nil,
			expectedSecurityDefinitions: nil,
			expectedErr:                 errors.New("securityDefinitions error"),
		},
	}

	for _, tc := range testCases {
		securitySchemes, securityDefinitions, err := getSecurity(tc.specAnalyser)
		if tc.expectedErr != nil {
			assert.Nil(t, securitySchemes, tc.name)
			assert.Nil(t, securityDefinitions, tc.name)
			assert.EqualError(t, err, tc.expectedErr.Error(), tc.name)
		} else {
			assert.NoError(t, err, tc.name)
			assert.Equal(t, tc.expectedSecuritySchemes, securitySchemes, tc.name)
			assert.Equal(t, tc.expectedSecurityDefinitions, securityDefinitions, tc.name)
		}
	}
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
			name: "happy path - obj prop with multiple child props (child props should be ordered by their hash)",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "obj_prop",
					Type: openapi.TypeObject,
					SpecSchemaDefinition: &openapi.SpecSchemaDefinition{
						Properties: openapi.SpecSchemaDefinitionProperties{
							{Name: "string_prop2", Type: openapi.TypeString},
							{Name: "string_prop3", Type: openapi.TypeString},
							{Name: "string_prop1", Type: openapi.TypeString},
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
						{Name: "string_prop3", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
						{Name: "string_prop1", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
						{Name: "string_prop2", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
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
		assert.NoError(t, err, tc.name)
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
			name: "happy path - obj prop with multiple child props (child props should be ordered according to their hash)",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "obj_prop",
					Type: openapi.TypeObject,
					SpecSchemaDefinition: &openapi.SpecSchemaDefinition{
						Properties: openapi.SpecSchemaDefinitionProperties{
							{Name: "string_prop3", Type: openapi.TypeString},
							{Name: "string_prop1", Type: openapi.TypeString},
							{Name: "string_prop2", Type: openapi.TypeString},
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
						{Name: "string_prop3", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
						{Name: "string_prop1", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
						{Name: "string_prop2", Type: "string", Required: false, Computed: true, IsOptionalComputed: true},
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
		assert.NoError(t, err, tc.name)
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
			name: "happy path - obj prop with multiple child props (child props should be ordered according to their hash)",
			openapiProps: openapi.SpecSchemaDefinitionProperties{
				&openapi.SpecSchemaDefinitionProperty{
					Name: "obj_prop",
					Type: openapi.TypeObject,
					SpecSchemaDefinition: &openapi.SpecSchemaDefinition{
						Properties: openapi.SpecSchemaDefinitionProperties{
							{Name: "string_prop3", Type: openapi.TypeString},
							{Name: "string_prop1", Type: openapi.TypeString},
							{Name: "string_prop2", Type: openapi.TypeString},
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
						{Name: "string_prop3", Type: "string", Required: false, Computed: false},
						{Name: "string_prop1", Type: "string", Required: false, Computed: false},
						{Name: "string_prop2", Type: "string", Required: false, Computed: false},
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
		assert.NoError(t, err, tc.name)
		assert.Equal(t, expectedResources, actualResources, tc.name)
	}
}

func TestGetProviderResources_HasParentProps(t *testing.T) {
	openapiResources := []openapi.SpecResource{
		&specStubResource{
			//name: "test_resource",
			schemaDefinition:    &openapi.SpecSchemaDefinition{},
			parentResourceNames: []string{"parentResourceName"},
		},
	}
	dg := TerraformProviderDocGenerator{}
	actualResources, err := dg.getProviderResources(openapiResources)

	assert.NoError(t, err)
	assert.Equal(t, "parentResourceName_id", actualResources[0].ParentProperties[0])
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

	assert.NoError(t, err)
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
		{Name: "prop2"},
		{Name: "prop1"},
		{Name: "prop3"},
	}
	assert.Equal(t, expectedProps, orderedProps)
}
