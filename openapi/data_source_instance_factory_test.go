package openapi

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestNewDataSourceInstanceFactory(t *testing.T) {
	openAPIResource := &specStubResource{}
	d := newDataSourceInstanceFactory(openAPIResource)
	assert.NotNil(t, d)
	assert.Equal(t, openAPIResource, d.openAPIResource)
}

func TestGetDataSourceInstanceName(t *testing.T) {
	d := newDataSourceInstanceFactory(&specStubResource{name: "cdn"})
	name := d.getDataSourceInstanceName()
	assert.Equal(t, "cdn_instance", name)
}

func TestCreateTerraformInstanceDataSource(t *testing.T) {
	testCases := []struct {
		name                string
		expectedError       error
		specStubResourceErr error
	}{
		{
			name:                "happy path - Terraform data source created as expected",
			expectedError:       nil,
			specStubResourceErr: nil,
		},
		{
			name:                "crappy path - Terraform data source schema has an error",
			expectedError:       errors.New("data source schema has an error"),
			specStubResourceErr: errors.New("data source schema has an error"),
		},
	}

	for _, tc := range testCases {
		dataSourceFactory := dataSourceInstanceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: specSchemaDefinitionProperties{
						newStringSchemaDefinitionPropertyWithDefaults("id", "", false, true, nil),
						newStringSchemaDefinitionPropertyWithDefaults("label", "", false, false, nil),
					},
				},
				error: tc.specStubResourceErr,
			},
		}

		dataSource, err := dataSourceFactory.createTerraformInstanceDataSource()

		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
			assert.NotNil(t, dataSource, tc.name)
			assert.NotNil(t, dataSource.Read, tc.name)
			assert.Nil(t, dataSource.Delete, tc.name)
			assert.Nil(t, dataSource.Create, tc.name)
			assert.Nil(t, dataSource.Update, tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func TestCreateTerraformDataSourceInstance(t *testing.T) {
	testCases := []struct {
		name            string
		openAPIResource SpecResource
		expectedError   error
	}{
		{
			name: "happy path - data source schema is configured as expected",
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: specSchemaDefinitionProperties{
						newStringSchemaDefinitionPropertyWithDefaults("id", "", false, true, nil),
						newStringSchemaDefinitionPropertyWithDefaults("label", "", false, false, nil),
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "crappy path - data source schema definition is nil",
			openAPIResource: &specStubResource{
				schemaDefinition: nil,
				error:            errors.New("error due to nil schema def"),
			},
			expectedError: errors.New("error due to nil schema def"),
		},
	}

	for _, tc := range testCases {
		dataSourceFactory := dataSourceInstanceFactory{
			openAPIResource: tc.openAPIResource,
		}
		s, err := dataSourceFactory.createTerraformDataSourceInstanceSchema()
		if tc.expectedError == nil {
			assert.NotNil(t, s, tc.name)
			// data source specific input properties
			assert.NotNil(t, s[dataSourceInstanceIDProperty], tc.name)
			assert.True(t, s[dataSourceInstanceIDProperty].Required, tc.name)
			assert.Equal(t, schema.TypeString, s[dataSourceInstanceIDProperty].Type, tc.name)

			// resource specific properties as per swagger def (this properties are meant to be populated by the read operation when a match is found as per the filters)
			assert.Contains(t, s, "label", tc.name)
			assert.True(t, s["label"].Computed, tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func TestDataSourceInstanceRead(t *testing.T) {

	testCases := []struct {
		name           string
		inputID        string
		client         *clientOpenAPIStub
		expectedResult map[string]interface{}
		expectedError  error
	}{
		{
			name:    "fetch selected data source as per input ID",
			inputID: "ID",
			client: &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					"id":     "someID",
					"label":  "someLabel",
					"owners": []string{"someOwner"},
				},
			},
			expectedError: nil,
		},
		{
			name:    "input ID is empty",
			inputID: "",
			client: &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					"id":     "someID",
					"label":  "someLabel",
					"owners": []string{"someOwner"},
				},
			},
			expectedError: errors.New("data source 'id' property value must be populated"),
		},
		{
			name:    "empty response from API",
			inputID: "ID",
			client: &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
			},
			expectedError: errors.New("response object returned from the API is missing mandatory identifier property 'id'"),
		},
		{
			name:    "api returns a non expected code 404",
			inputID: "ID",
			client: &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
				returnHTTPCode:  http.StatusNotFound,
			},
			expectedError: errors.New("[data source instance=''] GET  failed: HTTP Reponse Status Code 404 - Not Found. Could not find resource instance: "),
		},
		{
			name:    "get operation returns an error",
			inputID: "ID",
			client: &clientOpenAPIStub{
				error: errors.New("some api error in the get operation"),
			},
			expectedError: errors.New("some api error in the get operation"),
		},
	}

	dataSourceFactory := dataSourceInstanceFactory{
		openAPIResource: &specStubResource{
			schemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("id", "", false, true, nil),
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, false, nil),
					newListSchemaDefinitionPropertyWithDefaults("owners", "", true, false, false, []string{"value1"}, typeString, nil),
				},
			},
		},
	}

	for _, tc := range testCases {
		resourceSchema, err := dataSourceFactory.createTerraformDataSourceInstanceSchema()
		require.NoError(t, err)

		dataSourceUserInput := map[string]interface{}{
			dataSourceInstanceIDProperty: tc.inputID,
		}
		resourceData := schema.TestResourceDataRaw(t, resourceSchema, dataSourceUserInput)
		// When
		err = dataSourceFactory.read(resourceData, tc.client)
		// Then
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
			assert.Equal(t, tc.client.responsePayload["id"], resourceData.Id(), tc.name)
			assert.Equal(t, tc.client.responsePayload["label"], resourceData.Get("label"), tc.name)
			expectedOwners := tc.client.responsePayload["owners"].([]string)
			owners := resourceData.Get("owners").([]interface{})
			assert.NotNil(t, owners, tc.name)
			assert.NotNil(t, len(expectedOwners), len(owners), tc.name)
			assert.Equal(t, expectedOwners[0], owners[0], tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func TestDataSourceInstanceRead_Fails_Because_Schema_is_not_valid(t *testing.T) {
	dataSourceFactory := dataSourceInstanceFactory{
		openAPIResource: &specStubResource{
			schemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					&specSchemaDefinitionProperty{
						Name: "label",
						Type: "unknown",
					},
				},
			},
		},
	}
	_, err := dataSourceFactory.createTerraformDataSourceInstanceSchema()
	assert.EqualError(t, err, "non supported type unknown")
}

func TestDataSourceInstanceRead_Fails_Because_Cannot_extract_ParentsID(t *testing.T) {
	err := dataSourceInstanceFactory{}.read(nil, &clientOpenAPIStub{})
	assert.EqualError(t, err, "can't get parent ids from a resourceFactory with no openAPIResource")
}

func TestDataSourceInstanceRead_Subresource(t *testing.T) {

	dataSourceFactory := dataSourceInstanceFactory{
		openAPIResource: &specStubResource{
			path: "/v1/cdns/{id}/firewall",
			schemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("id", "", false, true, nil),
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
					newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", false, true, nil), // This simulates an openAPIResource that is subresource and the schema has already been populated with the parent property
				},
			},
			fullParentResourceName: "cdns_v1",
			parentResourceNames:    []string{"cdns_v1"},
			parentPropertyNames:    []string{"cdns_v1_id"},
		},
	}

	dataSourceInstanceSchema, err := dataSourceFactory.createTerraformDataSourceInstanceSchema()
	require.NoError(t, err)

	dataSourceInput := map[string]interface{}{
		"cdns_v1_id":                 "parentPropertyID", // Since the path is a sub-resource, the user is expected to provide the id of the parent
		dataSourceInstanceIDProperty: "someID",
	}
	resourceData := schema.TestResourceDataRaw(t, dataSourceInstanceSchema, dataSourceInput)

	client := &clientOpenAPIStub{
		responsePayload: map[string]interface{}{
			"id":    "someID",
			"label": "my_label",
		},
	}
	err = dataSourceFactory.read(resourceData, client)
	require.NoError(t, err)
	assert.Equal(t, []string{"parentPropertyID"}, client.parentIDsReceived) // check that the parent id is passed as expected
	assert.Equal(t, "someID", resourceData.Id())
	assert.Equal(t, "my_label", resourceData.Get("label"))
}
