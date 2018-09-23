package openapi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"testing"
)

// testSchemaDefinition defines a test schema that contains a list of properties.
type testSchemaDefinition []*SchemaDefinitionProperty

var idProperty = newStringSchemaDefinitionProperty("id", "", true, false, false, false, true, false, "id")
var someIdentifierProperty = newStringSchemaDefinitionProperty("somePropertyThatShouldBeUsedAsID", "", true, true, false, false, false, true, "idValue")
var stringProperty = newStringSchemaDefinitionPropertyWithDefaults("string_property", "", true, false, "updatedValue")
var stringWithPreferredNameProperty = newStringSchemaDefinitionPropertyWithDefaults("stringProperty", "string_property", true, false, "updatedValue")
var immutableProperty = newStringSchemaDefinitionProperty("string_immutable_property", "", true, false, false, false, true, false, "updatedImmutableValue")
var nonImmutableProperty = newStringSchemaDefinitionPropertyWithDefaults("other_string_property", "", true, false, "newValue")
var computedProperty = newStringSchemaDefinitionPropertyWithDefaults("computed_property", "", true, true, nil)
var intProperty = newIntSchemaDefinitionPropertyWithDefaults("int_property", "", true, false, 12)
var numberProperty = newNumberSchemaDefinitionPropertyWithDefaults("number_property", "", true, false, 13.99)
var boolProperty = newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", true, false, true)
var sliceProperty = newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, []string{"value1"})

var intZeroValueProperty = newIntSchemaDefinitionPropertyWithDefaults("int_property", "", true, false, 0)
var numberZeroValueProperty = newNumberSchemaDefinitionPropertyWithDefaults("number_property", "", true, false, 0)
var boolZeroValueProperty = newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", true, false, false)
var sliceZeroValueProperty = newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, []string{""})

func newTestSchema(schemaDefinitionProperty ...*SchemaDefinitionProperty) *testSchemaDefinition {
	testSchema := &testSchemaDefinition{}
	for _, s := range schemaDefinitionProperty {
		*testSchema = append(*testSchema, s)
	}
	return testSchema
}

func (s *testSchemaDefinition) getSchemaDefinition() *SchemaDefinition {
	schemaDefinitionProperties := SchemaDefinitionProperties{}
	for _, schemaProperty := range *s {
		schemaDefinitionProperties[schemaProperty.Name] = schemaProperty
	}
	return &SchemaDefinition{
		Properties: schemaDefinitionProperties,
	}
}

// getResourceData creates a ResourceData object from the testSchemaDefinition. The key values honor's Terraform key constraints
// where field names must be snake_case
func (s *testSchemaDefinition) getResourceData(t *testing.T) *schema.ResourceData {
	resourceSchema := map[string]*schema.Schema{}
	resourceDataMap := map[string]interface{}{}
	for _, schemaProperty := range *s {
		terraformName := schemaProperty.getTerraformCompliantPropertyName()
		resourceSchema[terraformName] = schemaProperty.terraformSchema()
		resourceDataMap[terraformName] = schemaProperty.Default
	}
	resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	return resourceLocalData
}
