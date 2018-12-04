package openapi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"testing"
)

// testSchemaDefinition defines a test schema that contains a list of properties.
type testSchemaDefinition []*specSchemaDefinitionProperty

// primitive testing vars
var idProperty = newStringSchemaDefinitionProperty("id", "", true, false, false, false, true, false, false, "id")
var stringProperty = newStringSchemaDefinitionPropertyWithDefaults("string_property", "", true, false, "updatedValue")
var intProperty = newIntSchemaDefinitionPropertyWithDefaults("int_property", "", true, false, 12)
var numberProperty = newNumberSchemaDefinitionPropertyWithDefaults("number_property", "", true, false, 13.99)
var boolProperty = newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", true, false, true)
var slicePrimitiveProperty = newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, []string{"value1"}, typeString, nil)

// testing properties with special configuration
var stringWithPreferredNameProperty = newStringSchemaDefinitionPropertyWithDefaults("stringProperty", "string_preferred_property", true, false, "updatedValue")
var nonImmutableProperty = newStringSchemaDefinitionPropertyWithDefaults("other_string_property", "", true, false, "newValue")
var someIdentifierProperty = newStringSchemaDefinitionProperty("somePropertyThatShouldBeUsedAsID", "", true, true, false, false, false, true, false, "idValue")
var immutableProperty = newStringSchemaDefinitionProperty("string_immutable_property", "", true, false, false, false, true, false, false, "updatedImmutableValue")
var computedProperty = newStringSchemaDefinitionPropertyWithDefaults("computed_property", "", true, true, nil)
var optionalProperty = newStringSchemaDefinitionPropertyWithDefaults("optional_property", "", false, false, "updatedValue")
var sensitiveProperty = newStringSchemaDefinitionProperty("sensitive_property", "", false, false, false, true, false, false, false, "sensitive")
var forceNewProperty = newBoolSchemaDefinitionProperty("bool_force_new_property", "", true, false, true, false, false, false, false, true)
var statusProperty = newStringSchemaDefinitionPropertyWithDefaults("status", "", false, true, "pending")

// testing properties with zero values set
var intZeroValueProperty = newIntSchemaDefinitionPropertyWithDefaults("int_property", "", true, false, 0)
var numberZeroValueProperty = newNumberSchemaDefinitionPropertyWithDefaults("number_property", "", true, false, 0)
var boolZeroValueProperty = newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", true, false, false)
var sliceZeroValueProperty = newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, []string{""}, typeString, nil)

func newStringSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newStringSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, defaultValue)
}

func newStringSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeString, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newIntSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newIntSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, defaultValue)
}

func newIntSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeInt, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newNumberSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newNumberSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, defaultValue)
}

func newNumberSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeFloat, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newBoolSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newBoolSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, defaultValue)
}

func newBoolSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeBool, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newObjectSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}, objectSpecSchemaDefinition *specSchemaDefinition) *specSchemaDefinitionProperty {
	return newObjectSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, defaultValue, objectSpecSchemaDefinition)
}

func newObjectSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}, objectSpecSchemaDefinition *specSchemaDefinition) *specSchemaDefinitionProperty {
	schemaDefProperty := newSchemaDefinitionProperty(name, preferredName, typeObject, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
	schemaDefProperty.SpecSchemaDefinition = objectSpecSchemaDefinition
	return schemaDefProperty
}

func newListSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}, itemsType schemaDefinitionPropertyType, objectSpecSchemaDefinition *specSchemaDefinition) *specSchemaDefinitionProperty {
	return newListSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, defaultValue, itemsType, objectSpecSchemaDefinition)
}

func newListSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}, itemsType schemaDefinitionPropertyType, objectSpecSchemaDefinition *specSchemaDefinition) *specSchemaDefinitionProperty {
	schemaDefProperty := newSchemaDefinitionProperty(name, preferredName, typeList, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
	schemaDefProperty.ArrayItemsType = itemsType
	schemaDefProperty.SpecSchemaDefinition = objectSpecSchemaDefinition
	return schemaDefProperty
}

func newSchemaDefinitionProperty(name, preferredName string, propertyType schemaDefinitionPropertyType, required, readOnly, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return &specSchemaDefinitionProperty{
		Name:               name,
		Type:               propertyType,
		PreferredName:      preferredName,
		Required:           required,
		ReadOnly:           readOnly,
		ForceNew:           forceNew,
		Sensitive:          sensitive,
		Immutable:          immutable,
		IsIdentifier:       isIdentifier,
		IsStatusIdentifier: isStatusIdentifier,
		Default:            defaultValue,
	}
}

func newTestSchema(schemaDefinitionProperty ...*specSchemaDefinitionProperty) *testSchemaDefinition {
	testSchema := &testSchemaDefinition{}
	for _, s := range schemaDefinitionProperty {
		*testSchema = append(*testSchema, s)
	}
	return testSchema
}

func (s *testSchemaDefinition) getSchemaDefinition() *specSchemaDefinition {
	schemaDefinitionProperties := specSchemaDefinitionProperties{}
	for _, schemaProperty := range *s {
		schemaDefinitionProperties = append(schemaDefinitionProperties, schemaProperty)
	}
	return &specSchemaDefinition{
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
		schema, err := schemaProperty.terraformSchema()
		if err != nil {
			log.Fatal(err)
		}
		resourceSchema[terraformName] = schema
		resourceDataMap[terraformName] = schemaProperty.Default
	}
	resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	return resourceLocalData
}
