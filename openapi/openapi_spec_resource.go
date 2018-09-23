package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

// SpecResource defines the behaviour related to terraform compliant OpenAPI Resources.
type SpecResource interface {
	getResourceName() string
	getResourcePath() string
	getResourceSchema() (*SchemaDefinition, error)

	shouldIgnoreResource() bool

	getResourcePostOperation() *ResourceOperation
	getResourceGetOperation() *ResourceOperation
	getResourcePutOperation() *ResourceOperation
	getResourceDeleteOperation() *ResourceOperation
}

// ResourceOperation defines a resource operation
type ResourceOperation struct {
	SecuritySchemes  SpecSecuritySchemes
	HeaderParameters SpecHeaderParameters
}

// SchemaDefinitionPropertyType defines the type of a property
type SchemaDefinitionPropertyType string

const (
	typeString SchemaDefinitionPropertyType = "string"
	typeInt    SchemaDefinitionPropertyType = "integer"
	typeFloat  SchemaDefinitionPropertyType = "number"
	typeBool   SchemaDefinitionPropertyType = "boolean"
	typeList   SchemaDefinitionPropertyType = "list"
)

// SchemaDefinitionProperty defines the attributes for a schema property
type SchemaDefinitionProperty struct {
	Name          string
	PreferredName string
	Type          SchemaDefinitionPropertyType
	Required      bool
	ReadOnly      bool
	ForceNew      bool
	Sensitive     bool
	Immutable     bool
	IsIdentifier  bool
	Default       interface{}
}

func newStringSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newStringSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newStringSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeString, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newIntSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newIntSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newIntSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeInt, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newNumberSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newNumberSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newNumberSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeFloat, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newBoolSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newBoolSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newBoolSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeBool, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newListSchemaDefinitionPropertyWithDefaults(name, preferredName string, required, readOnly bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newListSchemaDefinitionProperty(name, preferredName, required, readOnly, false, false, false, false, defaultValue)
}

func newListSchemaDefinitionProperty(name, preferredName string, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return newSchemaDefinitionProperty(name, preferredName, typeList, required, readOnly, forceNew, sensitive, immutable, isIdentifier, defaultValue)
}

func newSchemaDefinitionProperty(name, preferredName string, propertyType SchemaDefinitionPropertyType, required, readOnly, forceNew, sensitive, immutable, isIdentifier bool, defaultValue interface{}) *SchemaDefinitionProperty {
	return &SchemaDefinitionProperty{
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

func (s *SchemaDefinitionProperty) getTerraformCompliantPropertyName() string {
	if s.PreferredName != "" {
		return terraformutils.ConvertToTerraformCompliantName(s.PreferredName)
	}
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s *SchemaDefinitionProperty) isPropertyNamedID() bool {
	return s.getTerraformCompliantPropertyName() == "id"
}

func (s *SchemaDefinitionProperty) isArrayProperty() bool {
	return s.Type == typeList
}

func (s *SchemaDefinitionProperty) isRequired() bool {
	return s.Required
}

func (s *SchemaDefinitionProperty) isReadOnly() bool {
	return s.ReadOnly
}

// terraformSchema returns the terraform schema for a the given SchemaDefinitionProperty
func (s *SchemaDefinitionProperty) terraformSchema() *schema.Schema {
	terraformSchema := &schema.Schema{
		// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
		// it comes back from the api and is stored in the state. This properties are mostly informative.
		Computed: s.ReadOnly,
		// A sensitive property means that the expectedValue will not be disclosed in the state file, preventing secrets from
		// being leaked
		Sensitive: s.Sensitive,
		// If the expectedValue of the property is changed, it will force the deletion of the previous generated resource and
		// a new resource with this new expectedValue will be created
		ForceNew: s.ForceNew,
		Default:  s.Default,
		// Set the property as required or optional
		Required: s.Required,
	}
	switch s.Type {
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
	return terraformSchema
}

func (s *SchemaDefinitionProperty) validateFunc() schema.SchemaValidateFunc {
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

// SchemaDefinitionProperties defines a collection of schema definition properties
type SchemaDefinitionProperties map[string]*SchemaDefinitionProperty

// SchemaDefinition defines a struct for a schema definition
type SchemaDefinition struct {
	Properties SchemaDefinitionProperties
}

func (s *SchemaDefinition) getImmutableProperties() []string {
	var immutableProperties []string
	for _, property := range s.Properties {
		if property.isPropertyNamedID() {
			continue
		}
		if property.Immutable {
			immutableProperties = append(immutableProperties, property.Name)
		}
	}
	return immutableProperties
}

//// getResourceIdentifier returns the property name that is supposed to be used as the identifier. The resource id
//// is selected as follows:
//// 1.If the given schema definition contains a property configured with metadata 'x-terraform-id' set to true, that property value
//// will be used to set the state ID of the resource. Additionally, the value will be used when performing GET/PUT/DELETE requests to
//// identify the resource in question.
//// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
//// will have a property named 'id'
//// 3. If none of the above requirements is met, an error will be returned
func (s *SchemaDefinition) getResourceIdentifier() (string, error) {
	identifierProperty := ""
	for _, property := range s.Properties {
		if property.isPropertyNamedID() {
			identifierProperty = property.Name
			continue
		}
		if property.IsIdentifier {
			identifierProperty = property.Name
			break
		}
	}
	// if the identifier property is missing, there is not way for the resource to be identified and therefore an error is returned
	if identifierProperty == "" {
		return "", fmt.Errorf("could not find any identifier property in the resource schema definition")
	}
	return identifierProperty, nil
}

func (s *SchemaDefinition) getProperty(name string) (*SchemaDefinitionProperty, error) {
	if property, exists := s.Properties[name]; exists {
		return property, nil
	}
	return nil, fmt.Errorf("property with name '%s' not existing in resource schema definition", name)
}
