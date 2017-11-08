package main

import (
	"fmt"
	"net/http"
	"reflect"

	httpGoClient "github.com/dikhan/http_goclient"
	"github.com/hashicorp/terraform/helper/schema"
)

type ResourceFactory struct {
	httpClient   *http.Client
	ResourceInfo ResourceInfo
}

func (r ResourceFactory) createSchemaResource() *schema.Resource {
	return &schema.Resource{
		Schema: r.ResourceInfo.createTerraformResourceSchema(),
		Create: r.create,
		Read:   r.read,
		Delete: r.delete,
		Update: r.update,
	}
}

func (r ResourceFactory) create(data *schema.ResourceData, i interface{}) error {
	input := r.getPayloadFromData(data)
	output := map[string]interface{}{}
	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.getResourceUrl()
	_, err := httpClient.PostJson(url, nil, input, &output)
	if err != nil {
		return err
	}
	data.SetId(output["id"].(string))
	return nil
}

func (r ResourceFactory) read(data *schema.ResourceData, i interface{}) error {
	output := map[string]interface{}{}
	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.getResourceIdUrl(data.Id())
	_, err := httpClient.Get(url, nil, &output)
	if err != nil {
		return err
	}
	r.updateResourceState(output, data)
	return nil
}

func (r ResourceFactory) update(data *schema.ResourceData, i interface{}) error {
	input := r.getPayloadFromData(data)
	output := map[string]interface{}{}

	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.getResourceIdUrl(data.Id())
	_, err := httpClient.PutJson(url, nil, input, &output)
	if err != nil {
		return err
	}
	r.updateResourceState(output, data)
	return nil
}

func (r ResourceFactory) delete(data *schema.ResourceData, i interface{}) error {
	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.getResourceIdUrl(data.Id())
	_, err := httpClient.Delete(url, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r ResourceFactory) updateResourceState(input map[string]interface{}, data *schema.ResourceData) {
	for propertyName, propertyValue := range input {
		data.Set(propertyName, propertyValue)
	}
}

func (r ResourceFactory) getPayloadFromData(data *schema.ResourceData) map[string]interface{} {
	input := map[string]interface{}{}
	for propertyName, _ := range r.ResourceInfo.SchemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
		if reflect.TypeOf(data.Get(propertyName)).Kind() == reflect.Slice {
			input[propertyName] = data.Get(propertyName).([]interface{})
		} else {
			input[propertyName] = data.Get(propertyName).(string)
		}
	}
	return input
}

func (r ResourceFactory) getResourceUrl() string {
	return fmt.Sprintf("http://%s%s", r.ResourceInfo.Host, r.ResourceInfo.Path)
}

func (r ResourceFactory) getResourceIdUrl(id string) string {
	return fmt.Sprintf("%s/%s", r.getResourceUrl(), id)
}
