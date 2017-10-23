package terraform_provider_api

import (
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type ResourceFactory struct {
	Schema spec.Schema
	Create spec.PathItem // Access to Create operation from root path
	Rud    spec.PathItem // Access to Read, Update and Delete operations
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
	for propertyName, property := range r.Schema.Properties {
		s[propertyName] = &schema.Schema{
			Type:     r.getType(property),
			Optional: true,
		}
	}
	PrettyPrint(s)
	return s
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
