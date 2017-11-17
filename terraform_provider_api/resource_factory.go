package main

import (
	"fmt"
	"net/http"
	"reflect"

	"io/ioutil"

	"github.com/dikhan/http_goclient"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type ResourceFactory struct {
	httpClient   http_goclient.HttpClient
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
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("HTTP Reponse Status Code %d - Unauthorized: API access is denied due to invalid credentials", res.StatusCode)
		default:
			b, _ := ioutil.ReadAll(res.Body)
			if len(b) > 0 {
				return fmt.Errorf("HTTP Reponse Status Code %d not matching expected one %d. Error = %s", res.StatusCode, expectedHttpStatusCode, string(b))
			}
			return fmt.Errorf("HTTP Reponse Status Code %d not matching expected one %d", res.StatusCode, expectedHttpStatusCode)
		}

	}
	return nil
}

func (r ResourceFactory) create(data *schema.ResourceData, i interface{}) error {
	input := r.getPayloadFromData(data)
	output := map[string]interface{}{}

	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.CreatePathInfo.Post, i.(ProviderConfig), r.ResourceInfo.getResourceUrl())
	res, err := r.httpClient.PostJson(url, headers, input, &output)
	if err != nil {
		return err
	}
	if err := r.checkHttpStatusCode(res, http.StatusCreated); err != nil {
		return fmt.Errorf("POST %s failed: %s", url, err)
	}

	if output["id"] == nil {
		return fmt.Errorf("object returned from api is missing mandatory property 'id'")
	}

	data.SetId(output["id"].(string))
	return nil
}

func (r ResourceFactory) read(data *schema.ResourceData, i interface{}) error {
	output, err := r.readRemote(data.Id(), i.(ProviderConfig))
	if err != nil {
		return err
	}
	return r.updateResourceState(output, data)
}

func (r ResourceFactory) readRemote(id string, config ProviderConfig) (map[string]interface{}, error) {
	output := map[string]interface{}{}
	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.PathInfo.Get, config, r.ResourceInfo.getResourceIdUrl(id))
	res, err := r.httpClient.Get(url, headers, &output)
	if err != nil {
		return nil, err
	}
	if err := r.checkHttpStatusCode(res, http.StatusOK); err != nil {
		return nil, fmt.Errorf("GET %s failed: %s", url, err)
	}
	return output, nil
}

func (r ResourceFactory) update(data *schema.ResourceData, i interface{}) error {
	input := r.getPayloadFromData(data)
	output := map[string]interface{}{}

	if err := r.checkImmutableFields(data, i); err != nil {
		return err
	}

	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.PathInfo.Put, i.(ProviderConfig), r.ResourceInfo.getResourceIdUrl(data.Id()))
	res, err := r.httpClient.PutJson(url, headers, input, &output)
	if err != nil {
		return err
	}
	if err := r.checkHttpStatusCode(res, http.StatusOK); err != nil {
		return fmt.Errorf("UPDATE %s failed: %s", url, err)
	}
	return r.updateResourceState(output, data)
}

func (r ResourceFactory) delete(data *schema.ResourceData, i interface{}) error {
	headers, url := r.prepareApiKeyAuthentication(r.ResourceInfo.PathInfo.Delete, i.(ProviderConfig), r.ResourceInfo.getResourceIdUrl(data.Id()))
	res, err := r.httpClient.Delete(url, headers)
	if err != nil {
		return err
	}
	if err := r.checkHttpStatusCode(res, http.StatusNoContent); err != nil {
		return fmt.Errorf("DELETE %s failed: %s", url, err)
	}
	return nil
}

func (r ResourceFactory) checkImmutableFields(updated *schema.ResourceData, i interface{}) error {
	var remoteData map[string]interface{}
	var err error
	if remoteData, err = r.readRemote(updated.Id(), i.(ProviderConfig)); err != nil {
		return err
	}
	for _, immutablePropertyName := range r.ResourceInfo.getImmutableProperties() {
		if updated.Get(immutablePropertyName) != remoteData[immutablePropertyName] {
			// Rolling back data so tf values are not stored in the state file; otherwise terraform would store the
			// data inside the updated (*schema.ResourceData) in the state file
			r.updateResourceState(remoteData, updated)
			return fmt.Errorf("property %s is immutable and therefore can not be updated. Update operation was aborted; no updates were performed", immutablePropertyName)
		}
	}
	return nil
}

func (r ResourceFactory) updateResourceState(input map[string]interface{}, data *schema.ResourceData) error {
	for propertyName, propertyValue := range input {
		if propertyName == "id" {
			continue
		}
		if err := data.Set(propertyName, propertyValue); err != nil {
			return err
		}
	}
	return nil
}

func (r ResourceFactory) getPayloadFromData(data *schema.ResourceData) map[string]interface{} {
	input := map[string]interface{}{}
	for propertyName, property := range r.ResourceInfo.SchemaDefinition.Properties {
		// ReadOnly properties are not considered for the payload data
		if propertyName == "id" || property.ReadOnly {
			continue
		}
		switch reflect.TypeOf(data.Get(propertyName)).Kind() {
		case reflect.Slice:
			input[propertyName] = data.Get(propertyName).([]interface{})
		case reflect.String:
			input[propertyName] = data.Get(propertyName).(string)
		case reflect.Int:
			input[propertyName] = data.Get(propertyName).(int)
		case reflect.Float64:
			input[propertyName] = data.Get(propertyName).(float64)
		case reflect.Bool:
			input[propertyName] = data.Get(propertyName).(bool)
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
