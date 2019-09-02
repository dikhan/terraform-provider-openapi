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
	Name           string
	PreferredName  string
	Type           schemaDefinitionPropertyType
	ArrayItemsType schemaDefinitionPropertyType
	Required       bool
	// ReadOnly properties are included in responses but not in request
	ReadOnly bool
	// Computed properties describe properties where the value is computed by the API
	Computed bool
	// IsParentProperty defines whether the property is a parent property in which case it will be treated differently in
	// different parts of the code. For instance, the property will not be posted to the API.
	IsParentProperty   bool
	ForceNew           bool
	Sensitive          bool
	Immutable          bool
	IsIdentifier       bool
	IsStatusIdentifier bool
	// EnableLegacyComplexObjectBlockConfiguration defines whether this specSchemaDefinitionProperty should be handled with special treatment following
	// the recommendation from hashi maintainers (https://github.com/hashicorp/terraform/issues/22511#issuecomment-522655851)
	// to support complex object types with the legacy SDK (objects that contain properties with different types and configurations
	// like computed properties).
	EnableLegacyComplexObjectBlockConfiguration bool
	// Default field is only for informative purposes to know what the openapi spec for the property stated the default value is
	// As per the openapi spec default attributes, the value is expected to be computed by the API
	Default interface{}
	// only for object type properties or arrays type properties with array items of type object
	SpecSchemaDefinition *specSchemaDefinition
}

func (s *specSchemaDefinitionProperty) getTerraformCompliantPropertyName() string {
	if s.PreferredName != "" {
		return s.PreferredName
	}
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

// This is the workaround to be able to process objects that contain properties that are not of the same type and may
// contain other configurations like be computed properties (More info here: https://github.com/hashicorp/terraform/issues/22511)
// The object properties that have EnableLegacyComplexObjectBlockConfiguration set to true will be represented in Terraform schema
// as TypeList with MaxItems limited to 1. This will solve the current limitation in Terraform SDK 0.12 where blocks can only be
// translated to lists and sets, maps can not be used to represent complex objects at the moment as it will result into undefined behavior.
// TODO: unit test this method
func (s *specSchemaDefinitionProperty) isLegacyComplexObjectExtensionEnabled() bool {
	if !s.isObjectProperty() {
		return false
	}
	if s.EnableLegacyComplexObjectBlockConfiguration {
		return true
	}
	return false
}

func (s *specSchemaDefinitionProperty) isPropertyWithNestedObjects() (bool, error) {
	if !s.isObjectProperty() {
		return false, nil
	}
	if s.SpecSchemaDefinition == nil {
		return false, fmt.Errorf("missing spec schema definition for object property '%s'", s.Name)
	}
	for _, p := range s.SpecSchemaDefinition.Properties {
		if p.isObjectProperty() {
			return true, nil
		}
	}
	return false, nil
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

func (s *specSchemaDefinitionProperty) isReadOnly() bool {
	return s.ReadOnly
}

func (s *specSchemaDefinitionProperty) isRequired() bool {
	return s.Required
}

func (s *specSchemaDefinitionProperty) isOptional() bool {
	return !s.Required
}

// isOptionalComputed returns true if one of the following cases is met:
//- The property is optional (marked as required=false), in which case there few use cases:
//  - readOnly properties (marked as readOnly=true, computed=true):
//    - with default (default={some value})
//    - with no default (default=nil)
//  - optional-computed (marked as readOnly=false, computed=true):
//    - with no default (default=nil)
func (s *specSchemaDefinitionProperty) isComputed() bool {
	return s.isOptional() && (s.isReadOnly() || s.isOptionalComputed())
}

// isOptionalComputed returns true for properties that are optional and a value (not known at plan time) is computed by the API
// if the client does not provide a value. In order for a property to be considered optional computed it must meet:
// - The property must be optional, readOnly, computed and must not have a default value populated
// Note: optional-computed properties (marked as readOnly=false, computed=true, default={some value}) are not considered
// as optional computed since the way they will be treated as far as the terraform schema will differ. The terraform schema property
// for this properties will contain the default value and the property will not be computed
func (s *specSchemaDefinitionProperty) isOptionalComputed() bool {
	return s.isOptional() && !s.isReadOnly() && s.Computed && s.Default == nil
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
		if s.SpecSchemaDefinition == nil {
			return nil, fmt.Errorf("missing spec schema definition for property '%s' of type '%s'", s.Name, s.Type)
		}
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

// shouldUseLegacyTerraformSDKBlockApproachForComplexObjects returns true if one of the following scenarios match:
// - the specSchemaDefinitionProperty is of type object and in turn contains at lesat one nested property that is an object.
// - the specSchemaDefinitionProperty is of type object and also has the EnableLegacyComplexObjectBlockConfiguration set to true
// In both cases, in order to represent complex objects with the current version of the Terraform SDK (<= v0.12.2), the workaround
// suggested by hashi maintainers is to use TypeList limiting the MaxItems to 1.
// References to the issues opened:
// - https://github.com/hashicorp/terraform/issues/21217#issuecomment-489699737
// - https://github.com/hashicorp/terraform/issues/22511#issuecomment-522655851
// TODO: add unit test for this method
func (s *specSchemaDefinitionProperty) shouldUseLegacyTerraformSDKBlockApproachForComplexObjects() (bool, error) {
	isPropertyWithNestedObjects, err := s.isPropertyWithNestedObjects()
	if err != nil {
		return false, err
	}
	// is of type object and in turn contains at lesat one nested property that is an object.
	if isPropertyWithNestedObjects {
		return true, nil
	}
	// or is of type object and also has the EnableLegacyComplexObjectBlockConfiguration set to true
	return s.isLegacyComplexObjectExtensionEnabled(), nil
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
		// TODO: add coverage for this logic if not already done
		shouldUseLegacyTerraformSDKApproachForBlocks, err := s.shouldUseLegacyTerraformSDKBlockApproachForComplexObjects() // handle the error
		if shouldUseLegacyTerraformSDKApproachForBlocks {
			terraformSchema.Type = schema.TypeList
			terraformSchema.MaxItems = 1
		}
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

	// A computed property could be one of:
	// - property that is set as readOnly in the openapi spec
	// - property that is not readOnly, but it is an optional computed property. The following will comply with optional computed:
	//   - the property is not readOnly and default is nil (only possible when 'x-terraform-computed' extension is set)
	terraformSchema.Computed = s.isComputed()

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

	// Don't populate Default if property is readOnly as the property is expected to be computed by the API. Terraform does
	// not allow properties with Computed = true having the Default field populated, otherwise the following error will be
	// thrown at runtime: Default must be nil if computed
	if !s.isComputed() {
		terraformSchema.Default = s.Default
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
