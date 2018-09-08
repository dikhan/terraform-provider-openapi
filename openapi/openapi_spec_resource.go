package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

// SpecResource defines the behaviour related to terraform compliant OpenAPI Resources.
type SpecResource interface {
	getResourceName() string
	getResourcePath() string
	getResourceSchema() (SchemaDefinition, error)

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

// SchemaDefinition defines a struct for a schema definition
type SchemaDefinition struct {
	Properties map[string]SchemaDefinitionProperty
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

func (s *SchemaDefinition) getProperty(name string) (SchemaDefinitionProperty, error) {
	if property, exists := s.Properties[name]; exists {
		return property, nil
	}
	return SchemaDefinitionProperty{}, fmt.Errorf("property with name '%s' not existing in resource schema definition", name)
}
