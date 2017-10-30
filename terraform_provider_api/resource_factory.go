package terraform_provider_api

import (
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type ResourceFactory struct {
	ResourceInfo ResourceInfo
}

func (r ResourceFactory) createSchemaResource() *schema.Resource {
	return &schema.Resource{
		Schema: r.createSchema(),
		Create: create,
		Read:   read,
		Delete: delete,
		Update: update,
	}
}

func (r ResourceFactory) createSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{}
	for propertyName, property := range r.ResourceInfo.SchemaDefinition.Properties {
		s[propertyName] = r.createPropertySchema(propertyName, property, r.ResourceInfo.SchemaDefinition.Required)
	}
	return s
}

func (r ResourceFactory) createPropertySchema(propertyName string, property spec.Schema, requiredProps []string) *schema.Schema {
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
	r.setPropertyToRequiredOrOptional(propertySchema, propertyName, requiredProps)
	return propertySchema
}

func (r ResourceFactory) setPropertyToRequiredOrOptional(propertySchema *schema.Schema, propertyName string, requiredProps []string) {
	var required bool = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	if required {
		propertySchema.Required = true
	} else {
		propertySchema.Optional = true
	}
}

func (r ResourceFactory) getType(property spec.Schema) schema.ValueType {
	if property.Type.Contains("array") {
		return schema.TypeList
	}
	return schema.TypeString
}

func create(data *schema.ResourceData, i interface{}) error {
	return nil
}

func read(data *schema.ResourceData, i interface{}) error {
	return nil
}

func update(data *schema.ResourceData, i interface{}) error {
	return nil
}

func delete(data *schema.ResourceData, i interface{}) error {
	return nil
}
