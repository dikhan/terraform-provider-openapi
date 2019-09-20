package openapi

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestCreateTerraformDataSource(t *testing.T) {
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
		dataSourceFactory := dataSourceFactory{
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

		dataSource, err := dataSourceFactory.createTerraformDataSource()

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

func TestCreateTerraformDataSourceSchema(t *testing.T) {
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
		dataSourceFactory := dataSourceFactory{
			openAPIResource: tc.openAPIResource,
		}
		s, err := dataSourceFactory.createTerraformDataSourceSchema()
		if tc.expectedError == nil {
			assert.NotNil(t, s, tc.name)
			// data source specific properties for filtering purposes (exposed to the user to be able to provide filters)
			assert.Contains(t, s, dataSourceFilterPropertyName, tc.name)
			assert.IsType(t, &schema.Resource{}, s[dataSourceFilterPropertyName].Elem, tc.name)
			assert.Contains(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema, dataSourceFilterSchemaNamePropertyName, tc.name)
			assert.Equal(t, schema.TypeString, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaNamePropertyName].Type, tc.name)
			assert.True(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaNamePropertyName].Required, tc.name)
			assert.Contains(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema, dataSourceFilterSchemaValuesPropertyName, tc.name)
			assert.Equal(t, schema.TypeList, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaValuesPropertyName].Type, tc.name)
			assert.True(t, s[dataSourceFilterPropertyName].Elem.(*schema.Resource).Schema[dataSourceFilterSchemaValuesPropertyName].Required, tc.name)

			// resource specific properties as per swagger def (this properties are meant to be populated by the read operation when a match is found as per the filters)
			assert.Nil(t, s["id"], tc.name) // we assert that s["id"] is Nil because during the creation of the schema id is treated in a special way and should not be populated at creation time (must be set in read() method)
			assert.Contains(t, s, "label", tc.name)
			assert.True(t, s["label"].Computed, tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func TestDataSourceRead(t *testing.T) {
	// Given
	dataSourceFactory := dataSourceFactory{
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

	testCases := []struct {
		name            string
		filtersInput    []map[string]interface{}
		responsePayload []map[string]interface{}
		expectedResult  map[string]interface{}
		expectedError   error
	}{
		// TODO: add a test to cover sub-resource use case too (need to cover this before releasing data source support)
		{
			name: "fetch selected data source as per filter configuration (label=someLabel) when the filter is into a nested object",
			filtersInput: []map[string]interface{}{
				newFilter("label", []string{"someLabel"}),
			},
			responsePayload: []map[string]interface{}{
				{
					"id":     "someID",
					"label":  "someLabel",
					"owners": []string{"someOwner"},
				},
				{
					"id":     "someOtherID",
					"label":  "someOtherLabel",
					"owners": []string{},
				},
			},
			expectedError: nil,
		},
		//{
		//	name: "fetch selected data source as per filter configuration (label=someLabel) when the filter a top level property of a nested object",
		//	filtersInput: []map[string]interface{}{
		//		newFilter("label", []string{"someLabel"}),
		//	},
		//	responsePayload: []map[string]interface{}{
		//		{
		//			"id":     "someID",
		//			"label":  "someLabel",
		//			"owners": []string{"someOwner"},
		//		},
		//		{
		//			"id":     "someOtherID",
		//			"label":  "someOtherLabel",
		//			"owners": []string{},
		//		},
		//	},
		//	expectedError: nil,
		//},
		//{
		//	name: "fetch selected data source as per filter configuration (label=someLabel)",
		//	filtersInput: []map[string]interface{}{
		//		newFilter("label", []string{"someLabel"}),
		//	},
		//	responsePayload: []map[string]interface{}{
		//		{
		//			"id":     "someID",
		//			"label":  "someLabel",
		//			"owners": []string{"someOwner"},
		//		},
		//		{
		//			"id":     "someOtherID",
		//			"label":  "someOtherLabel",
		//			"owners": []string{},
		//		},
		//	},
		//	expectedError: nil,
		//},
		//{
		//	name: "no filter match",
		//	filtersInput: []map[string]interface{}{
		//		newFilter("label", []string{"some non existing label"}),
		//	},
		//	responsePayload: []map[string]interface{}{
		//		{
		//			"id":    "someID",
		//			"label": "someLabel",
		//		},
		//	},
		//	expectedError: errors.New("your query returned no results. Please change your search criteria and try again"),
		//},
		//{
		//	name: "after filtering the result contains more than one element",
		//	filtersInput: []map[string]interface{}{
		//		newFilter("label", []string{"my_label"}),
		//	},
		//	responsePayload: []map[string]interface{}{
		//		{
		//			"id":    "someID",
		//			"label": "my_label",
		//		},
		//		{
		//			"id":    "someOtherID",
		//			"label": "my_label",
		//		},
		//	},
		//	expectedError: errors.New("your query returned contains more than one result. Please change your search criteria to make it more specific"),
		//},
		//{
		//	name: "validate input fails",
		//	filtersInput: []map[string]interface{}{
		//		newFilter("non_existing_property", []string{"my_label"}),
		//	},
		//	responsePayload: []map[string]interface{}{},
		//	expectedError:   errors.New("filter name does not match any of the schema properties: property with name 'non_existing_property' not existing in resource schema definition"),
		//},
	}

	for _, tc := range testCases {
		resourceSchema, err := dataSourceFactory.createTerraformDataSourceSchema()
		require.NoError(t, err)

		filtersInput := map[string]interface{}{
			dataSourceFilterPropertyName: tc.filtersInput,
		}
		resourceData := schema.TestResourceDataRaw(t, resourceSchema, filtersInput)
		client := &clientOpenAPIStub{
			responseListPayload: tc.responsePayload,
		}
		// When
		err = dataSourceFactory.read(resourceData, client)
		// Then
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
			// assert that the filtered data source contains the same values as the ones returned by the API
			assert.Equal(t, 8, len(resourceData.State().Attributes), tc.name)                //this asserts that ONLY 1 element is returned when the filter is applied (2 prop of the elelemnt + 4 prop given by the filter)
			assert.Equal(t, client.responseListPayload[0]["id"], resourceData.Id(), tc.name) //resourceData.Id() is being called instead of resourceData.Get("id") because id property is a special one kept by Terraform
			assert.Equal(t, client.responseListPayload[0]["label"], resourceData.Get("label"), tc.name)
			expectedOwners := client.responseListPayload[0]["owners"].([]string)
			owners := resourceData.Get("owners").([]interface{})
			assert.NotNil(t, owners, tc.name)
			assert.NotNil(t, len(expectedOwners), len(owners), tc.name)
			assert.Equal(t, expectedOwners[0], owners[0], tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func TestDataSourceRead_ForNestedObjects(t *testing.T) {
	nestedObjectSchemaDefinition := &specSchemaDefinition{
		Properties: specSchemaDefinitionProperties{
			newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
			newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
		},
	}
	nestedObjectDefault := map[string]interface{}{
		"origin_port": nestedObjectSchemaDefinition.Properties[0].Default,
		"protocol":    nestedObjectSchemaDefinition.Properties[1].Default,
	}
	nestedObject := newObjectSchemaDefinitionPropertyWithDefaults("nested_object", "", true, false, false, nestedObjectDefault, nestedObjectSchemaDefinition)
	propertyWithNestedObjectSchemaDefinition := &specSchemaDefinition{
		Properties: specSchemaDefinitionProperties{
			idProperty,
			nestedObject,
		},
	}
	dataValue := map[string]interface{}{
		"id":            propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
		"nested_object": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
	}

	dataSourceSchema := newObjectSchemaDefinitionPropertyWithDefaults("nested-oobj", "", true, false, false, dataValue, propertyWithNestedObjectSchemaDefinition)

	// Given
	dataSourceFactory := dataSourceFactory{
		openAPIResource: &specStubResource{
			schemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{dataSourceSchema},
			},
		},
	}

	fmt.Println(dataSourceFactory) //todo work on this
	fs, err := dataSourceFactory.createTerraformDataSourceSchema()
	fmt.Println(err, fs)

}

func TestDataSourceRead_Fails_Because_Cannot_extract_ParentsID(t *testing.T) {
	err := dataSourceFactory{}.read(nil, &clientOpenAPIStub{})
	assert.EqualError(t, err, "can't get parent ids from a resourceFactory with no openAPIResource")
}

func TestDataSourceRead_Fails_Because_List_Operation_Returns_Err(t *testing.T) {
	dataSourceFactory := dataSourceFactory{
		openAPIResource: &specStubResource{
			schemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, false, nil),
				},
			},
		},
	}
	resourceSchema, err := dataSourceFactory.createTerraformDataSourceSchema()
	require.NoError(t, err)

	filtersInput := map[string]interface{}{
		dataSourceFilterPropertyName: []map[string]interface{}{
			newFilter("label", []string{"someLabel"}),
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceSchema, filtersInput)
	client := &clientOpenAPIStub{
		responseListPayload: []map[string]interface{}{
			{
				"label": "someLabel",
			},
		},
		error: errors.New("some error"),
	}
	err = dataSourceFactory.read(resourceData, client)
	assert.EqualError(t, err, "some error")
}

func TestDataSourceRead_Fails_Because_Bad_Status_Code(t *testing.T) {
	// Given
	dataSourceFactory := dataSourceFactory{
		openAPIResource: &specStubResource{
			name: "some resource",
			schemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, false, nil),
				},
			},
		},
	}
	resourceSchema, err := dataSourceFactory.createTerraformDataSourceSchema()
	require.NoError(t, err)

	filtersInput := map[string]interface{}{
		dataSourceFilterPropertyName: []map[string]interface{}{
			newFilter("label", []string{"someLabel"}),
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceSchema, filtersInput)
	client := &clientOpenAPIStub{
		returnHTTPCode: 400,
	}
	// When
	err = dataSourceFactory.read(resourceData, client)
	// Then
	assert.Equal(t, errors.New("[data source='some resource'] GET  failed: [resource='some resource'] HTTP Response Status Code 400 not matching expected one [200] ()"), err)
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
		resourceSchema, err := dataSourceFactory.createTerraformDataSourceSchema()
		require.NoError(t, err)

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
			expectedResult: true,
			expectedError:  nil,
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
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "happy path - payloadItem matches the filter for float property WHEN FLOAT HAS A DECIMAL PART == 0 ",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newNumberSchemaDefinitionPropertyWithDefaults("float property name", "", false, true, nil),
			},
			filters: filters{
				filter{"float property name", "6.0"},
			},
			payloadItem: map[string]interface{}{
				"float property name": 6.0, //because 6.0 is treateted as an interface golang keeps only the int part (6) so we need to treat thi case specially
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "happy path - payloadItem matches the filter for float property WHEN FLOAT HAS A DECIMAL PART != 0 ",
			specSchemaDefinitionProperties: specSchemaDefinitionProperties{
				newNumberSchemaDefinitionPropertyWithDefaults("float property name", "", false, true, nil),
			},
			filters: filters{
				filter{"float property name", "6.89"},
			},
			payloadItem: map[string]interface{}{
				"float property name": 6.89,
			},
			expectedResult: true,
			expectedError:  nil,
		},
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
			expectedResult: true,
			expectedError:  nil,
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
			expectedResult: false,
			expectedError:  nil,
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
			expectedResult: false,
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		// Given
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: tc.specSchemaDefinitionProperties,
				},
			},
		}
		// When
		match := dataSourceFactory.filterMatch(tc.filters, tc.payloadItem)
		// Then
		assert.Equal(t, tc.expectedResult, match, tc.name)
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
