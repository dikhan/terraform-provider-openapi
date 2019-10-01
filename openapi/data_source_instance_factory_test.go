package openapi

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
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
