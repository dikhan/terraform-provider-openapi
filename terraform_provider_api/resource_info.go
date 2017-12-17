package main

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const EXT_TF_IMMUTABLE = "x-terraform-immutable"
const EXT_TF_FORCE_NEW = "x-terraform-force-new"

type CrudResourcesInfo map[string]ResourceInfo

// ResourceInfo serves as translator between swagger definitions and terraform schemas
type ResourceInfo struct {
	Name string
	// Path contains relative path to the resource e,g: /v1/resource
	Path             string
	Host             string
	HttpSchemes      []string
	SchemaDefinition spec.Schema
	// CreatePathInfo contains info about /resource
	CreatePathInfo spec.PathItem
	// PathInfo contains info about /resource/{id}
	PathInfo spec.PathItem
}

func (r ResourceInfo) createTerraformResourceSchema() (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}
	for propertyName, property := range r.SchemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
		required := r.isRequired(propertyName, r.SchemaDefinition.Required)
		schema, err := r.createTerraformPropertySchema(propertyName, property, required)
		if err != nil {
			return nil, err
		}
		s[propertyName] = schema
	}
	return s, nil
}

func (r ResourceInfo) createTerraformPropertySchema(propertyName string, property spec.Schema, required bool) (*schema.Schema, error) {
	propertySchema, err := r.createTerraformBasicSchema(propertyName, property)
	if err != nil {
		return nil, err
	}
	if required {
		propertySchema.Required = true
	} else {
		propertySchema.Optional = true
	}
	// ValidateFunc is not yet supported on lists or sets
	if !r.isArrayProperty(property) {
		propertySchema.ValidateFunc = r.validateFunc(propertyName, property)
	}
	return propertySchema, nil
}

func (r ResourceInfo) validateFunc(propertyName string, property spec.Schema) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if property.Default != nil {
			if property.ReadOnly {
				err := fmt.Errorf(
					"'%s.%s' is configured as 'readOnly' and can not have a default value. The value is expected to be computed by the API. To fix the issue, pick one of the following options:\n"+
						"1. Remove the 'readOnly' attribute from %s in the swagger file so the default value '%v' can be applied\n"+
						"OR\n"+
						"2. Remove the 'default' attribute from %s in the swagger file, this means that the API will compute the value as specified by the 'readOnly' attribute\n", r.Name, k, k, property.Default, k)
				errors = append(errors, err)
			}
		}
		return
	}
}

func (r ResourceInfo) isRequired(propertyName string, requiredProps []string) bool {
	var required bool = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (r ResourceInfo) createTerraformBasicSchema(propertyName string, property spec.Schema) (*schema.Schema, error) {
	var propertySchema *schema.Schema
	// Arrays only support 'string' items at the moment
	if r.isArrayProperty(property) {
		propertySchema = &schema.Schema{
			Type: schema.TypeList,
			Elem: &schema.Schema{Type: schema.TypeString},
		}
	} else if property.Type.Contains("string") {
		propertySchema = &schema.Schema{
			Type: schema.TypeString,
		}
	} else if property.Type.Contains("integer") {
		propertySchema = &schema.Schema{
			Type: schema.TypeInt,
		}
	} else if property.Type.Contains("number") {
		propertySchema = &schema.Schema{
			Type: schema.TypeFloat,
		}
	} else if property.Type.Contains("boolean") {
		propertySchema = &schema.Schema{
			Type: schema.TypeBool,
		}
	}

	// If the value of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new value will be created
	if forceNew, ok := property.Extensions.GetBool(EXT_TF_FORCE_NEW); ok && forceNew {
		propertySchema.ForceNew = true
	}

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	if property.ReadOnly {
		propertySchema.Computed = true
	}

	if property.Default != nil {
		if property.ReadOnly {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error
			log.Printf("[WARN] '%s.%s' is readOnly and can not have a default value. The value is expected to be computed by the API.", r.Name, propertyName)
		} else {
			propertySchema.Default = property.Default
		}
	}
	return propertySchema, nil
}

func (r ResourceInfo) isArrayProperty(property spec.Schema) bool {
	return property.Type.Contains("array")
}

func (r ResourceInfo) getImmutableProperties() []string {
	var immutableProperties []string
	for propertyName, property := range r.SchemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
		if immutable, ok := property.Extensions.GetBool(EXT_TF_IMMUTABLE); ok && immutable {
			immutableProperties = append(immutableProperties, propertyName)
		}
	}
	return immutableProperties
}

func (r ResourceInfo) getResourceUrl() string {
	defaultScheme := "http"
	for _, scheme := range r.HttpSchemes {
		if scheme == "https" {
			defaultScheme = "https"
		}
	}
	return fmt.Sprintf("%s://%s%s", defaultScheme, r.Host, r.Path)
}

func (r ResourceInfo) getResourceIdUrl(id string) string {
	return fmt.Sprintf("%s/%s", r.getResourceUrl(), id)
}
