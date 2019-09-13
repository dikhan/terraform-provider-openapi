package openapi

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateInput(t *testing.T) {

	testCases := []struct {
		name                 string
		specSchemaDefinition *specSchemaDefinition
		filtersInput         map[string]interface{}
		expectedError        error
	}{
		{
			name: "data source populated with a correct filters",
			specSchemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("owner", "", false, true, nil),
					newStringSchemaDefinitionPropertyWithDefaults("label", "", false, true, nil),
				},
			},
			filtersInput: map[string]interface{}{
				dataSourceFilterPropertyName: []map[string]interface{}{
					newFilter("owner", []string{"some_owner"}),
					newFilter("label", []string{"label_to_fetch"}),
				},
			},
			expectedError: nil,
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
			expectedError: errors.New("filter name does not match any of the schema properties: property with name 'non_matching_property_name' not existing in resource schema definition"),
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
			expectedError: errors.New("property not supported as as filter: not_primitive"),
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
			expectedError: errors.New("filters for primitive properties can not have more than one value in the values field"),
		},
	}

	for _, tc := range testCases {
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: tc.specSchemaDefinition,
			},
		}
		resourceSchema := dataSourceFactory.createTerraformDataSourceSchema()
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, tc.filtersInput)
		err := dataSourceFactory.validateInput(resourceLocalData)
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func newFilter(name string, values []string) map[string]interface{} {
	return map[string]interface{}{
		dataSourceFilterSchemaNamePropertyName:   name,
		dataSourceFilterSchemaValuesPropertyName: values,
	}
}
