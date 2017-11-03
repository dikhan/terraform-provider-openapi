package terraform_provider_api

import (
	"fmt"
	httpGoClient "github.com/dikhan/http_goclient"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
)

type ResourceFactory struct {
	ResourceInfo ResourceInfo
}

func (r ResourceFactory) createSchemaResource() *schema.Resource {
	return &schema.Resource{
		Schema: r.createSchema(),
		Create: r.create,
		Read:   read,
		Delete: delete,
		Update: update,
	}
}

func (r ResourceFactory) createSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{}
	for propertyName, property := range r.ResourceInfo.SchemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
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

func (r ResourceFactory) create(data *schema.ResourceData, i interface{}) error {
	input := map[string]string{}
	output := map[string]string{}
	for propertyName, _ := range r.ResourceInfo.SchemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
		input[propertyName] = data.Get(propertyName).(string)
	}
	httpClient := httpGoClient.HttpClient{&http.Client{}}
	url := fmt.Sprintf("http://%s/%s", r.ResourceInfo.Host, r.ResourceInfo.Name)
	_, err := httpClient.PostJson(url, nil, input, &output)
	if err != nil {
		return err
	}
	data.SetId(output["id"])
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
