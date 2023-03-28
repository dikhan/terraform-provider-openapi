package openapi

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// testSchemaDefinition defines a test schema that contains a list of properties.
type testSchemaDefinition []*SpecSchemaDefinitionProperty

// primitive testing vars
var idProperty = newStringSchemaDefinitionProperty("id", "", true, false, false, false, false, true, false, false, "id")
var stringProperty = newStringSchemaDefinitionPropertyWithDefaults("string_property", "", true, false, "updatedValue")
var intProperty = newIntSchemaDefinitionPropertyWithDefaults("int_property", "", true, false, 12)
var numberProperty = newNumberSchemaDefinitionPropertyWithDefaults("number_property", "", true, false, 13.99)
var boolProperty = newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", true, false, true)
var slicePrimitiveProperty = newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, false, []interface{}{"value1"}, TypeString, nil)

// testing properties with special configuration
var stringWithPreferredNameProperty = newStringSchemaDefinitionPropertyWithDefaults("stringProperty", "string_preferred_property", true, false, "updatedValue")
var someIdentifierProperty = newStringSchemaDefinitionProperty("somePropertyThatShouldBeUsedAsID", "", true, true, false, false, false, false, true, false, "idValue")
var immutableProperty = newStringSchemaDefinitionProperty("string_immutable_property", "", true, false, false, false, false, true, false, false, "updatedImmutableValue")
var computedProperty = newStringSchemaDefinitionPropertyWithDefaults("computed_property", "", false, true, nil)
var readOnlyProperty = newStringSchemaDefinitionPropertyWithDefaults("read_only_property", "", false, true, "some_value")
var optionalProperty = newStringSchemaDefinitionPropertyWithDefaults("optional_property", "", false, false, "updatedValue")
var sensitiveProperty = newStringSchemaDefinitionProperty("sensitive_property", "", false, false, false, false, true, false, false, false, "sensitive")
var forceNewProperty = newBoolSchemaDefinitionProperty("bool_force_new_property", "", true, false, false, true, false, false, false, false, true)
var statusProperty = newStringSchemaDefinitionPropertyWithDefaults("status", "", false, true, "pending")

// testing properties with zero values set
var stringZeroValueProperty = newStringSchemaDefinitionPropertyWithDefaults("string_property", "", true, false, "")
var intZeroValueProperty = newIntSchemaDefinitionPropertyWithDefaults("int_property", "", true, false, 0)
var numberZeroValueProperty = newNumberSchemaDefinitionPropertyWithDefaults("number_property", "", true, false, 0)
var boolZeroValueProperty = newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", true, false, false)
var sliceZeroValueProperty = newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, false, []interface{}{""}, TypeString, nil)

func setSchemaDefinitionPropertyWriteOnly(propertySchemaDefinition *SpecSchemaDefinitionProperty) *SpecSchemaDefinitionProperty {
	propertySchemaDefinition.WriteOnly = true
	return propertySchemaDefinition
}

func newStringSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newStringSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, false, defaultValue)
}

func newParentStringSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	p := newStringSchemaDefinitionPropertyWithDefaults(name, preferredName, required, readOnly, defaultValue)
	p.IsParentProperty = true
	return p
}

func newStringSchemaDefinitionProperty(name, preferredName string, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, TypeString, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newIntSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newIntSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, false, defaultValue)
}

func newIntSchemaDefinitionProperty(name, preferredName string, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, TypeInt, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newNumberSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newNumberSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, false, defaultValue)
}

func newNumberSchemaDefinitionProperty(name, preferredName string, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, TypeFloat, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newBoolSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newBoolSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, false, false, defaultValue)
}

func newBoolSchemaDefinitionProperty(name, preferredName string, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, TypeBool, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
}

func newObjectSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly, computed bool, defaultValue interface{}, objectSpecSchemaDefinition *SpecSchemaDefinition) *SpecSchemaDefinitionProperty {
	return newObjectSchemaDefinitionProperty(name, preferredName, required, readOnly, computed, false, false, false, false, false, defaultValue, objectSpecSchemaDefinition)
}

func newObjectSchemaDefinitionProperty(name, preferredName string, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}, objectSpecSchemaDefinition *SpecSchemaDefinition) *SpecSchemaDefinitionProperty {
	schemaDefProperty := newSchemaDefinitionProperty(name, preferredName, TypeObject, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
	schemaDefProperty.SpecSchemaDefinition = objectSpecSchemaDefinition
	return schemaDefProperty
}

func newListSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly, computed bool, defaultValue interface{}, itemsType schemaDefinitionPropertyType, objectSpecSchemaDefinition *SpecSchemaDefinition) *SpecSchemaDefinitionProperty {
	return newListSchemaDefinitionProperty(name, preferredName, required, readOnly, computed, false, false, false, false, false, defaultValue, itemsType, objectSpecSchemaDefinition)
}

func newListSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, computed, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}, itemsType schemaDefinitionPropertyType, objectSpecSchemaDefinition *SpecSchemaDefinition) *SpecSchemaDefinitionProperty {
	schemaDefProperty := newSchemaDefinitionProperty(name, preferredName, TypeList, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier, defaultValue)
	schemaDefProperty.ArrayItemsType = itemsType
	schemaDefProperty.SpecSchemaDefinition = objectSpecSchemaDefinition
	return schemaDefProperty
}

func newSchemaDefinitionProperty(name, preferredName string, propertyType schemaDefinitionPropertyType, required, readOnly, computed, forceNew, sensitive, immutable, isIdentifier, isStatusIdentifier bool, defaultValue interface{}) *SpecSchemaDefinitionProperty {
	return &SpecSchemaDefinitionProperty{
		Name:               name,
		Type:               propertyType,
		PreferredName:      preferredName,
		Required:           required,
		ReadOnly:           readOnly,
		Computed:           computed,
		ForceNew:           forceNew,
		Sensitive:          sensitive,
		Immutable:          immutable,
		IsIdentifier:       isIdentifier,
		IsStatusIdentifier: isStatusIdentifier,
		Default:            defaultValue,
	}
}

func newTestSchema(schemaDefinitionProperty ...*SpecSchemaDefinitionProperty) *testSchemaDefinition {
	testSchema := &testSchemaDefinition{}
	for _, s := range schemaDefinitionProperty {
		*testSchema = append(*testSchema, s)
	}
	return testSchema
}

func (s *testSchemaDefinition) getSchemaDefinition() *SpecSchemaDefinition {
	schemaDefinitionProperties := SpecSchemaDefinitionProperties{}
	for _, schemaProperty := range *s {
		schemaDefinitionProperties = append(schemaDefinitionProperties, schemaProperty)
	}
	return &SpecSchemaDefinition{
		Properties: schemaDefinitionProperties,
	}
}

// getResourceData creates a ResourceData object from the testSchemaDefinition. The key values honor's Terraform key constraints
// where field names must be snake_case
func (s *testSchemaDefinition) getResourceData(t *testing.T) *schema.ResourceData {
	resourceSchema := map[string]*schema.Schema{}
	resourceDataMap := map[string]interface{}{}
	for _, schemaProperty := range *s {
		terraformName := schemaProperty.GetTerraformCompliantPropertyName()
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

func initAPISpecFile(swaggerContent string) *os.File {
	file, err := ioutil.TempFile("", "testSpec")
	if err != nil {
		log.Fatal(err)
	}
	swagger := json.RawMessage([]byte(swaggerContent))
	_, err = file.Write(swagger)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func assertDataSourceSchemaProperty(t *testing.T, actual *schema.Schema, expectedType schema.ValueType, msgAndArgs ...interface{}) {
	assertTerraformSchemaProperty(t, actual, expectedType, false, true)
}

func assertTerraformSchemaNestedObjectProperty(t *testing.T, actual *schema.Schema, expectedRequired, expectedComputed bool, msgAndArgs ...interface{}) {
	assertTerraformSchemaProperty(t, actual, schema.TypeList, expectedRequired, expectedComputed)
	assert.Equal(t, 1, actual.MaxItems, msgAndArgs)
}

func assertTerraformSchemaProperty(t *testing.T, actual *schema.Schema, expectedType schema.ValueType, expectedRequired, expectedComputed bool, msgAndArgs ...interface{}) {
	assert.NotNil(t, actual, msgAndArgs)
	assert.Equal(t, expectedType, actual.Type, msgAndArgs)
	if expectedRequired {
		assert.True(t, actual.Required, msgAndArgs)
		assert.False(t, actual.Optional, msgAndArgs)
	} else {
		assert.True(t, actual.Optional, msgAndArgs)
		assert.False(t, actual.Required, msgAndArgs)
	}
	assert.Equal(t, expectedComputed, actual.Computed, msgAndArgs)
}
