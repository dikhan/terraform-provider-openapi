package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
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
	Required           bool
	OptionalComputed   bool
	ReadOnly           bool
	ForceNew           bool
	Sensitive          bool
	Immutable          bool
	IsIdentifier       bool
	IsStatusIdentifier bool
	Default            interface{}
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

func (s *specSchemaDefinitionProperty) isOptionalComputed() bool {
	return s.OptionalComputed
}

func (s *specSchemaDefinitionProperty) isReadOnly() bool {
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

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	terraformSchema.Computed = s.ReadOnly || s.OptionalComputed
	// A sensitive property means that the expectedValue will not be disclosed in the state file, preventing secrets from
	// being leaked
	terraformSchema.Sensitive = s.Sensitive
	// If the expectedValue of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new expectedValue will be created
	terraformSchema.ForceNew = s.ForceNew
	terraformSchema.Default = s.Default

	// Set the property as required or optional
	if s.Required {
		terraformSchema.Required = true
	} else {
		terraformSchema.Optional = true
	}
	if s.Default != nil {
		if s.ReadOnly || s.OptionalComputed {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error:
			// * resource swaggercodegen_cdn_v1: optional_computed: Default must be nil if computed]
			log.Printf("[WARN] '%s' is readOnly and can not have a default expectedValue. The expectedValue is expected to be computed by the API. Terraform will fail on runtime when performing the property validation check", s.Name)
		} else {
			terraformSchema.Default = s.Default
		}
	}

	// ValidateFunc is not yet supported on lists or sets
	if !s.isArrayProperty() && !s.isObjectProperty() {
		terraformSchema.ValidateFunc = s.validateFunc()
	}

	return terraformSchema, nil
}

func (s *specSchemaDefinitionProperty) validateFunc() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if s.Default != nil {
			if s.ReadOnly || s.OptionalComputed {
				err := fmt.Errorf(
					"'%s.%s' is configured as 'readOnly' and can not have a default expectedValue. The expectedValue is expected to be computed by the API. To fix the issue, pick one of the following options:\n"+
						"1. Remove the 'readOnly' attribute from %s in the swagger file so the default expectedValue '%v' can be applied. Default must be nil if computed\n"+
						"OR\n"+
						"2. Remove the 'default' attribute from %s in the swagger file, this means that the API will compute the expectedValue as described with the 'readOnly' attribute\n", s.Name, k, k, s.Default, k)
				errors = append(errors, err)
			}
		}
		if s.ReadOnly && s.OptionalComputed {
			errors = append(errors, fmt.Errorf("property '%s' is configured as readOnly and can not be configured with '%s' too", s.Name, extTfComputed))
		}
		if s.ForceNew && s.Immutable {
			errors = append(errors, fmt.Errorf("property '%s' is configured as immutable and can not be configured with forceNew too", s.Name))
		}
		if s.Required && s.ReadOnly {
			errors = append(errors, fmt.Errorf("property '%s' is configured as required and can not be configured as computed too", s.Name))
		}
		return
	}
}
