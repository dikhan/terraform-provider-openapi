package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

// specSchemaDefinitionProperties defines a collection of schema definition properties
type specSchemaDefinitionProperties map[string]*specSchemaDefinitionProperty

// specSchemaDefinition defines a struct for a schema definition
type specSchemaDefinition struct {
	Properties specSchemaDefinitionProperties
}

func (s *specSchemaDefinition) getImmutableProperties() []string {
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
func (s *specSchemaDefinition) getResourceIdentifier() (string, error) {
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

// getStatusIdentifier returns the property name that is supposed to be used as the status field. The status field
// is selected as follows:
// 1.If the given schema definition contains a property configured with metadata 'x-terraform-field-status' set to true, that property
// will be used to check the different statues for the asynchronous pooling mechanism.
// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// will have a property named 'status'
// 3. If none of the above requirements is met, an error will be returned
func (s *specSchemaDefinition) getStatusIdentifier() (string, error) {
	statusProperty := ""
	for propertyName, property := range s.Properties {
		if s.isIDProperty(propertyName) {
			continue
		}
		if s.isStatusProperty(propertyName) {
			statusProperty = propertyName
			continue
		}
		// field with extTfFieldStatus metadata takes preference over 'status' fields as the service provider is the one acknowledging
		// the fact that this field should be used as identifier of the resource
		if property.IsStatusIdentifier {
			statusProperty = propertyName
			break
		}
	}
	// if the id field is missing and there isn't any properties set with extTfFieldStatus, there is not way for the resource
	// to be identified and therefore an error is returned
	if statusProperty == "" {
		return "", fmt.Errorf("could not find any status property. Please make sure the resource schema definition has either one property named '%s' or one property is marked with IsStatusIdentifier set to true", statusDefaultPropertyName)
	}
	if !s.Properties[statusProperty].ReadOnly {
		return "", fmt.Errorf("schema definition status property '%s' must be readOnly", statusProperty)
	}
	return statusProperty, nil
}

func (s *specSchemaDefinition) isIDProperty(propertyName string) bool {
	return s.propertyNameMatchesDefaultName(propertyName, idDefaultPropertyName)
}

func (s *specSchemaDefinition) isStatusProperty(propertyName string) bool {
	return s.propertyNameMatchesDefaultName(propertyName, statusDefaultPropertyName)
}

func (s *specSchemaDefinition) propertyNameMatchesDefaultName(propertyName, expectedPropertyName string) bool {
	return terraformutils.ConvertToTerraformCompliantName(propertyName) == expectedPropertyName
}

func (s *specSchemaDefinition) getProperty(name string) (*specSchemaDefinitionProperty, error) {
	if property, exists := s.Properties[name]; exists {
		return property, nil
	}
	return nil, fmt.Errorf("property with name '%s' not existing in resource schema definition", name)
}
