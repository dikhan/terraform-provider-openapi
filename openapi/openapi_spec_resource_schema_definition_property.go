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
	// TODO: remove this once the isPropertyWithNestedObjects() method has been implemented
	IsNestedObject bool
	Required       bool
	// ReadOnly properties are included in responses but not in request
	ReadOnly bool
	// Computed properties describe properties where the value is computed by the API
	Computed           bool
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

		// As per @apparentlymart comment in https://github.com/hashicorp/terraform/issues/21217#issuecomment-489699737, currently (terraform sdk <= v0.12.2) the only
		// way to configure nested structs is by defining a TypeList property which contains the object schema in the elem
		// AND the list is restricted to 1 element. The below is a workaround to support this, however this should go away
		// once the SDK supports this out-of-the-box. Note: When this behaviour changes, it will require a new major release
		// of the provider since the terraform configuration will most likely be different (as well as the way the data is stored
		// in the state file) AND the change will NOT be backwards compatible.
		if s.IsNestedObject {
			terraformSchema.Type = schema.TypeList
			terraformSchema.MaxItems = 1
			//s.ArrayItemsType = typeObject // todo: added fradiben
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
