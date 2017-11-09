package main

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type CrudResourcesInfo map[string]ResourceInfo

type ResourceInfo struct {
	Name             string
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
		s[propertyName] = r.createPropertySchema(propertyName, property, required)
	}
	return s
}

func (r ResourceInfo) createPropertySchema(propertyName string, property spec.Schema, required bool) *schema.Schema {
	propertySchema := r.createBasicSchema(property)
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

func (r ResourceInfo) createBasicSchema(property spec.Schema) *schema.Schema {
	var propertySchema *schema.Schema
	if property.Type.Contains("array") {
		propertySchema = &schema.Schema{
			Type: schema.TypeList,
			Elem: &schema.Schema{Type: schema.TypeString},
		}
	} else {
		propertySchema = &schema.Schema{
			Type: schema.TypeString,
		}
	}
	return propertySchema
}

func (r ResourceInfo) getType(property spec.Schema) schema.ValueType {
	if property.Type.Contains("array") {
		return schema.TypeList
	}
	return schema.TypeString
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
