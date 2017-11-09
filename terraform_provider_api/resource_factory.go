package main

import (
	"fmt"
	"net/http"
	"reflect"

	"io/ioutil"

	httpGoClient "github.com/dikhan/http_goclient"
	"github.com/go-openapi/spec"
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

func (r ResourceFactory) checkHttpStatusCode(res *http.Response, expectedHttpStatusCode int) error {
	if res.StatusCode != expectedHttpStatusCode {
		b, _ := ioutil.ReadAll(res.Body)
		if len(b) > 0 {
			return fmt.Errorf("response status code %d not matching expected one %d. Error = %s", res.StatusCode, expectedHttpStatusCode, string(b))
		}
		return fmt.Errorf("response status code %d not matching expected one %d", res.StatusCode, expectedHttpStatusCode)
	}
	return nil
}

func (r ResourceFactory) create(data *schema.ResourceData, i interface{}) error {
	input := r.getPayloadFromData(data)
	output := map[string]interface{}{}
	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.ResourceInfo.getResourceUrl()

	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.CreatePathInfo.Post, i.(ProviderConfig), url)
	res, err := httpClient.PostJson(url, headers, input, &output)
	if err != nil {
		return err
	}
	if err := r.checkHttpStatusCode(res, http.StatusCreated); err != nil {
		return fmt.Errorf("POST %s returned an error. Error = %s", url, err)
	}
	data.SetId(output["id"].(string))
	return nil
}

func (r ResourceFactory) read(data *schema.ResourceData, i interface{}) error {
	output := map[string]interface{}{}
	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.ResourceInfo.getResourceIdUrl(data.Id())

	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.PathInfo.Get, i.(ProviderConfig), url)
	res, err := httpClient.Get(url, headers, &output)
	if err != nil {
		return err
	}
	if err := r.checkHttpStatusCode(res, http.StatusOK); err != nil {
		return fmt.Errorf("GET %s returned an error. Error = %s", url, err)
	}
	r.updateResourceState(output, data)
	return nil
}

func (r ResourceFactory) update(data *schema.ResourceData, i interface{}) error {
	input := r.getPayloadFromData(data)
	output := map[string]interface{}{}

	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.ResourceInfo.getResourceIdUrl(data.Id())

	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.PathInfo.Put, i.(ProviderConfig), url)
	res, err := httpClient.PutJson(url, headers, input, &output)
	if err != nil {
		return err
	}
	if err := r.checkHttpStatusCode(res, http.StatusOK); err != nil {
		return fmt.Errorf("UPDATE %s returned an error. Error = %s", url, err)
	}
	r.updateResourceState(output, data)
	return nil
}

func (r ResourceFactory) delete(data *schema.ResourceData, i interface{}) error {
	httpClient := httpGoClient.HttpClient{HttpClient: r.httpClient}
	url := r.ResourceInfo.getResourceIdUrl(data.Id())

	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.PathInfo.Delete, i.(ProviderConfig), url)
	res, err := httpClient.Delete(url, headers)
	if err != nil {
		return err
	}
	if err := r.checkHttpStatusCode(res, http.StatusNoContent); err != nil {
		return fmt.Errorf("DELETE %s returned an error. Error = %s", url, err)
	}
	return nil
}

func (r ResourceFactory) updateResourceState(input map[string]interface{}, data *schema.ResourceData) {
	for propertyName, propertyValue := range input {
		if propertyName == "id" {
			continue
		}
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

func (r ResourceFactory) authRequired(operation *spec.Operation, providerConfig ProviderConfig) (bool, string) {
	for _, operationSecurityPolicy := range operation.Security {
		for operationSecurityDefName, _ := range operationSecurityPolicy {
			for providerSecurityDefName, _ := range providerConfig.SecuritySchemaDefinitions {
				if operationSecurityDefName == providerSecurityDefName {
					return true, operationSecurityDefName
				}
			}
		}
	}
	return false, ""
}

func (r ResourceFactory) prepareApiKeyAuthentication(operation *spec.Operation, providerConfig ProviderConfig, url string) (map[string]string, string) {
	if required, securityDefinitionName := r.authRequired(operation, providerConfig); required {
		headers := map[string]string{}
		if &providerConfig != nil {
			securitySchemaDefinition := providerConfig.SecuritySchemaDefinitions[securityDefinitionName]
			if &securitySchemaDefinition.ApiKeyHeader != nil {
				headers[securitySchemaDefinition.ApiKeyHeader.Name] = securitySchemaDefinition.ApiKeyHeader.Value
			} else if &securitySchemaDefinition.ApiKeyQuery != nil {
				url = fmt.Sprintf("%s?%s=%s", url, securitySchemaDefinition.ApiKeyQuery.Name, securitySchemaDefinition.ApiKeyQuery.Value)
			}
		}
		return headers, url
	}
	return nil, url
}
