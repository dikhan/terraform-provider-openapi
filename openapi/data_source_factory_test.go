package openapi

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceRead(t *testing.T) {

	testCases := []struct {
		name            string
		filtersInput    []map[string]interface{}
		responsePayload []map[string]interface{}
		expectedResult  map[string]interface{}
		expectedError   error
	}{
		{
			name: "fetch selected data source as per filter configuration (label=someLabel)",
			filtersInput: []map[string]interface{}{
				newFilter("label", []string{"someLabel"}),
			},
			responsePayload: []map[string]interface{}{
				{
					"id":    "someID",
					"label": "someLabel",
				},
				{
					"id":    "someOtherID",
					"label": "someOtherLabel",
				},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		// Given
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: specSchemaDefinitionProperties{
						newStringSchemaDefinitionPropertyWithDefaults("id", "", false, true, nil),
						newStringSchemaDefinitionPropertyWithDefaults("label", "", false, false, nil),
					},
				},
			},
		}
		resourceSchema := dataSourceFactory.createTerraformDataSourceSchema()
		filtersInput := map[string]interface{}{
			dataSourceFilterPropertyName: tc.filtersInput,
		}
		resourceData := schema.TestResourceDataRaw(t, resourceSchema, filtersInput)
		client := &clientOpenAPIStub{
			responseListPayload: tc.responsePayload,
		}
		// When
		err := dataSourceFactory.read(resourceData, client)
		// Then
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
		// assert that the filtered data source contains the same values as the ones returned by the API
		assert.Equal(t, client.responseListPayload[0]["id"], resourceData.Get("id"), tc.name)
		assert.Equal(t, client.responseListPayload[0]["label"], resourceData.Get("label"), tc.name)
	}
}

func TestValidateInput(t *testing.T) {

	testCases := []struct {
		name                 string
		specSchemaDefinition *specSchemaDefinition
		filtersInput         map[string]interface{}
		expectedError        error
		expectedFilters      filters
	}{
		{
			name: "data source populated with a different filters of primitive property types",
			specSchemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newBoolSchemaDefinitionPropertyWithDefaults("bool_primitive", "", false, true, nil),
					newNumberSchemaDefinitionPropertyWithDefaults("number_primitive", "", false, true, nil),
					newIntSchemaDefinitionPropertyWithDefaults("integer_primitive", "", false, true, nil),
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
				},
			},
			filtersInput: map[string]interface{}{
				dataSourceFilterPropertyName: []map[string]interface{}{
					newFilter("integer_primitive", []string{"12345"}),
					newFilter("label", []string{"label_to_fetch"}),
					newFilter("number_primitive", []string{"12.56"}),
					newFilter("bool_primitive", []string{"true"}),
				},
			},
			expectedFilters: filters{},
			expectedError:   nil,
		},
		{
			name: "data source populated with an incorrect filter containing a property that does not match any of the schema definition",
			specSchemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
				},
			},
			filtersInput: map[string]interface{}{
				dataSourceFilterPropertyName: []map[string]interface{}{
					newFilter("non_matching_property_name", []string{"label_to_fetch"}),
				},
			},
			expectedFilters: nil,
			expectedError:   errors.New("filter name does not match any of the schema properties: property with name 'non_matching_property_name' not existing in resource schema definition"),
		},
		{
			name: "data source populated with an incorrect filter containing a property that is not a primitive",
			specSchemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newListSchemaDefinitionPropertyWithDefaults("not_primitive", "", false, true, false, nil, typeString, nil),
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
				},
			},
			filtersInput: map[string]interface{}{
				dataSourceFilterPropertyName: []map[string]interface{}{
					newFilter("label", []string{"my_label"}),
					newFilter("not_primitive", []string{"filters for non primitive properties are not supported at the moment"}),
				},
			},
			expectedFilters: nil,
			expectedError:   errors.New("property not supported as as filter: not_primitive"),
		},
		{
			name: "data source populated with an incorrect filter containing multiple values for a primitive property",
			specSchemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
				},
			},
			filtersInput: map[string]interface{}{
				dataSourceFilterPropertyName: []map[string]interface{}{
					newFilter("label", []string{"value1", "value2"}),
				},
			},
			expectedFilters: nil,
			expectedError:   errors.New("filters for primitive properties can not have more than one value in the values field"),
		},
	}

	for _, tc := range testCases {
		// Given
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: tc.specSchemaDefinition,
			},
		}
		resourceSchema := dataSourceFactory.createTerraformDataSourceSchema()
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, tc.filtersInput)
		// When
		filters, err := dataSourceFactory.validateInput(resourceLocalData)
		// Then
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
		for _, expectedFilter := range tc.expectedFilters {
			assertFilter(t, filters, expectedFilter, tc.name)
		}
	}
}

func TestFilterMatch(t *testing.T) {
	testCases := []struct {
		name                           string
		specSchemaDefinitionProperties specSchemaDefinitionProperties
		filters                        filters
		payloadItem                    map[string]interface{}
		expectedResult                 bool
		expectedError                  error
		resourceSchemaErr              error
	}{
		{
			name: "happy path - payloadItem matches the filter",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
			},
			filters: filters{
				filter{"label", "some label"},
			},
			payloadItem: map[string]interface{}{
				"label": "some label",
			},
			expectedResult:    true,
			expectedError:     nil,
			resourceSchemaErr: nil,
		},
		{
			name: "crappy path - invalid specSchemaDefinition",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
			},
			filters: filters{
				filter{"label", "some label"},
			},
			payloadItem: map[string]interface{}{
				"label": "some label",
			},
			expectedResult:    false,
			expectedError:     errors.New("invalid specSchemaDefinition"),
			resourceSchemaErr: errors.New("invalid specSchemaDefinition"),
		},
		{
			name: "crappy path - payloadItem doesn't match the filter name",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
			},
			filters: filters{
				filter{"invalid filter name", "some label"},
			},
			payloadItem: map[string]interface{}{
				"label": "some label",
			},
			expectedResult:    false,
			expectedError:     nil,
			resourceSchemaErr: nil,
		},
		{
			name: "crappy path - payloadItem doesn't match the filter value",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
			},
			filters: filters{
				filter{"label", "invalid filter value"},
			},
			payloadItem: map[string]interface{}{
				"label": "some label",
			},
			expectedResult:    false,
			expectedError:     nil,
			resourceSchemaErr: nil,
		},
	}

	for _, tc := range testCases {
		// Given
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: tc.specSchemaDefinitionProperties,
				},
				error: tc.resourceSchemaErr,
			},
		}
		// When
		match, err := dataSourceFactory.filterMatch(tc.filters, tc.payloadItem)
		// Then
		assert.Equal(t, tc.expectedResult, match, tc.name)
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func assertFilter(t *testing.T, filters filters, expectedFilter filter, msgAndArgs ...interface{}) bool {
	for _, f := range filters {
		if f.name == expectedFilter.name {
			assert.Equal(t, expectedFilter.value, f.value, msgAndArgs)
		}
	}
	return false
}

func newFilter(name string, values []string) map[string]interface{} {
	return map[string]interface{}{
		dataSourceFilterSchemaNamePropertyName:   name,
		dataSourceFilterSchemaValuesPropertyName: values,
	}
}
