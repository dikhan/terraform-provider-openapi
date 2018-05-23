package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/dikhan/http_goclient"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"strconv"
)

type resourceFactory struct {
	httpClient       http_goclient.HttpClient
	resourceInfo     resourceInfo
	apiAuthenticator apiAuthenticator
}

func (r resourceFactory) createSchemaResource() (*schema.Resource, error) {
	s, err := r.resourceInfo.createTerraformResourceSchema()
	if err != nil {
		return nil, err
	}
	return &schema.Resource{
		Schema: s,
		Create: r.create,
		Read:   r.read,
		Delete: r.delete,
		Update: r.update,
	}, nil
}

func (r resourceFactory) create(data *schema.ResourceData, i interface{}) error {
	input := r.getPayloadFromData(data)
	responsePayload := map[string]interface{}{}

	resourceURL, err := r.resourceInfo.getResourceURL()
	if err != nil {
		return err
	}

	authContext, err := r.apiAuthenticator.prepareAuth(r.resourceInfo.createPathInfo.Post.ID, resourceURL, r.resourceInfo.createPathInfo.Post.Security, i.(providerConfig))
	if err != nil {
		return err
	}

	res, err := r.httpClient.PostJson(authContext.url, authContext.headers, input, &responsePayload)
	if err != nil {
		return err
	}

	if err := r.checkHTTPStatusCode(res, []int{http.StatusCreated, http.StatusAccepted}); err != nil {
		return fmt.Errorf("POST %s failed: %s", resourceURL, err)
	}
	return r.updateLocalState(data, responsePayload)
}

func (r resourceFactory) read(data *schema.ResourceData, i interface{}) error {
	output, err := r.readRemote(data.Id(), i.(providerConfig))
	if err != nil {
		return err
	}
	return r.updateStateWithPayloadData(output, data)
}

func (r resourceFactory) readRemote(id string, config providerConfig) (map[string]interface{}, error) {
	var err error
	responsePayload := map[string]interface{}{}
	resourceIDURL, err := r.resourceInfo.getResourceIDURL(id)
	if err != nil {
		return nil, err
	}

	authContext, err := r.apiAuthenticator.prepareAuth(r.resourceInfo.pathInfo.Get.ID, resourceIDURL, r.resourceInfo.pathInfo.Get.Security, config)
	if err != nil {
		return nil, err
	}
	res, err := r.httpClient.Get(authContext.url, authContext.headers, &responsePayload)
	if err != nil {
		return nil, err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK}); err != nil {
		return nil, fmt.Errorf("GET %s failed: %s", resourceIDURL, err)
	}
	return responsePayload, nil
}

func (r resourceFactory) update(data *schema.ResourceData, i interface{}) error {
	if r.resourceInfo.pathInfo.Put == nil {
		return fmt.Errorf("%s resource does not support PUT opperation, check the swagger file exposed on '%s'", r.resourceInfo.name, r.resourceInfo.host)
	}
	input := r.getPayloadFromData(data)
	responsePayload := map[string]interface{}{}

	if err := r.checkImmutableFields(data, i); err != nil {
		return err
	}

	resourceIDURL, err := r.resourceInfo.getResourceIDURL(data.Id())
	if err != nil {
		return err
	}

	authContext, err := r.apiAuthenticator.prepareAuth(r.resourceInfo.pathInfo.Put.ID, resourceIDURL, r.resourceInfo.pathInfo.Put.Security, i.(providerConfig))
	if err != nil {
		return err
	}

	res, err := r.httpClient.PutJson(authContext.url, authContext.headers, input, &responsePayload)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK}); err != nil {
		return fmt.Errorf("UPDATE %s failed: %s", resourceIDURL, err)
	}
	return r.updateStateWithPayloadData(responsePayload, data)
}

func (r resourceFactory) delete(data *schema.ResourceData, i interface{}) error {
	if r.resourceInfo.pathInfo.Delete == nil {
		return fmt.Errorf("%s resource does not support DELETE opperation, check the swagger file exposed on '%s'", r.resourceInfo.name, r.resourceInfo.host)
	}
	resourceIDURL, err := r.resourceInfo.getResourceIDURL(data.Id())
	if err != nil {
		return err
	}

	authContext, err := r.apiAuthenticator.prepareAuth(r.resourceInfo.pathInfo.Delete.ID, resourceIDURL, r.resourceInfo.pathInfo.Delete.Security, i.(providerConfig))
	if err != nil {
		return err
	}

	res, err := r.httpClient.Delete(authContext.url, authContext.headers)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusNoContent}); err != nil {
		return fmt.Errorf("DELETE %s failed: %s", resourceIDURL, err)
	}
	return nil
}

// setStateID sets the local resource's data ID with the newly identifier created in the POST API request. Refer to
// r.resourceInfo.getResourceIdentifier() for more info regarding what property is selected as the identifier.
func (r resourceFactory) setStateID(data *schema.ResourceData, payload map[string]interface{}) error {
	identifierProperty, err := r.resourceInfo.getResourceIdentifier()
	if err != nil {
		return err
	}
	if payload[identifierProperty] == nil {
		return fmt.Errorf("response object returned from the API is missing mandatory identifier property '%s'", identifierProperty)
	}

	switch payload[identifierProperty].(type) {
	case int:
		data.SetId(strconv.Itoa(payload[identifierProperty].(int)))
	case float64:
		data.SetId(strconv.Itoa(int(payload[identifierProperty].(float64))))
	default:
		data.SetId(payload[identifierProperty].(string))
	}
	return nil
}

// updateLocalState populates the state of the schema resource data with the payload data received from the POST API request
func (r resourceFactory) updateLocalState(data *schema.ResourceData, payload map[string]interface{}) error {
	err := r.setStateID(data, payload)
	if err != nil {
		return err
	}
	return r.updateStateWithPayloadData(payload, data)
}

func (r resourceFactory) checkHTTPStatusCode(res *http.Response, expectedHTTPStatusCodes []int) error {
	if !responseContainsExpectedStatus(expectedHTTPStatusCodes, res.StatusCode) {
		var resBody string
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("HTTP Reponse Status Code %d - Error '%s' occurred while reading the response body", res.StatusCode, err)
		}
		if len(b) > 0 {
			resBody = string(b)
		}
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("HTTP Reponse Status Code %d - Unauthorized: API access is denied due to invalid credentials (%s)", res.StatusCode, resBody)
		default:
			return fmt.Errorf("HTTP Reponse Status Code %d not matching expected one %v (%s)", res.StatusCode, expectedHTTPStatusCodes, resBody)
		}
	}
	return nil
}

func (r resourceFactory) checkImmutableFields(updated *schema.ResourceData, i interface{}) error {
	var remoteData map[string]interface{}
	var err error
	if remoteData, err = r.readRemote(updated.Id(), i.(providerConfig)); err != nil {
		return err
	}
	for _, immutablePropertyName := range r.resourceInfo.getImmutableProperties() {
		if updated.Get(immutablePropertyName) != remoteData[immutablePropertyName] {
			// Rolling back data so tf values are not stored in the state file; otherwise terraform would store the
			// data inside the updated (*schema.ResourceData) in the state file
			r.updateStateWithPayloadData(remoteData, updated)
			return fmt.Errorf("property %s is immutable and therefore can not be updated. Update operation was aborted; no updates were performed", immutablePropertyName)
		}
	}
	return nil
}

func (r resourceFactory) updateStateWithPayloadData(input map[string]interface{}, data *schema.ResourceData) error {
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

func (r resourceFactory) getPayloadFromData(data *schema.ResourceData) map[string]interface{} {
	input := map[string]interface{}{}
	for propertyName, property := range r.resourceInfo.schemaDefinition.Properties {
		// ReadOnly properties are not considered for the payload data
		if propertyName == "id" || property.ReadOnly {
			continue
		}
		if dataValue, ok := data.GetOk(propertyName); ok {
			switch reflect.TypeOf(dataValue).Kind() {
			case reflect.Slice:
				input[propertyName] = dataValue.([]interface{})
			case reflect.String:
				input[propertyName] = dataValue.(string)
			case reflect.Int:
				input[propertyName] = dataValue.(int)
			case reflect.Float64:
				input[propertyName] = dataValue.(float64)
			case reflect.Bool:
				input[propertyName] = dataValue.(bool)
			}
		}

	}
	return input
}
