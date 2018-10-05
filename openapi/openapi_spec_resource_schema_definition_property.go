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
	Required           bool
	ReadOnly           bool
	ForceNew           bool
	Sensitive          bool
	Immutable          bool
	IsIdentifier       bool
	IsStatusIdentifier bool
	Default            interface{}
	// only for object type properties
	specSchemaDefinition specSchemaDefinition
}

func newStringSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newStringSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newStringSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeString, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newIntSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newIntSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newIntSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeInt, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newNumberSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newNumberSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newNumberSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeFloat, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newBoolSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newBoolSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newBoolSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeBool, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newListSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newListSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newListSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeList, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newSchemaDefinitionProperty(name, preferredName string, propertyType schemaDefinitionPropertyType, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *specSchemaDefinitionProperty {
	return &specSchemaDefinitionProperty{
		Name:          name,
		PreferredName: preferredName,
		Type:          propertyType,
		Required:      required,
		ReadOnly:      readOnly,
		ForceNew:      forceNew,
		Sensitive:     sensitive,
		Immutable:     immutable,
		IsIdentifier:  isIdentifier,
		Default:       defaultValue,
	}
}

func (s *specSchemaDefinitionProperty) getTerraformCompliantPropertyName() string {
	if s.PreferredName != "" {
		return terraformutils.ConvertToTerraformCompliantName(s.PreferredName)
	}
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s *specSchemaDefinitionProperty) isPropertyNamedID() bool {
	return s.getTerraformCompliantPropertyName() == "id"
}

func (s *specSchemaDefinitionProperty) isObjectProperty() bool {
	return s.Type == typeObject
}

func (s *specSchemaDefinitionProperty) isArrayProperty() bool {
	return s.Type == typeList
}

func (s *specSchemaDefinitionProperty) isRequired() bool {
	return s.Required
}

func (s *specSchemaDefinitionProperty) isReadOnly() bool {
	return s.ReadOnly
}

// terraformSchema returns the terraform schema for a the given specSchemaDefinitionProperty
func (s *specSchemaDefinitionProperty) terraformSchema() (*schema.Schema, error) {
	var terraformSchema = &schema.Schema{}
	switch s.Type {
	case typeObject:
		objectSchema, err := s.specSchemaDefinition.createResourceSchemaKeepID()
		if err != nil {
			return nil, err
		}
		terraformSchema = &schema.Schema{
			Type: schema.TypeMap,
			Elem: &schema.Resource{
				Schema: objectSchema,
			},
		}
	case typeString:
		terraformSchema.Type = schema.TypeString
	case typeInt:
		terraformSchema.Type = schema.TypeInt
	case typeFloat:
		terraformSchema.Type = schema.TypeFloat
	case typeBool:
		terraformSchema.Type = schema.TypeBool
	case typeList:
		terraformSchema.Type = schema.TypeList
		terraformSchema.Elem = &schema.Schema{Type: schema.TypeString}
	}

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	terraformSchema.Computed = s.ReadOnly
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
		if s.ReadOnly {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error
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
			if s.ReadOnly {
				err := fmt.Errorf(
					"'%s.%s' is configured as 'readOnly' and can not have a default expectedValue. The expectedValue is expected to be computed by the API. To fix the issue, pick one of the following options:\n"+
						"1. Remove the 'readOnly' attribute from %s in the swagger file so the default expectedValue '%v' can be applied. Default must be nil if computed\n"+
						"OR\n"+
						"2. Remove the 'default' attribute from %s in the swagger file, this means that the API will compute the expectedValue as specified by the 'readOnly' attribute\n", s.Name, k, k, s.Default, k)
				errors = append(errors, err)
			}
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
