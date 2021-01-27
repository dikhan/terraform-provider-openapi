package openapi

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"reflect"

	"github.com/dikhan/terraform-provider-openapi/v2/openapi/terraformutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// schemaDefinitionPropertyType defines the type of a property
type schemaDefinitionPropertyType string

const (
	// TypeString defines a schema definition property of type int
	TypeString schemaDefinitionPropertyType = "string"
	// TypeInt defines a schema definition property of type int
	TypeInt schemaDefinitionPropertyType = "integer"
	// TypeFloat defines a schema definition property of type float
	TypeFloat schemaDefinitionPropertyType = "number"
	// TypeBool defines a schema definition property of type bool
	TypeBool schemaDefinitionPropertyType = "boolean"
	// TypeList defines a schema definition property of type list
	TypeList schemaDefinitionPropertyType = "list"
	// TypeObject defines a schema definition property of type object
	TypeObject schemaDefinitionPropertyType = "object"
)

const idDefaultPropertyName = "id"
const statusDefaultPropertyName = "status"

// SpecSchemaDefinitionProperty defines the attributes for a schema property
type SpecSchemaDefinitionProperty struct {
	Name           string
	PreferredName  string
	Type           schemaDefinitionPropertyType
	ArrayItemsType schemaDefinitionPropertyType
	Description    string

	// IgnoreItemsOrder if set to true means that the array items order should be ignored
	IgnoreItemsOrder bool

	Required bool
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
	// Default field is only for informative purposes to know what the openapi spec for the property stated the default value is
	// As per the openapi spec default attributes, the value is expected to be computed by the API
	Default interface{}
	// only for object type properties or arrays type properties with array items of type object
	SpecSchemaDefinition *SpecSchemaDefinition
}

func (s *SpecSchemaDefinitionProperty) isPrimitiveProperty() bool {
	if s.Type == TypeString || s.Type == TypeInt || s.Type == TypeFloat || s.Type == TypeBool {
		return true
	}
	return false
}

// GetTerraformCompliantPropertyName returns the property name converted to a terraform compliant name if needed following the snake_case naming convention
func (s *SpecSchemaDefinitionProperty) GetTerraformCompliantPropertyName() string {
	if s.PreferredName != "" {
		return s.PreferredName
	}
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s *SpecSchemaDefinitionProperty) isPropertyWithNestedObjects() bool {
	if !s.isObjectProperty() || s.SpecSchemaDefinition == nil {
		return false
	}
	for _, p := range s.SpecSchemaDefinition.Properties {
		if p.isObjectProperty() {
			return true
		}
	}
	return false
}

func (s *SpecSchemaDefinitionProperty) isPropertyNamedID() bool {
	return s.GetTerraformCompliantPropertyName() == idDefaultPropertyName
}

func (s *SpecSchemaDefinitionProperty) isPropertyNamedStatus() bool {
	return s.GetTerraformCompliantPropertyName() == statusDefaultPropertyName
}

func (s *SpecSchemaDefinitionProperty) isObjectProperty() bool {
	return s.Type == TypeObject
}

func (s *SpecSchemaDefinitionProperty) isArrayProperty() bool {
	return s.Type == TypeList
}

func (s *SpecSchemaDefinitionProperty) shouldIgnoreOrder() bool {
	return s.Type == TypeList && s.IgnoreItemsOrder
}

func (s *SpecSchemaDefinitionProperty) isArrayOfObjectsProperty() bool {
	return s.Type == TypeList && s.ArrayItemsType == TypeObject
}

func (s *SpecSchemaDefinitionProperty) isReadOnly() bool {
	return s.ReadOnly
}

// IsRequired exposes whether a property is required
func (s *SpecSchemaDefinitionProperty) IsRequired() bool {
	return s.Required
}

func (s *SpecSchemaDefinitionProperty) isOptional() bool {
	return !s.Required
}

func (s *SpecSchemaDefinitionProperty) shouldIgnoreArrayItemsOrder() bool {
	return s.isArrayProperty() && s.IgnoreItemsOrder
}

// isComputed returns true if one of the following cases is met:
//- The property is optional (marked as required=false), in which case there few use cases:
//  - readOnly properties (marked as readOnly=true):
//  - optional-computed (marked as readOnly=false, computed=true):
//    - with no default (default=nil)
func (s *SpecSchemaDefinitionProperty) isComputed() bool {
	return s.isOptional() && (s.isReadOnly() || s.IsOptionalComputed())
}

// IsOptionalComputed returns true for properties that are optional and a value (not known at plan time) is computed by the API
// if the client does not provide a value. In order for a property to be considered optional computed it must meet:
// - The property must be optional, not readOnly, computed and must not have a default value populated
// Note: optional-computed properties (marked as readOnly=false, computed=true, default={some value}) are not considered
// as optional computed since the way they will be treated as far as the terraform schema will differ. The terraform schema property
// for this properties will contain the default value and the property will not be computed
func (s *SpecSchemaDefinitionProperty) IsOptionalComputed() bool {
	return s.isOptional() && !s.isReadOnly() && s.Computed && s.Default == nil
}

func (s *SpecSchemaDefinitionProperty) terraformType() (schema.ValueType, error) {
	switch s.Type {
	case TypeString:
		return schema.TypeString, nil
	case TypeInt:
		return schema.TypeInt, nil
	case TypeFloat:
		return schema.TypeFloat, nil
	case TypeBool:
		return schema.TypeBool, nil
	case TypeObject, TypeList:
		return schema.TypeList, nil
	}
	return schema.TypeInvalid, fmt.Errorf("non supported type %s", s.Type)
}

func (s *SpecSchemaDefinitionProperty) isTerraformListOfSimpleValues() (bool, *schema.Schema) {
	switch s.ArrayItemsType {
	case TypeString:
		return true, &schema.Schema{Type: schema.TypeString}
	case TypeInt:
		return true, &schema.Schema{Type: schema.TypeInt}
	case TypeFloat:
		return true, &schema.Schema{Type: schema.TypeFloat}
	case TypeBool:
		return true, &schema.Schema{Type: schema.TypeBool}
	}
	return false, nil
}

func (s *SpecSchemaDefinitionProperty) terraformObjectSchema() (*schema.Resource, error) {
	if s.Type == TypeObject || (s.Type == TypeList && s.ArrayItemsType == TypeObject) {
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
	return nil, fmt.Errorf("object schema can only be formed for types %s or types %s with elems of type %s: found type='%s' elemType='%s' instead", TypeObject, TypeList, TypeObject, s.Type, s.ArrayItemsType)
}

// shouldUseLegacyTerraformSDKBlockApproachForComplexObjects returns true if one of the following scenarios match:
// - the SpecSchemaDefinitionProperty is of type object and in turn contains at least one nested property that is an object.
// - the SpecSchemaDefinitionProperty is of type object and also has the EnableLegacyComplexObjectBlockConfiguration set to true
// In both cases, in order to represent complex objects with the current version of the Terraform SDK (<= v0.12.2), the workaround
// suggested by hashi maintainers is to use TypeList limiting the MaxItems to 1.
// References to the issues opened:
// - https://github.com/hashicorp/terraform-plugin-sdk/issues/616
// - https://github.com/hashicorp/terraform/issues/21217#issuecomment-489699737
// - https://github.com/hashicorp/terraform/issues/22511#issuecomment-522655851
func (s *SpecSchemaDefinitionProperty) shouldUseLegacyTerraformSDKBlockApproachForComplexObjects() bool {
	// Terraform SDK 2.0 upgrade: https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#more-robust-validation-of-helper-schema-typemap-elems
	// Treating all object types as helper/schema.TypeList with Elem *helper/schema.Resource and MaxItems 1
	if s.isObjectProperty() {
		return true
	}
	return false
}

// terraformSchema returns the terraform schema for a the given SpecSchemaDefinitionProperty
func (s *SpecSchemaDefinitionProperty) terraformSchema() (*schema.Schema, error) {
	var terraformSchema = &schema.Schema{}

	schemaType, err := s.terraformType()
	if err != nil {
		return nil, err
	}
	terraformSchema.Type = schemaType
	terraformSchema.Description = s.Description

	// complex data structures
	switch s.Type {
	case TypeObject:
		// Deprecated, treating all objects equally regardless whether they are simple or complex objects
		//if s.shouldUseLegacyTerraformSDKBlockApproachForComplexObjects() {
		//	terraformSchema.Type = schema.TypeList
		//	terraformSchema.MaxItems = 1
		//}
		terraformSchema.Type = schema.TypeList
		terraformSchema.MaxItems = 1
		objectSchema, err := s.terraformObjectSchema()
		if err != nil {
			return nil, err
		}
		terraformSchema.Elem = objectSchema

	case TypeList:
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
		terraformSchema.ValidateDiagFunc = s.validateDiagFunc()
	}

	// Don't populate Default if property is readOnly as the property is expected to be computed by the API. Terraform does
	// not allow properties with Computed = true having the Default field populated, otherwise the following error will be
	// thrown at runtime: Default must be nil if computed
	if !s.isComputed() {
		terraformSchema.Default = s.Default
	}

	return terraformSchema, nil
}

func (s *SpecSchemaDefinitionProperty) validateDiagFunc() schema.SchemaValidateDiagFunc {
	return func(v interface{}, p cty.Path) diag.Diagnostics {
		_, errs := s.validateFunc()(v, "") // it's not clear what would be the value of k with the new schema.SchemaValidateDiagFunc and whether it can be extracted from the cty.Path
		var diags diag.Diagnostics
		if errs != nil && len(errs) > 0 {
			for _, e := range errs {
				diags = append(diags, diag.FromErr(e)...)
			}
		}
		return diags
	}
}

func (s *SpecSchemaDefinitionProperty) validateFunc() schema.SchemaValidateFunc {
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

func (s *SpecSchemaDefinitionProperty) equal(item1, item2 interface{}) bool {
	return s.equalItems(s.Type, item1, item2)
}

func (s *SpecSchemaDefinitionProperty) equalItems(itemsType schemaDefinitionPropertyType, item1, item2 interface{}) bool {
	switch itemsType {
	case TypeString:
		if !s.validateValueType(item1, reflect.String) || !s.validateValueType(item2, reflect.String) {
			return false
		}
	case TypeInt:
		if !s.validateValueType(item1, reflect.Int) || !s.validateValueType(item2, reflect.Int) {
			return false
		}
	case TypeFloat:
		if !s.validateValueType(item1, reflect.Float64) || !s.validateValueType(item2, reflect.Float64) {
			return false
		}
	case TypeBool:
		if !s.validateValueType(item1, reflect.Bool) || !s.validateValueType(item2, reflect.Bool) {
			return false
		}
	case TypeList:
		if !s.validateValueType(item1, reflect.Slice) || !s.validateValueType(item2, reflect.Slice) {
			return false
		}
		list1 := item1.([]interface{})
		list2 := item2.([]interface{})
		if len(list1) != len(list2) {
			return false
		}
		if s.shouldIgnoreOrder() {
			for idx := range list1 {
				match := false
				for idx2 := range list2 {
					if s.equalItems(s.ArrayItemsType, list1[idx], list2[idx2]) {
						match = true
						break
					}
				}
				if !match {
					return false
				}
			}
			return true
		}
		for idx := range list1 {
			return s.equalItems(s.ArrayItemsType, list1[idx], list2[idx])
		}
	case TypeObject:
		if !s.validateValueType(item1, reflect.Map) || !s.validateValueType(item2, reflect.Map) {
			return false
		}
		object1 := item1.(map[string]interface{})
		object2 := item2.(map[string]interface{})
		for _, objectProperty := range s.SpecSchemaDefinition.Properties {
			objectPropertyValue1 := object1[objectProperty.Name]
			objectPropertyValue2 := object2[objectProperty.Name]
			if !objectProperty.equal(objectPropertyValue1, objectPropertyValue2) {
				return false
			}
		}
		return true
	default:
		return false
	}
	return item1 == item2
}

func (s *SpecSchemaDefinitionProperty) validateValueType(item interface{}, expectedKind reflect.Kind) bool {
	if reflect.TypeOf(item).Kind() != expectedKind {
		return false
	}
	return true
}
