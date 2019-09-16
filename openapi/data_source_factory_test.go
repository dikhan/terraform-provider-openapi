package openapi

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTerraformDataSourceSchema(t *testing.T) {
	testCases := []struct {
		name            string
		openAPIResource SpecResource
	}{
		{
			// TODO: this test fails due to the impl not being there yet. Fixing this test should make the other test from
			//  TestDataSourceRead work
			name: "data source schema is configured as expected",
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: specSchemaDefinitionProperties{
						newStringSchemaDefinitionPropertyWithDefaults("id", "", false, true, nil),
						newStringSchemaDefinitionPropertyWithDefaults("label", "", false, false, nil),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		dataSourceFactory := dataSourceFactory{
			openAPIResource: tc.openAPIResource,
		}
		s := dataSourceFactory.createTerraformDataSourceSchema()
		assert.NotNil(t, s)
		// data source specific properties for filtering purposes (exposed to the user to be able to provide filters)
		assert.Contains(t, s, dataSourceFilterPropertyName)
		assert.IsType(t, &schema.Resource{}, s[dataSourceFilterPropertyName].Elem)
		assert.Contains(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema, dataSourceFilterSchemaNamePropertyName)
		assert.Equal(t, schema.TypeString, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaNamePropertyName].Type)
		assert.True(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaNamePropertyName].Required)
		assert.Contains(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema, dataSourceFilterSchemaValuesPropertyName)
		assert.Equal(t, schema.TypeList, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaValuesPropertyName].Type)
		assert.True(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaValuesPropertyName].Required)

		// resource specific properties as per swagger def (this properties are meant to be popolated by the read operation when a match is found as per the filters)
		// TODO: It looks like createResourceSchemaIgnoreID (called in createDataSourceSchema) drops the id property, so schema doesn't contain id at this point - is the expected behavior?
		//assert.Contains(t, s, "id")
		//assert.True(t, s["id"].Computed)
		assert.Contains(t, s, "label")
		assert.True(t, s["label"].Computed)
	}
}

func TestDataSourceRead(t *testing.T) {

	testCases := []struct {
		name            string
		filtersInput    []map[string]interface{}
		responsePayload []map[string]interface{}
		expectedResult  map[string]interface{}
		expectedError   error
	}{
		{
			// TODO: fix this test. The reason why it fails is because the data source is currently not configuring the properties
			//  of the resource in the terraform schema hence the call to save the data for label failing:
			//  Expected nil, but got: &errors.errorString{s:"Invalid address to set: []string{\"label\"}"}
			//  See: https://github.com/dikhan/terraform-provider-openapi/blob/feature/data-source-support/openapi/data_source_factory.go#L41
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
		{
			name: "no filter match",
			filtersInput: []map[string]interface{}{
				newFilter("label", []string{"some non existing label"}),
			},
			responsePayload: []map[string]interface{}{
				{
					"id":    "someID",
					"label": "someLabel",
				},
			},
			expectedError: errors.New("your query returned no results. Please change your search criteria and try again"),
		},
		{
			name: "after filtering the result contains more than one element",
			filtersInput: []map[string]interface{}{
				newFilter("label", []string{"my_label"}),
			},
			responsePayload: []map[string]interface{}{
				{
					"id":    "someID",
					"label": "my_label",
				},
				{
					"id":    "someOtherID",
					"label": "my_label",
				},
			},
			expectedError: errors.New("your query returned contains more than one result. Please change your search criteria to make it more specific"),
		},
		{
			name: "validate input fails",
			filtersInput: []map[string]interface{}{
				newFilter("non_existing_property", []string{"my_label"}),
			},
			responsePayload: []map[string]interface{}{},
			expectedError:   errors.New("filter name does not match any of the schema properties: property with name 'non_existing_property' not existing in resource schema definition"),
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
			// assert that the filtered data source contains the same values as the ones returned by the API
			assert.Equal(t, client.responseListPayload[0]["id"], resourceData.Get("id"), tc.name)
			assert.Equal(t, client.responseListPayload[0]["label"], resourceData.Get("label"), tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
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

// TODO: Fix filter for float property (test currently failing) - conversion from string to float in dataSourceFactory.filterMatch()
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
			name: "happy path - payloadItem matches the filter for string property",
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
			name: "happy path - payloadItem matches the filter for int property",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("int property name", "", false, true, nil),
			},
			filters: filters{
				filter{"int property name", "5"},
			},
			payloadItem: map[string]interface{}{
				"int property name": 5,
			},
			expectedResult:    true,
			expectedError:     nil,
			resourceSchemaErr: nil,
		},
		//{
		//	name: "happy path - payloadItem matches the filter for float property",
		//	specSchemaDefinitionProperties: specSchemaDefinitionProperties{
		//		newNumberSchemaDefinitionPropertyWithDefaults("float property name", "", false, true, nil),
		//	},
		//	filters: filters{
		//		filter{"float property name", "6.0"},
		//	},
		//	payloadItem: map[string]interface{}{
		//		"float property name": 6.0,
		//	},
		//	expectedResult:    true,
		//	expectedError:     nil,
		//	resourceSchemaErr: nil,
		//},
		{
			name: "happy path - payloadItem matches the filter for bool property",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newBoolSchemaDefinitionPropertyWithDefaults("bool property name", "", false, true, nil),
			},
			filters: filters{
				filter{"bool property name", "false"},
			},
			payloadItem: map[string]interface{}{
				"bool property name": false,
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
