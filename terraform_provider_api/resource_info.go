package main

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
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

func (r ResourceInfo) createTerraformResourceSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{}
	for propertyName, property := range r.SchemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
		required := r.isRequired(propertyName, r.SchemaDefinition.Required)
		s[propertyName] = r.createTerraformPropertySchema(propertyName, property, required)
	}
	return s
}

func (r ResourceInfo) createTerraformPropertySchema(propertyName string, property spec.Schema, required bool) *schema.Schema {
	propertySchema := r.createTerraformBasicSchema(property)
	if required {
		propertySchema.Required = true
	} else {
		propertySchema.Optional = true
	}
	return propertySchema
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

func (r ResourceInfo) createTerraformBasicSchema(property spec.Schema) *schema.Schema {
	var propertySchema *schema.Schema
	// Arrays only support 'string' items at the moment
	if property.Type.Contains("array") {
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

	return propertySchema
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
