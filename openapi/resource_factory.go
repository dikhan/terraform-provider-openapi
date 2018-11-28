package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapierr"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type resourceFactory struct {
	openAPIResource       SpecResource
	defaultTimeout        time.Duration
	defaultPollInterval   time.Duration
	defaultPollMinTimeout time.Duration
	defaultPollDelay      time.Duration
}

// only applicable when remote resource no longer exists and GET operations return 404 NotFound
const defaultDestroyStatus = "destroyed"

var defaultPollInterval = time.Duration(5 * time.Second)
var defaultPollMinTimeout = time.Duration(10 * time.Second)
var defaultPollDelay = time.Duration(1 * time.Second)
var defaultTimeout = time.Duration(10 * time.Minute)

func newResourceFactory(openAPIResource SpecResource) resourceFactory {
	return resourceFactory{
		openAPIResource:       openAPIResource,
		defaultPollDelay:      defaultPollDelay,
		defaultPollInterval:   defaultPollInterval,
		defaultPollMinTimeout: defaultPollMinTimeout,
		defaultTimeout:        defaultTimeout,
	}
}

func (r resourceFactory) createTerraformResource() (*schema.Resource, error) {
	s, err := r.createTerraformResourceSchema()
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
		Importer: r.importer(),
		Timeouts: timeouts,
	}, nil
}

func (r resourceFactory) createSchemaResourceTimeout() (*schema.ResourceTimeout, error) {
	var timeouts *specTimeouts
	var err error
	if timeouts, err = r.openAPIResource.getTimeouts(); err != nil {
		return nil, err
	}
	return &schema.ResourceTimeout{
		Create:  timeouts.Post,
		Read:    timeouts.Get,
		Update:  timeouts.Put,
		Delete:  timeouts.Delete,
		Default: &r.defaultTimeout,
	}, nil
}

func (r resourceFactory) createTerraformResourceSchema() (map[string]*schema.Schema, error) {
	schemaDefinition, err := r.openAPIResource.getResourceSchema()
	if err != nil {
		return nil, err
	}
	return schemaDefinition.createResourceSchema()
}

func (r resourceFactory) create(data *schema.ResourceData, i interface{}) error {
	providerClient := i.(ClientOpenAPI)
	operation := r.openAPIResource.getResourceOperations().Post
	requestPayload := r.createPayloadFromLocalStateData(data)
	responsePayload := map[string]interface{}{}
	res, err := providerClient.Post(r.openAPIResource, requestPayload, &responsePayload)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK, http.StatusCreated, http.StatusAccepted}); err != nil {
		return fmt.Errorf("[resource='%s'] POST %s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), err)
	}

	err = r.setStateID(data, responsePayload)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Resource '%s' ID: %s", r.openAPIResource.getResourcePath(), data.Id())

	err = r.handlePollingIfConfigured(&responsePayload, data, providerClient, operation, res.StatusCode, schema.TimeoutCreate)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after POST %s call with response status code (%d): %s", r.openAPIResource.getResourcePath(), res.StatusCode, err)
	}

	return r.updateStateWithPayloadData(responsePayload, data)
}

func (r resourceFactory) read(data *schema.ResourceData, i interface{}) error {
	openAPIClient := i.(ClientOpenAPI)
	remoteData, err := r.readRemote(data.Id(), openAPIClient)

	if err != nil {
		if openapiErr, ok := err.(openapierr.Error); ok {
			if openapierr.NotFound == openapiErr.Code() {
				return nil
			}
		}
		return fmt.Errorf("[resource='%s'] GET %s/%s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), data.Id(), err)
	}

	return r.updateStateWithPayloadData(remoteData, data)
}

func (r resourceFactory) readRemote(id string, providerClient ClientOpenAPI) (map[string]interface{}, error) {
	var err error
	responsePayload := map[string]interface{}{}
	resp, err := providerClient.Get(r.openAPIResource, id, &responsePayload)
	if err != nil {
		return nil, err
	}

	if err := r.checkHTTPStatusCode(resp, []int{http.StatusOK}); err != nil {
		return nil, err
	}

	return responsePayload, nil
}

func (r resourceFactory) update(data *schema.ResourceData, i interface{}) error {
	providerClient := i.(ClientOpenAPI)
	operation := r.openAPIResource.getResourceOperations().Put
	if operation == nil {
		return fmt.Errorf("[resource='%s'] resource does not support PUT operation, check the swagger file exposed on '%s'", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath())
	}
	requestPayload := r.createPayloadFromLocalStateData(data)
	responsePayload := map[string]interface{}{}
	if err := r.checkImmutableFields(data, providerClient); err != nil {
		return err
	}
	res, err := providerClient.Put(r.openAPIResource, data.Id(), requestPayload, &responsePayload)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK, http.StatusAccepted}); err != nil {
		return fmt.Errorf("[resource='%s'] UPDATE %s/%s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), data.Id(), err)
	}

	err = r.handlePollingIfConfigured(&responsePayload, data, providerClient, operation, res.StatusCode, schema.TimeoutUpdate)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after PUT %s call with response status code (%d): %s", r.openAPIResource.getResourcePath(), res.StatusCode, err)
	}

	return r.updateStateWithPayloadData(responsePayload, data)
}

func (r resourceFactory) delete(data *schema.ResourceData, i interface{}) error {
	providerClient := i.(ClientOpenAPI)
	operation := r.openAPIResource.getResourceOperations().Delete
	if operation == nil {
		return fmt.Errorf("[resource='%s'] resource does not support DELETE operation, check the swagger file exposed on '%s'", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath())
	}
	res, err := providerClient.Delete(r.openAPIResource, data.Id())
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}); err != nil {
		return fmt.Errorf("[resource='%s'] DELETE %s/%s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), data.Id(), err)
	}

	err = r.handlePollingIfConfigured(nil, data, providerClient, operation, res.StatusCode, schema.TimeoutDelete)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after DELETE %s call with response status code (%d): %s", r.openAPIResource.getResourcePath(), res.StatusCode, err)
	}

	return nil
}

func (r resourceFactory) importer() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		State: func(data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
			results := make([]*schema.ResourceData, 1, 1)
			results[0] = data
			err := r.read(data, i)
			return results, err
		},
	}
}

func (r resourceFactory) handlePollingIfConfigured(responsePayload *map[string]interface{}, resourceLocalData *schema.ResourceData, providerClient ClientOpenAPI, operation *specResourceOperation, responseStatusCode int, timeoutFor string) error {
	response := operation.responses.getResponse(responseStatusCode)

	if response == nil || !response.isPollingEnabled {
		return nil
	}

	targetStatuses := response.pollTargetStatuses
	pendingStatuses := response.pollPendingStatuses

	// This is a use case where payload does not contain payload data and hence status field is not available; e,g: DELETE operations
	// The default behaviour for this case is to consider the resource as destroyed. Hence, the below code pre-populates
	// the target extension with the expected status that the polling mechanism expects when dealing with NotFound resources (should only happen on delete operations).
	// Since this is internal behaviour it is not expected that the service provider will populate this field; and if so, it
	// will be overridden
	if responsePayload == nil {
		if len(targetStatuses) > 0 {
			log.Printf("[WARN] resource speficied poll target statuses for a DELETE operation. This is not expected as the normal behaviour is the resource to no longer exists once the DELETE operation is completed; hence subsequent GET calls should return 404 NotFound instead")
		}
		log.Printf("[WARN] overriding target status with default destroy status")
		targetStatuses = []string{defaultDestroyStatus}
	}

	log.Printf("[DEBUG] target statuses (%s); pending statuses (%s)", targetStatuses, pendingStatuses)
	log.Printf("[INFO] Waiting for resource '%s' to reach a completion status (%s)", r.openAPIResource.getResourcePath(), targetStatuses)

	stateConf := &resource.StateChangeConf{
		Pending:      pendingStatuses,
		Target:       targetStatuses,
		Refresh:      r.resourceStateRefreshFunc(resourceLocalData, providerClient),
		Timeout:      resourceLocalData.Timeout(timeoutFor),
		PollInterval: r.defaultPollInterval,
		MinTimeout:   r.defaultPollMinTimeout,
		Delay:        r.defaultPollDelay,
	}

	// Wait, catching any errors
	remoteData, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for resource to reach a completion status (%s) [valid pending statuses (%s)]: %s", targetStatuses, pendingStatuses, err)
	}
	if responsePayload != nil {
		remoteDataCasted, ok := remoteData.(map[string]interface{})
		if ok {
			*responsePayload = remoteDataCasted
		} else {
			return fmt.Errorf("failed to convert remote data (%s) to map[string]interface{}", reflect.TypeOf(remoteData))
		}
	}
	return nil
}

func (r resourceFactory) resourceStateRefreshFunc(resourceLocalData *schema.ResourceData, providerClient ClientOpenAPI) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		remoteData, err := r.readRemote(resourceLocalData.Id(), providerClient)

		if err != nil {
			if openapiErr, ok := err.(openapierr.Error); ok {
				if openapierr.NotFound == openapiErr.Code() {
					return 0, defaultDestroyStatus, nil
				}
			}
			return nil, "", fmt.Errorf("error on retrieving resource '%s' (%s) when waiting: %s", r.openAPIResource.getResourcePath(), resourceLocalData.Id(), err)
		}

		newStatus, err := r.getStatusValueFromPayload(remoteData)
		if err != nil {
			return nil, "", fmt.Errorf("error occurred while retrieving status identifier value from payload for resource '%s' (%s): %s", r.openAPIResource.getResourcePath(), resourceLocalData.Id(), err)
		}
		log.Printf("[DEBUG] resource '%s' status (%s): %s", r.openAPIResource.getResourcePath(), resourceLocalData.Id(), newStatus)
		return remoteData, newStatus, nil
	}
}

// setStateID sets the local resource's data ID with the newly identifier created in the POST API request. Refer to
// r.resourceInfo.getResourceIdentifier() for more info regarding what property is selected as the identifier.
func (r resourceFactory) setStateID(resourceLocalData *schema.ResourceData, payload map[string]interface{}) error {
	resourceSchema, err := r.openAPIResource.getResourceSchema()
	if err != nil {
		return err
	}
	identifierProperty, err := resourceSchema.getResourceIdentifier()
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] payload = %+v", payload)
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

func (r resourceFactory) checkHTTPStatusCode(res *http.Response, expectedHTTPStatusCodes []int) error {
	if !r.responseContainsExpectedStatus(expectedHTTPStatusCodes, res.StatusCode) {
		var resBody string
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("[resource='%s'] HTTP Reponse Status Code %d - Error '%s' occurred while reading the response body", r.openAPIResource.getResourceName(), res.StatusCode, err)
		}
		if b != nil && len(b) > 0 {
			resBody = string(b)
		}
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d - Unauthorized: API access is denied due to invalid credentials (%s)", r.openAPIResource.getResourceName(), res.StatusCode, resBody)
		case http.StatusNotFound:
			return &openapierr.NotFoundError{OriginalError: fmt.Errorf("HTTP Reponse Status Code %d - Not Found. Could not find resource instance: %s", res.StatusCode, resBody)}
		default:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d not matching expected one %v (%s)", r.openAPIResource.getResourceName(), res.StatusCode, expectedHTTPStatusCodes, resBody)
		}
	}
	return nil
}

func (r resourceFactory) responseContainsExpectedStatus(expectedStatusCodes []int, responseStatusCode int) bool {
	for _, expectedStatusCode := range expectedStatusCodes {
		if expectedStatusCode == responseStatusCode {
			return true
		}
	}
	return false
}

func (r resourceFactory) checkImmutableFields(updatedResourceLocalData *schema.ResourceData, openAPIClient ClientOpenAPI) error {
	resourceSchema, err := r.openAPIResource.getResourceSchema()
	if err != nil {
		return err
	}
	immutableProperties := resourceSchema.getImmutableProperties()
	if len(immutableProperties) > 0 {
		remoteData, err := r.readRemote(updatedResourceLocalData.Id(), openAPIClient)
		if err != nil {
			return err
		}
		for _, immutablePropertyName := range immutableProperties {
			if localValue, exists := r.getResourceDataOKExists(immutablePropertyName, updatedResourceLocalData); exists {
				if localValue != remoteData[immutablePropertyName] {
					// Rolling back data so tf values are not stored in the state file; otherwise terraform would store the
					// data inside the updated (*schema.ResourceData) in the state file
					r.updateStateWithPayloadData(remoteData, updatedResourceLocalData)
					return fmt.Errorf("property %s is immutable and therefore can not be updated. Update operation was aborted; no updates were performed", immutablePropertyName)
				}
			}
		}
	}
	return nil
}

// updateStateWithPayloadData is in charge of saving the given payload into the state file. The property names are
// converted into compliant terraform names if needed.
func (r resourceFactory) updateStateWithPayloadData(remoteData map[string]interface{}, resourceLocalData *schema.ResourceData) error {
	resourceSchema, err := r.openAPIResource.getResourceSchema()
	if err != nil {
		return err
	}
	for propertyName, propertyValue := range remoteData {
		property, err := resourceSchema.getProperty(propertyName)
		if err != nil {
			return fmt.Errorf("failed to update state with remote data. This usually happends when the API returns properties that are not specified in the resource's schema definition in the OpenAPI document - error = %s", err)
		}
		if property.isPropertyNamedID() {
			continue
		}
		// TODO: validate that the data returned by the API matches the data configured by the user. This is a edge case scenario but can likely happen with inconsistent APIs
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
// are always converted to terraform compatible names
func (r resourceFactory) createPayloadFromLocalStateData(resourceLocalData *schema.ResourceData) map[string]interface{} {
	input := map[string]interface{}{}
	resourceSchema, _ := r.openAPIResource.getResourceSchema()
	for _, property := range resourceSchema.Properties {
		propertyName := property.Name
		// ReadOnly properties are not considered for the payload data
		if property.isPropertyNamedID() || property.ReadOnly {
			continue
		}
		if dataValue, ok := r.getResourceDataOKExists(propertyName, resourceLocalData); ok {
			switch reflect.TypeOf(dataValue).Kind() {
			case reflect.Map:
				input[propertyName] = dataValue.(map[string]interface{})
			case reflect.Slice, reflect.Array:
				input[propertyName] = dataValue.([]interface{})
			case reflect.String:
				input[propertyName] = dataValue.(string)
			case reflect.Int:
				input[propertyName] = dataValue.(int)
			case reflect.Float64:
				input[propertyName] = dataValue.(float64)
			case reflect.Bool:
				input[propertyName] = dataValue.(bool)
			default:
				log.Printf("[WARN] [resource='%s'] createPayloadFromLocalStateData unkown dataValue type [propertyName: %s; propertyValueType: %+v;]", r.openAPIResource.getResourceName(), propertyName, reflect.TypeOf(dataValue).Kind())
			}
		}
		log.Printf("[DEBUG] [resource='%s'] createPayloadFromLocalStateData [propertyName: %s; propertyValue: %+v]", r.openAPIResource.getResourceName(), propertyName, input[propertyName])
	}
	return input
}

func (r resourceFactory) getStatusValueFromPayload(payload map[string]interface{}) (string, error) {
	resourceSchema, err := r.openAPIResource.getResourceSchema()
	if err != nil {
		return "", err
	}
	statuses, err := resourceSchema.getStatusIdentifier()
	if err != nil {
		return "", err
	}
	var property = payload
	for _, statusField := range statuses {
		propertyValue, statusExistsInPayload := property[statusField]
		if !statusExistsInPayload {
			return "", fmt.Errorf("payload does not match resouce schema, could not find the status field: %s", statuses)
		}
		switch reflect.TypeOf(propertyValue).Kind() {
		case reflect.Map:
			property = propertyValue.(map[string]interface{})
		case reflect.String:
			return propertyValue.(string), nil
		default:
			return "", fmt.Errorf("status property value '%s' does not have a supported type [string/map]", statuses)
		}
	}
	return "", fmt.Errorf("could not find status value [%s] in the payload provided", statuses)
}

// getResourceDataOK returns the data for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) getResourceDataOKExists(schemaDefinitionPropertyName string, resourceLocalData *schema.ResourceData) (interface{}, bool) {
	resourceSchema, _ := r.openAPIResource.getResourceSchema()
	schemaDefinitionProperty, err := resourceSchema.getProperty(schemaDefinitionPropertyName)
	if err != nil {
		return nil, false
	}
	return resourceLocalData.GetOkExists(schemaDefinitionProperty.getTerraformCompliantPropertyName())
}

// setResourceDataProperty sets the expectedValue for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) setResourceDataProperty(schemaDefinitionPropertyName string, value interface{}, resourceLocalData *schema.ResourceData) error {
	resourceSchema, _ := r.openAPIResource.getResourceSchema()
	schemaDefinitionProperty, err := resourceSchema.getProperty(schemaDefinitionPropertyName)
	if err != nil {
		return fmt.Errorf("could not find schema definition property name %s in the resource data: %s", schemaDefinitionPropertyName, err)
	}
	return resourceLocalData.Set(schemaDefinitionProperty.getTerraformCompliantPropertyName(), value)
}
