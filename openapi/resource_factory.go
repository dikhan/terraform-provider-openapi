package openapi

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/dikhan/http_goclient"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapierr"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type resourceFactory struct {
	httpClient       http_goclient.HttpClient
	resourceInfo     resourceInfo
	apiAuthenticator apiAuthenticator
}

// only applicable when remote resource no longer exists and GET operations return 404 NotFound
var defaultDestroyStatus = "destroyed"

var defaultTimeout = time.Duration(10 * time.Minute)

func (r resourceFactory) createSchemaResource() (*schema.Resource, error) {
	s, err := r.resourceInfo.createTerraformResourceSchema()
	if err != nil {
		return nil, err
	}
	timeouts, err := r.createSchemaResourceTimeout()
	if err != nil {
		return nil, err
	}
	return &schema.Resource{
		Schema:   s,
		Create:   r.create,
		Read:     r.read,
		Delete:   r.delete,
		Update:   r.update,
		Timeouts: timeouts,
	}, nil
}

func (r resourceFactory) createSchemaResourceTimeout() (*schema.ResourceTimeout, error) {
	var postTimeout *time.Duration
	var getTimeout *time.Duration
	var putTimeout *time.Duration
	var deleteTimeout *time.Duration
	var err error
	if postTimeout, err = r.resourceInfo.getResourceTimeout(r.resourceInfo.createPathInfo.Post); err != nil {
		return nil, err
	}
	if getTimeout, err = r.resourceInfo.getResourceTimeout(r.resourceInfo.pathInfo.Get); err != nil {
		return nil, err
	}
	if putTimeout, err = r.resourceInfo.getResourceTimeout(r.resourceInfo.pathInfo.Put); err != nil {
		return nil, err
	}
	if deleteTimeout, err = r.resourceInfo.getResourceTimeout(r.resourceInfo.pathInfo.Delete); err != nil {
		return nil, err
	}
	return &schema.ResourceTimeout{
		Create:  postTimeout,
		Read:    getTimeout,
		Update:  putTimeout,
		Delete:  deleteTimeout,
		Default: &defaultTimeout,
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

	err = r.setStateID(resourceLocalData, responsePayload)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Resource '%s' ID: %s", r.resourceInfo.name, resourceLocalData.Id())

	err = r.handlePollingIfConfigured(&responsePayload, resourceLocalData, providerConfig, operation.Responses, res.StatusCode, schema.TimeoutCreate)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after POST %s call with response status code (%d): %s", resourceURL, res.StatusCode, err)
	}
	return r.updateStateWithPayloadData(responsePayload, resourceLocalData)
}

func (r resourceFactory) read(resourceLocalData *schema.ResourceData, i interface{}) error {
	providerConfig := i.(providerConfig)
	remoteData, err := r.readRemote(resourceLocalData.Id(), providerConfig)

	if err != nil {
		if openapiErr, ok := err.(openapierr.Error); ok {
			if openapierr.NotFound == openapiErr.Code() {
				return nil
			}
		}
		return fmt.Errorf("GET %s/%s failed: %s", r.resourceInfo.path, resourceLocalData.Id(), err)
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
		return nil, err
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

	err = r.handlePollingIfConfigured(&responsePayload, resourceLocalData, providerConfig, operation.Responses, res.StatusCode, schema.TimeoutUpdate)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after PUT %s call with response status code (%d): %s", resourceIDURL, res.StatusCode, err)
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
	if err := r.checkHTTPStatusCode(res, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted, http.StatusNotFound}); err != nil {
		return fmt.Errorf("DELETE %s failed: %s", resourceIDURL, err)
	}

	err = r.handlePollingIfConfigured(nil, resourceLocalData, providerConfig, operation.Responses, res.StatusCode, schema.TimeoutDelete)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after DELETE %s call with response status code (%d): %s", resourceIDURL, res.StatusCode, err)
	}

	return nil
}

func (r resourceFactory) handlePollingIfConfigured(responsePayload *map[string]interface{}, resourceLocalData *schema.ResourceData, providerConfig providerConfig, responses *spec.Responses, responseStatusCode int, timeoutFor string) error {
	pollingEnabled, response := r.resourceInfo.isResourcePollingEnabled(responses, responseStatusCode)

	if !pollingEnabled {
		return nil
	}

	// This is a use case where payload does not contain payload data and hence status field is not available; e,g: DELETE operations
	// The default behaviour for this case is to consider the resource as destroyed. Hence, the below code pre-populates
	// the target extension with the expected status that the polling mechanism expects when dealing with NotFound resources (should only happen on delete operations).
	// Since this is internal behaviour it is not expected that the service provider will populate this field; and if so, it
	// will be overridden
	if responsePayload == nil {
		if value, exists := response.Extensions.GetString(extTfResourcePollTargetStatuses); exists {
			log.Printf("[WARN] service provider speficied '%s': %s for a DELETE operation. This is not expected as the normal behaviour is the resource to no longer exists once the DELETE operation is completed; hence subsequent GET calls should return 404 NotFound instead", extTfResourcePollTargetStatuses, value)
		}
		log.Printf("[WARN] setting extension '%s' with default value '%s'", extTfResourcePollTargetStatuses, defaultDestroyStatus)
		response.Extensions.Add(extTfResourcePollTargetStatuses, defaultDestroyStatus)
	}

	targetStatuses, err := r.resourceInfo.getResourcePollTargetStatuses(*response)
	if err != nil {
		return err
	}

	pendingStatuses, err := r.resourceInfo.getResourcePollPendingStatuses(*response)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] target statuses (%s); pending statuses (%s)", targetStatuses, pendingStatuses)
	log.Printf("[INFO] Waiting for resource '%s' to reach a completion status (%s)", r.resourceInfo.name, targetStatuses)

	stateConf := &resource.StateChangeConf{
		Pending:      pendingStatuses,
		Target:       targetStatuses,
		Refresh:      r.resourceStateRefreshFunc(resourceLocalData, providerConfig),
		Timeout:      resourceLocalData.Timeout(timeoutFor),
		PollInterval: 5 * time.Second,
		MinTimeout:   10 * time.Second,
		Delay:        1 * time.Second,
	}

	// Wait, catching any errors
	remoteData, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for resource to reach a completion status (%s) [valid pending statuses (%s)]: %s", targetStatuses, pendingStatuses, err)
	}
	if responsePayload != nil {
		*responsePayload = remoteData.(map[string]interface{})
	}
	return nil
}

func (r resourceFactory) resourceStateRefreshFunc(resourceLocalData *schema.ResourceData, providerConfig providerConfig) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		remoteData, err := r.readRemote(resourceLocalData.Id(), providerConfig)

		if err != nil {
			if openapiErr, ok := err.(openapierr.Error); ok {
				if openapierr.NotFound == openapiErr.Code() {
					return remoteData, defaultDestroyStatus, nil
				}
			}
			return nil, "", fmt.Errorf("error on retrieving resource '%s' (%s) when waiting: %s", r.resourceInfo.name, resourceLocalData.Id(), err)
		}

		statusIdentifier, err := r.resourceInfo.getStatusIdentifier()
		if err != nil {
			return nil, "", fmt.Errorf("error occurred while retrieving status identifier for resource '%s' (%s): %s", r.resourceInfo.name, resourceLocalData.Id(), err)
		}

		value, statusIdentifierPresentInResponse := remoteData[statusIdentifier]
		if !statusIdentifierPresentInResponse {
			return nil, "", fmt.Errorf("response payload received from GET /%s/%s  missing the status identifier field", r.resourceInfo.path, resourceLocalData.Id())
		}
		newStatus := value.(string)
		log.Printf("[DEBUG] resource '%s' status (%s): %s", r.resourceInfo.name, resourceLocalData.Id(), newStatus)
		return remoteData, newStatus, nil
	}
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
			return fmt.Errorf("HTTP Reponse Status Code %d - Unauthorized. API access is denied due to invalid credentials: %s", res.StatusCode, resBody)
		case http.StatusNotFound:
			return &openapierr.NotFoundError{OriginalError: fmt.Errorf("HTTP Reponse Status Code %d - Not Found. Could not find resource instance: %s", res.StatusCode, resBody)}
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
		if localValue, exists := r.getResourceDataOKExists(immutablePropertyName, updatedResourceLocalData); exists {
			if localValue != remoteData[immutablePropertyName] {
				// Rolling back data so tf values are not stored in the state file; otherwise terraform would store the
				// data inside the updated (*schema.ResourceData) in the state file
				r.updateStateWithPayloadData(remoteData, updatedResourceLocalData)
				return fmt.Errorf("property %s is immutable and therefore can not be updated. Update operation was aborted; no updates were performed", immutablePropertyName)
			}
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
		if dataValue, ok := r.getResourceDataOKExists(propertyName, resourceLocalData); ok {
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
		log.Printf("[DEBUG] createPayloadFromLocalStateData [%s] - newValue[%+v]", propertyName, input[propertyName])
	}
	return input
}

// getResourceDataOK returns the data for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) getResourceDataOKExists(schemaDefinitionPropertyName string, resourceLocalData *schema.ResourceData) (interface{}, bool) {
	schemaDefinitionProperty := r.resourceInfo.schemaDefinition.Properties[schemaDefinitionPropertyName]
	dataPropertyName := r.resourceInfo.convertToTerraformCompliantFieldName(schemaDefinitionPropertyName, schemaDefinitionProperty)
	return resourceLocalData.GetOkExists(dataPropertyName)
}

// setResourceDataProperty sets the value for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) setResourceDataProperty(schemaDefinitionPropertyName string, value interface{}, resourceLocalData *schema.ResourceData) error {
	schemaDefinitionProperty := r.resourceInfo.schemaDefinition.Properties[schemaDefinitionPropertyName]
	dataPropertyName := r.resourceInfo.convertToTerraformCompliantFieldName(schemaDefinitionPropertyName, schemaDefinitionProperty)
	return resourceLocalData.Set(dataPropertyName, value)
}
