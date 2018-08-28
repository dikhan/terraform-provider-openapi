package openapi

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/dikhan/http_goclient"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
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

func (r resourceFactory) create(resourceLocalData *schema.ResourceData, i interface{}) error {
	providerConfig := i.(providerConfig)
	input := r.createPayloadFromLocalStateData(resourceLocalData)
	responsePayload := map[string]interface{}{}

	resourceURL, err := r.resourceInfo.getResourceURL()
	if err != nil {
		return err
	}

	operation := r.resourceInfo.createPathInfo.Post

	reqContext, err := r.apiAuthenticator.prepareAuth(operation.ID, resourceURL, operation.Security, providerConfig)
	if err != nil {
		return err
	}

	reqContext.headers = r.appendOperationHeaders(operation, providerConfig, reqContext.headers)

	res, err := r.httpClient.PostJson(reqContext.url, reqContext.headers, input, &responsePayload)
	if err != nil {
		return err
	}

	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK, http.StatusCreated, http.StatusAccepted}); err != nil {
		return fmt.Errorf("POST %s failed: %s", resourceURL, err)
	}
	return r.updateLocalState(resourceLocalData, responsePayload)
}

func (r resourceFactory) read(resourceLocalData *schema.ResourceData, i interface{}) error {
	providerConfig := i.(providerConfig)
	remoteData, err := r.readRemote(resourceLocalData.Id(), providerConfig)
	if err != nil {
		return err
	}
	return r.updateStateWithPayloadData(remoteData, resourceLocalData)
}

func (r resourceFactory) readRemote(id string, providerConfig providerConfig) (map[string]interface{}, error) {
	var err error
	responsePayload := map[string]interface{}{}
	resourceIDURL, err := r.resourceInfo.getResourceIDURL(id)
	if err != nil {
		return nil, err
	}

	operation := r.resourceInfo.pathInfo.Get

	reqContext, err := r.apiAuthenticator.prepareAuth(operation.ID, resourceIDURL, operation.Security, providerConfig)
	if err != nil {
		return nil, err
	}

	reqContext.headers = r.appendOperationHeaders(operation, providerConfig, reqContext.headers)

	res, err := r.httpClient.Get(reqContext.url, reqContext.headers, &responsePayload)
	if err != nil {
		return nil, err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK}); err != nil {
		return nil, fmt.Errorf("GET %s failed: %s", resourceIDURL, err)
	}
	return responsePayload, nil
}

func (r resourceFactory) update(resourceLocalData *schema.ResourceData, i interface{}) error {
	providerConfig := i.(providerConfig)
	operation := r.resourceInfo.pathInfo.Put
	if operation == nil {
		return fmt.Errorf("%s resource does not support PUT opperation, check the swagger file exposed on '%s'", r.resourceInfo.name, r.resourceInfo.host)
	}
	input := r.createPayloadFromLocalStateData(resourceLocalData)
	responsePayload := map[string]interface{}{}

	if err := r.checkImmutableFields(resourceLocalData, providerConfig); err != nil {
		return err
	}

	resourceIDURL, err := r.resourceInfo.getResourceIDURL(resourceLocalData.Id())
	if err != nil {
		return err
	}

	reqContext, err := r.apiAuthenticator.prepareAuth(operation.ID, resourceIDURL, operation.Security, providerConfig)
	if err != nil {
		return err
	}

	reqContext.headers = r.appendOperationHeaders(operation, providerConfig, reqContext.headers)

	res, err := r.httpClient.PutJson(reqContext.url, reqContext.headers, input, &responsePayload)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK, http.StatusAccepted}); err != nil {
		return fmt.Errorf("UPDATE %s failed: %s", resourceIDURL, err)
	}
	return r.updateStateWithPayloadData(responsePayload, resourceLocalData)
}

func (r resourceFactory) delete(resourceLocalData *schema.ResourceData, i interface{}) error {
	providerConfig := i.(providerConfig)
	operation := r.resourceInfo.pathInfo.Delete
	if operation == nil {
		return fmt.Errorf("%s resource does not support DELETE opperation, check the swagger file exposed on '%s'", r.resourceInfo.name, r.resourceInfo.host)
	}
	resourceIDURL, err := r.resourceInfo.getResourceIDURL(resourceLocalData.Id())
	if err != nil {
		return err
	}

	reqContext, err := r.apiAuthenticator.prepareAuth(operation.ID, resourceIDURL, operation.Security, providerConfig)
	if err != nil {
		return err
	}

	reqContext.headers = r.appendOperationHeaders(operation, providerConfig, reqContext.headers)
	res, err := r.httpClient.Delete(reqContext.url, reqContext.headers)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}); err != nil {
		return fmt.Errorf("DELETE %s failed: %s", resourceIDURL, err)
	}
	return nil
}

// appendOperationHeaders returns a maps containing the headers passed in and adds whatever headers the operation requires. The values
// are retrieved from the provider configuration.
func (r resourceFactory) appendOperationHeaders(operation *spec.Operation, providerConfig providerConfig, headers map[string]string) map[string]string {
	if operation != nil {
		headerConfigProps := openapiutils.GetHeaderConfigurations(operation.Parameters)
		for headerConfigProp, headerConfiguration := range headerConfigProps {
			// Setting the actual name of the header with the value coming from the provider configuration
			headers[headerConfiguration.Name] = providerConfig.Headers[headerConfigProp]
		}
	}
	return headers
}

// setStateID sets the local resource's data ID with the newly identifier created in the POST API request. Refer to
// r.resourceInfo.getResourceIdentifier() for more info regarding what property is selected as the identifier.
func (r resourceFactory) setStateID(resourceLocalData *schema.ResourceData, payload map[string]interface{}) error {
	identifierProperty, err := r.resourceInfo.getResourceIdentifier()
	if err != nil {
		return err
	}
	if payload[identifierProperty] == nil {
		return fmt.Errorf("response object returned from the API is missing mandatory identifier property '%s'", identifierProperty)
	}

	switch payload[identifierProperty].(type) {
	case int:
		resourceLocalData.SetId(strconv.Itoa(payload[identifierProperty].(int)))
	case float64:
		resourceLocalData.SetId(strconv.Itoa(int(payload[identifierProperty].(float64))))
	default:
		resourceLocalData.SetId(payload[identifierProperty].(string))
	}
	return nil
}

// updateLocalState populates the state of the schema resource data with the payload data received from the POST API request
func (r resourceFactory) updateLocalState(resourceLocalData *schema.ResourceData, payload map[string]interface{}) error {
	err := r.setStateID(resourceLocalData, payload)
	if err != nil {
		return err
	}
	return r.updateStateWithPayloadData(payload, resourceLocalData)
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

func (r resourceFactory) checkImmutableFields(updatedResourceLocalData *schema.ResourceData, providerConfig providerConfig) error {
	var remoteData map[string]interface{}
	var err error
	if remoteData, err = r.readRemote(updatedResourceLocalData.Id(), providerConfig); err != nil {
		return err
	}
	for _, immutablePropertyName := range r.resourceInfo.getImmutableProperties() {
		if r.getResourceDataValue(immutablePropertyName, updatedResourceLocalData) != remoteData[immutablePropertyName] {
			// Rolling back data so tf values are not stored in the state file; otherwise terraform would store the
			// data inside the updated (*schema.ResourceData) in the state file
			r.updateStateWithPayloadData(remoteData, updatedResourceLocalData)
			return fmt.Errorf("property %s is immutable and therefore can not be updated. Update operation was aborted; no updates were performed", immutablePropertyName)
		}
	}
	return nil
}

// updateStateWithPayloadData is in charge of saving the given payload into the state file. The property names are
// converted into compliant terraform names if needed.
func (r resourceFactory) updateStateWithPayloadData(remoteData map[string]interface{}, resourceLocalData *schema.ResourceData) error {
	for propertyName, propertyValue := range remoteData {
		if r.resourceInfo.isIDProperty(propertyName) {
			continue
		}
		if err := r.setResourceDataProperty(propertyName, propertyValue, resourceLocalData); err != nil {
			return err
		}
	}
	return nil
}

// createPayloadFromLocalStateData is in charge of translating the values saved in the local state into a payload that can be posted/put
// to the API. Note that when reading the properties from the schema definition, there's a conversion to a compliant
// will automatically translate names into terraform compatible names that can be saved in the state file; otherwise
// terraform name so the look up in the local state operation works properly. The property names saved in the local state
// are alaways converted to terraform compatible names
func (r resourceFactory) createPayloadFromLocalStateData(resourceLocalData *schema.ResourceData) map[string]interface{} {
	input := map[string]interface{}{}
	for propertyName, property := range r.resourceInfo.schemaDefinition.Properties {
		// ReadOnly properties are not considered for the payload data
		if r.resourceInfo.isIDProperty(propertyName) || property.ReadOnly {
			continue
		}
		if dataValue, ok := r.getResourceDataOK(propertyName, resourceLocalData); ok {
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
		} else {
			// Special case to handle changes for integer properties that are set to 0 when creating a resource or are updated to 0 value
			// (d *ResourceData) GetOk(key string) terraform function ignores integer properties
			old, new := r.getResourceDataChange(propertyName, resourceLocalData)
			switch reflect.TypeOf(dataValue).Kind() {
			case reflect.String:
				if new == "" {
					input[propertyName] = new.(string)
				}
			case reflect.Int:
				if new == 0 {
					input[propertyName] = new.(int)
				}
			case reflect.Float64:
				if new == 0.0 {
					input[propertyName] = new.(float64)
				}
			case reflect.Bool:
				if new == false {
					input[propertyName] = new.(bool)
				}
			}
			log.Printf("[DEBUG] createPayloadFromLocalStateData [%s] - oldvalue[%+v]", propertyName, old)
		}
		log.Printf("[DEBUG] createPayloadFromLocalStateData [%s] - newValue[%+v]", propertyName, input[propertyName])
	}
	return input
}

// getResourceDataValue returns the data for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) getResourceDataValue(schemaDefinitionPropertyName string, resourceLocalData *schema.ResourceData) interface{} {
	dataPropertyName := terraformutils.ConvertToTerraformCompliantName(schemaDefinitionPropertyName)
	return resourceLocalData.Get(dataPropertyName)
}

// getResourceDataOK returns the data for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) getResourceDataOK(schemaDefinitionPropertyName string, resourceLocalData *schema.ResourceData) (interface{}, bool) {
	dataPropertyName := terraformutils.ConvertToTerraformCompliantName(schemaDefinitionPropertyName)
	return resourceLocalData.GetOk(dataPropertyName)
}

// getResourceDataChange returns the old and new value for a given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) getResourceDataChange(schemaDefinitionPropertyName string, resourceLocalData *schema.ResourceData) (interface{}, interface{}) {
	dataPropertyName := terraformutils.ConvertToTerraformCompliantName(schemaDefinitionPropertyName)
	return resourceLocalData.GetChange(dataPropertyName)
}

// setResourceDataProperty sets the value for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) setResourceDataProperty(schemaDefinitionPropertyName string, value interface{}, resourceLocalData *schema.ResourceData) error {
	dataPropertyName := terraformutils.ConvertToTerraformCompliantName(schemaDefinitionPropertyName)
	return resourceLocalData.Set(dataPropertyName, value)
}
