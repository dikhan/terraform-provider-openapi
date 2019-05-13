package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/hashicorp/terraform/helper/schema"
)

// schemaDefinitionPropertyType defines the type of a property
type schemaDefinitionPropertyType string

const (
	typeString schemaDefinitionPropertyType = "string"
	typeInt    schemaDefinitionPropertyType = "integer"
	typeFloat  schemaDefinitionPropertyType = "number"
	typeBool   schemaDefinitionPropertyType = "boolean"
	typeList   schemaDefinitionPropertyType = "list"
	typeObject schemaDefinitionPropertyType = "object"
)

const idDefaultPropertyName = "id"
const statusDefaultPropertyName = "status"

// specSchemaDefinitionProperty defines the attributes for a schema property
type specSchemaDefinitionProperty struct {
	Name               string
	PreferredName      string
	Type               schemaDefinitionPropertyType
	ArrayItemsType     schemaDefinitionPropertyType
	OptionalComputed   bool
	Required           bool
	ReadOnly           bool
	ForceNew           bool
	Sensitive          bool
	Immutable          bool
	IsIdentifier       bool
	IsStatusIdentifier bool
	// Default field is only for informative purposes to know what the openapi spec for the property stated the default value is
	// As per the openapi spec default attributes, the value is expected to be computed by the API
	Default interface{}
	// only for object type properties or arrays type properties with array items of type object
	SpecSchemaDefinition *specSchemaDefinition
}

func (s *specSchemaDefinitionProperty) getTerraformCompliantPropertyName() string {
	if s.PreferredName != "" {
		return terraformutils.ConvertToTerraformCompliantName(s.PreferredName)
	}
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s *specSchemaDefinitionProperty) isPropertyNamedID() bool {
	return s.getTerraformCompliantPropertyName() == idDefaultPropertyName
}

func (s *specSchemaDefinitionProperty) isPropertyNamedStatus() bool {
	return s.getTerraformCompliantPropertyName() == statusDefaultPropertyName
}

func (s *specSchemaDefinitionProperty) isObjectProperty() bool {
	return s.Type == typeObject
}

func (s *specSchemaDefinitionProperty) isArrayProperty() bool {
	return s.Type == typeList
}

func (s *specSchemaDefinitionProperty) isArrayOfObjectsProperty() bool {
	return s.Type == typeList && s.ArrayItemsType == typeObject
}

func (s *specSchemaDefinitionProperty) isRequired() bool {
	return s.Required
}

func (s *specSchemaDefinitionProperty) isOptional() bool {
	return !s.Required
}

func (s *specSchemaDefinitionProperty) isOptionalComputed() bool {
	return s.OptionalComputed
}

func (s *specSchemaDefinitionProperty) isComputed() bool {
	return s.ReadOnly
}

func (s *specSchemaDefinitionProperty) terraformType() (schema.ValueType, error) {
	switch s.Type {
	case typeObject:
		return schema.TypeMap, nil
	case typeString:
		return schema.TypeString, nil
	case typeInt:
		return schema.TypeInt, nil
	case typeFloat:
		return schema.TypeFloat, nil
	case typeBool:
		return schema.TypeBool, nil
	case typeList:
		return schema.TypeList, nil
	}
	return schema.TypeInvalid, fmt.Errorf("non supported type %s", s.Type)
}

func (s *specSchemaDefinitionProperty) isTerraformListOfSimpleValues() (bool, *schema.Schema) {
	switch s.ArrayItemsType {
	case typeString:
		return true, &schema.Schema{Type: schema.TypeString}
	case typeInt:
		return true, &schema.Schema{Type: schema.TypeInt}
	case typeFloat:
		return true, &schema.Schema{Type: schema.TypeFloat}
	case typeBool:
		return true, &schema.Schema{Type: schema.TypeBool}
	}
	return false, nil
}

func (s *specSchemaDefinitionProperty) terraformObjectSchema() (*schema.Resource, error) {
	if s.Type == typeObject || (s.Type == typeList && s.ArrayItemsType == typeObject) {
		objectSchema, err := s.SpecSchemaDefinition.createResourceSchemaKeepID()
		if err != nil {
			return nil, err
		}
		elem := &schema.Resource{
			Schema: objectSchema,
		}
		return elem, nil
	}
	return nil, fmt.Errorf("object schema can only be formed for types %s or types %s with elems of type %s: found type='%s' elemType='%s' instead", typeObject, typeList, typeObject, s.Type, s.ArrayItemsType)
}

// terraformSchema returns the terraform schema for a the given specSchemaDefinitionProperty
func (s *specSchemaDefinitionProperty) terraformSchema() (*schema.Schema, error) {
	var terraformSchema = &schema.Schema{}

	schemaType, err := s.terraformType()
	if err != nil {
		return nil, err
	}
	terraformSchema.Type = schemaType

	// complex data structures
	switch s.Type {
	case typeObject:
		objectSchema, err := s.terraformObjectSchema()
		if err != nil {
			return nil, err
		}
		terraformSchema.Elem = objectSchema
	case typeList:
		if isListOfPrimitives, elemSchema := s.isTerraformListOfSimpleValues(); isListOfPrimitives {
			terraformSchema.Elem = elemSchema
		} else {
			objectSchema, err := s.terraformObjectSchema()
			if err != nil {
				return nil, err
			}
			terraformSchema.Elem = objectSchema
		}
	}

	// A readOnly property is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	// A optional computed property is exposed to the user and if the value is not provided by the user, the API
	// will compute the value itself. Considering that, setting the property schema as computed, which allows also
	// for the value to be provided as input.
	terraformSchema.Computed = s.isComputed() || s.isOptionalComputed()
	// A sensitive property means that the expectedValue will not be disclosed in the state file, preventing secrets from
	// being leaked
	terraformSchema.Sensitive = s.Sensitive
	// If the expectedValue of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new expectedValue will be created
	terraformSchema.ForceNew = s.ForceNew

	// Set the property as required or optional
	if s.Required {
		terraformSchema.Required = true
	} else {
		terraformSchema.Optional = true
	}

	// ValidateFunc is not yet supported on lists or sets
	if !s.isArrayProperty() && !s.isObjectProperty() {
		terraformSchema.ValidateFunc = s.validateFunc()
	}

	return terraformSchema, nil
}

func (s *specSchemaDefinitionProperty) validateFunc() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if s.ForceNew && s.Immutable {
			errors = append(errors, fmt.Errorf("property '%s' is configured as immutable and can not be configured with forceNew too", s.Name))
		}
		if s.Required && s.ReadOnly {
			errors = append(errors, fmt.Errorf("property '%s' is configured as required and can not be configured as computed too", s.Name))
		}
		return
	}
}
