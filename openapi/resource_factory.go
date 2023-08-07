package openapi

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/dikhan/terraform-provider-openapi/v3/openapi/openapierr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	//log.Printf("[DEBUG] '%s' terraform schema: %+v", r.openAPIResource.GetResourceName(), s)
	//spew.Dump(s)
	timeouts, err := r.createSchemaResourceTimeout()
	if err != nil {
		return nil, err
	}
	resourceName := r.openAPIResource.GetResourceName()
	return &schema.Resource{
		Schema:        s,
		CreateContext: crudWithContext(r.create, schema.TimeoutCreate, resourceName),
		ReadContext:   crudWithContext(r.read, schema.TimeoutRead, resourceName),
		DeleteContext: crudWithContext(r.delete, schema.TimeoutDelete, resourceName),
		UpdateContext: crudWithContext(r.update, schema.TimeoutUpdate, resourceName),
		Importer:      r.importer(),
		Timeouts:      timeouts,
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
	schemaDefinition, err := r.openAPIResource.GetResourceSchema()
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] resource '%s' schemaDefinition: %s", r.openAPIResource.GetResourceName(), sPrettyPrint(schemaDefinition))
	return schemaDefinition.createResourceSchema()
}

func (r resourceFactory) create(data *schema.ResourceData, i interface{}) error {
	providerClient := i.(ClientOpenAPI)

	if r.openAPIResource == nil {
		return fmt.Errorf("missing openAPI resource configuration")
	}
	resourceName := r.openAPIResource.GetResourceName()

	submitTelemetryMetric(providerClient, TelemetryResourceOperationCreate, resourceName, "")

	parentIDs, resourcePath, err := getParentIDsAndResourcePath(r.openAPIResource, data)
	if err != nil {
		return err
	}

	operation := r.openAPIResource.getResourceOperations().Post
	requestPayload := r.createPayloadFromLocalStateData(data)
	responsePayload := map[string]interface{}{}

	res, err := providerClient.Post(r.openAPIResource, requestPayload, &responsePayload, parentIDs...)
	if err != nil {
		return err
	}
	if err := checkHTTPStatusCode(r.openAPIResource, res, []int{http.StatusOK, http.StatusCreated, http.StatusAccepted}); err != nil {
		return fmt.Errorf("[resource='%s'] POST %s failed: %s", r.openAPIResource.GetResourceName(), resourcePath, err)
	}

	err = setStateID(r.openAPIResource, data, responsePayload)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Resource '%s' ID: %s", resourcePath, data.Id())

	err = r.handlePollingIfConfigured(&responsePayload, data, providerClient, operation, res.StatusCode, schema.TimeoutCreate)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after POST %s call with response status code (%d): %s", resourcePath, res.StatusCode, err)
	}

	return updateStateWithPayloadData(r.openAPIResource, responsePayload, data)
}

func (r resourceFactory) readWithOptions(data *schema.ResourceData, i interface{}, handleNotFoundErr bool) error {
	openAPIClient := i.(ClientOpenAPI)

	if r.openAPIResource == nil {
		return fmt.Errorf("missing openAPI resource configuration")
	}
	resourceName := r.openAPIResource.GetResourceName()

	submitTelemetryMetric(openAPIClient, TelemetryResourceOperationRead, resourceName, "")

	parentsIDs, resourcePath, err := getParentIDsAndResourcePath(r.openAPIResource, data)
	if err != nil {
		return err
	}

	remoteData, err := r.readRemote(data.Id(), openAPIClient, parentsIDs...)

	if err != nil {
		if openapiErr, ok := err.(openapierr.Error); ok {
			if openapierr.NotFound == openapiErr.Code() && !handleNotFoundErr {
				return nil
			}
		}
		return fmt.Errorf("[resource='%s'] GET %s/%s failed: %s", r.openAPIResource.GetResourceName(), resourcePath, data.Id(), err)
	}

	return updateStateWithPayloadData(r.openAPIResource, remoteData, data)
}

func (r resourceFactory) read(data *schema.ResourceData, i interface{}) error {
	return r.readWithOptions(data, i, false)
}

func (r resourceFactory) readRemote(id string, providerClient ClientOpenAPI, parentIDs ...string) (map[string]interface{}, error) {
	var err error
	responsePayload := map[string]interface{}{}
	resp, err := providerClient.Get(r.openAPIResource, id, &responsePayload, parentIDs...)
	if err != nil {
		return nil, err
	}

	if err := checkHTTPStatusCode(r.openAPIResource, resp, []int{http.StatusOK}); err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] GET '%s' response received", r.openAPIResource.GetResourceName())
	return responsePayload, nil
}

func (r resourceFactory) getParentIDs(data *schema.ResourceData) ([]string, error) {
	if r.openAPIResource == nil {
		return []string{}, errors.New("can't get parent ids from a resourceFactory with no openAPIResource")
	}

	parentResourceInfo := r.openAPIResource.GetParentResourceInfo()
	if parentResourceInfo != nil {
		parentResourceNames := parentResourceInfo.GetParentPropertiesNames()

		parentIDs := []string{}
		for _, parentResourceName := range parentResourceNames {
			parentResourceID := data.Get(parentResourceName)
			if parentResourceID == nil {
				return nil, fmt.Errorf("could not find ID value in the state file for subresource parent property '%s'", parentResourceName)
			}
			parentIDs = append(parentIDs, parentResourceID.(string))
		}
		return parentIDs, nil
	}

	return []string{}, nil
}

func (r resourceFactory) update(data *schema.ResourceData, i interface{}) error {
	providerClient := i.(ClientOpenAPI)

	if r.openAPIResource == nil {
		return fmt.Errorf("missing openAPI resource configuration")
	}
	resourceName := r.openAPIResource.GetResourceName()

	submitTelemetryMetric(providerClient, TelemetryResourceOperationUpdate, resourceName, "")

	parentsIDs, resourcePath, err := getParentIDsAndResourcePath(r.openAPIResource, data)
	if err != nil {
		return err
	}

	operation := r.openAPIResource.getResourceOperations().Put
	if operation == nil {
		return fmt.Errorf("[resource='%s'] resource does not support PUT operation, check the swagger file exposed on '%s'", r.openAPIResource.GetResourceName(), resourcePath)
	}

	requestPayload := r.createPayloadFromTerraformConfig(data)

	if err := r.checkImmutableFields(data, providerClient, parentsIDs...); err != nil {
		return err
	}

	if operation.responses.getResponse(http.StatusNoContent) != nil {
		// Don't populate responsePayload if the API's successful update response is 204 No Content
		res, err := providerClient.Put(r.openAPIResource, data.Id(), requestPayload, nil, parentsIDs...)
		if err != nil {
			return err
		}
		// If the target resource does have a current representation and that representation is successfully modified in
		// accordance with the state of the enclosed representation, then the origin server must send either a 200 (OK) or
		// a 204 (No Content) response to indicate successful completion of the request.
		// Ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PUT
		if err := checkHTTPStatusCode(r.openAPIResource, res, []int{http.StatusNoContent}); err != nil {
			return fmt.Errorf("[resource='%s'] UPDATE %s/%s failed: %s", r.openAPIResource.GetResourceName(), resourcePath, data.Id(), err)
		}
		return nil
	}

	var responsePayload map[string]interface{}
	res, err := providerClient.Put(r.openAPIResource, data.Id(), requestPayload, &responsePayload, parentsIDs...)
	if err != nil {
		return err
	}
	if err := checkHTTPStatusCode(r.openAPIResource, res, []int{http.StatusOK, http.StatusAccepted}); err != nil {
		return fmt.Errorf("[resource='%s'] UPDATE %s/%s failed: %s", r.openAPIResource.GetResourceName(), resourcePath, data.Id(), err)
	}

	err = r.handlePollingIfConfigured(&responsePayload, data, providerClient, operation, res.StatusCode, schema.TimeoutUpdate)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after PUT %s call with response status code (%d): %s", resourcePath, res.StatusCode, err)
	}

	return updateStateWithPayloadData(r.openAPIResource, responsePayload, data)
}

func (r resourceFactory) delete(data *schema.ResourceData, i interface{}) error {
	providerClient := i.(ClientOpenAPI)

	if r.openAPIResource == nil {
		return fmt.Errorf("missing openAPI resource configuration")
	}
	resourceName := r.openAPIResource.GetResourceName()

	submitTelemetryMetric(providerClient, TelemetryResourceOperationDelete, resourceName, "")

	parentsIDs, resourcePath, err := getParentIDsAndResourcePath(r.openAPIResource, data)
	if err != nil {
		return err
	}

	operation := r.openAPIResource.getResourceOperations().Delete
	if operation == nil {
		return fmt.Errorf("[resource='%s'] resource does not support DELETE operation, check the swagger file exposed on '%s'", r.openAPIResource.GetResourceName(), resourcePath)
	}
	res, err := providerClient.Delete(r.openAPIResource, data.Id(), parentsIDs...)
	if err != nil {
		return err
	}
	if err := checkHTTPStatusCode(r.openAPIResource, res, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}); err != nil {
		if openapiErr, ok := err.(openapierr.Error); ok {
			if openapierr.NotFound == openapiErr.Code() {
				return nil
			}
		}
		return fmt.Errorf("[resource='%s'] DELETE %s/%s failed: %s", r.openAPIResource.GetResourceName(), resourcePath, data.Id(), err)
	}

	err = r.handlePollingIfConfigured(nil, data, providerClient, operation, res.StatusCode, schema.TimeoutDelete)
	if err != nil {
		return fmt.Errorf("polling mechanism failed after DELETE %s call with response status code (%d): %s", resourcePath, res.StatusCode, err)
	}

	return nil
}

func (r resourceFactory) importer() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		State: func(data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
			providerClient := i.(ClientOpenAPI)

			if r.openAPIResource == nil {
				return nil, fmt.Errorf("missing openAPI resource configuration")
			}
			resourceName := r.openAPIResource.GetResourceName()

			submitTelemetryMetric(providerClient, TelemetryResourceOperationImport, resourceName, "")

			results := make([]*schema.ResourceData, 1, 1)
			results[0] = data
			parentResourceInfo := r.openAPIResource.GetParentResourceInfo()
			if parentResourceInfo != nil {
				parentPropertyNames := parentResourceInfo.GetParentPropertiesNames()

				// The expected format for the ID provided when importing a sub-resource is 1234/567 where 1234 would be the parentID and 567 the instance ID
				ids := strings.Split(data.Id(), "/")
				if len(ids) < 2 {
					return results, fmt.Errorf("can not import a subresource without providing all the parent IDs (%d) and the instance ID", len(parentPropertyNames))
				}
				parentIDsLen := len(ids) - 1
				if len(parentPropertyNames) < parentIDsLen {
					return results, fmt.Errorf("the number of parent IDs provided %d is greater than the expected number of parent IDs %d", parentIDsLen, len(parentPropertyNames))
				}
				if len(parentPropertyNames) > parentIDsLen {
					return results, fmt.Errorf("can not import a subresource without all the parent ids, expected %d and got %d parent IDs", len(parentPropertyNames), parentIDsLen)
				}
				for idx, parentPropertyName := range parentPropertyNames {
					err := data.Set(parentPropertyName, ids[idx])
					if err != nil {
						return nil, err
					}
				}
				data.SetId(ids[len(ids)-1])
			}
			// If the resources is NOT a sub-resource and just a top level resource then the array passed in will just contain
			// 	the data object we get from terraform core without any updates.
			err := r.readWithOptions(data, i, true)
			if err != nil {
				return nil, err
			}
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
	log.Printf("[INFO] Waiting for resource '%s' to reach a completion status (%s)", r.openAPIResource.GetResourceName(), targetStatuses)

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
			return nil, "", fmt.Errorf("error on retrieving resource '%s' (%s) when waiting: %s", r.openAPIResource.GetResourceName(), resourceLocalData.Id(), err)
		}

		newStatus, err := r.getStatusValueFromPayload(remoteData)
		if err != nil {
			return nil, "", fmt.Errorf("error occurred while retrieving status identifier value from payload for resource '%s' (%s): %s", r.openAPIResource.GetResourceName(), resourceLocalData.Id(), err)
		}

		log.Printf("[DEBUG] resource status '%s' (%s): %s", r.openAPIResource.GetResourceName(), resourceLocalData.Id(), newStatus)
		return remoteData, newStatus, nil
	}
}

func (r resourceFactory) checkImmutableFields(updatedResourceLocalData *schema.ResourceData, openAPIClient ClientOpenAPI, parentIDs ...string) error {
	remoteData, err := r.readRemote(updatedResourceLocalData.Id(), openAPIClient, parentIDs...)
	if err != nil {
		return err
	}
	localData := r.createPayloadFromLocalStateData(updatedResourceLocalData)
	s, _ := r.openAPIResource.GetResourceSchema()
	for _, p := range s.Properties {
		err := r.validateImmutableProperty(p, remoteData[p.Name], localData[p.Name], false)
		if err != nil {
			// Rolling back data so tf values are not stored in the state file; otherwise terraform would store the
			// data inside the updated (*schema.ResourceData) in the state file
			updateError := updateStateWithPayloadData(r.openAPIResource, remoteData, updatedResourceLocalData)
			if updateError != nil {
				return updateError
			}
			return fmt.Errorf("validation for immutable properties failed: %s. Update operation was aborted; no updates were performed", err)
		}
	}
	return nil
}

func (r resourceFactory) validateImmutableProperty(property *SpecSchemaDefinitionProperty, remoteData interface{}, localData interface{}, checkObjectPropertiesUpdates bool) error {
	if property.ReadOnly || property.IsParentProperty || property.WriteOnly {
		return nil
	}
	switch property.Type {
	case TypeList:
		if property.Immutable {
			localList := localData.([]interface{})
			remoteList := make([]interface{}, 0)
			if remoteList != nil {
				remoteList = remoteData.([]interface{})
			}
			if len(localList) != len(remoteList) {
				return fmt.Errorf("user attempted to update an immutable list property ('%s') size: [user input list size: %d; actual list size: %d]", property.Name, len(localList), len(remoteList))
			}
			if isListOfPrimitives, _ := property.isTerraformListOfSimpleValues(); isListOfPrimitives {

				for idx, elem := range localList {
					if elem != remoteList[idx] {
						return fmt.Errorf("user attempted to update an immutable list property ('%s') element: [user input: %+v; actual: %+v]", property.Name, localList, remoteList)
					}
				}
			} else {
				for idx, localListObj := range localList {
					remoteListObj := remoteList[idx]
					localObj := localListObj.(map[string]interface{})
					remoteObj := remoteListObj.(map[string]interface{})
					for _, objectProp := range property.SpecSchemaDefinition.Properties {
						err := r.validateImmutableProperty(objectProp, remoteObj[objectProp.Name], localObj[objectProp.GetTerraformCompliantPropertyName()], property.Immutable)
						if err != nil {
							return fmt.Errorf("user attempted to update an immutable list of objects ('%s'): [user input: %s; actual: %s]", property.Name, localData, remoteData)
						}
					}
				}
			}
		}
	case TypeObject:
		localObject := localData.(map[string]interface{})
		remoteObject := make(map[string]interface{})
		if remoteData != nil {
			remoteObject = remoteData.(map[string]interface{})
		}
		for _, objProp := range property.SpecSchemaDefinition.Properties {
			err := r.validateImmutableProperty(objProp, remoteObject[objProp.Name], localObject[objProp.GetTerraformCompliantPropertyName()], property.Immutable)
			if err != nil {
				return fmt.Errorf("user attempted to update an immutable object ('%s') property ('%s'): [user input: %s; actual: %s]", property.Name, objProp.Name, localData, remoteData)
			}
		}
	default:
		if property.Immutable || checkObjectPropertiesUpdates { // checkObjectPropertiesUpdates covers the recursive call from objects that are immutable which also make all its properties immutable
			switch remoteData.(type) {
			case float64: // this is due to the json marshalling always mapping ints to float64d
				if property.Type == TypeFloat {
					if localData != remoteData {
						return fmt.Errorf("user attempted to update an immutable float property ('%s'): [user input: %s; actual: %s]", property.Name, localData, remoteData)
					}
				} else {
					if property.Type == TypeInt {
						if localData != int(remoteData.(float64)) {
							return fmt.Errorf("user attempted to update an immutable integer property ('%s'): [user input: %d; actual: %d]", property.Name, localData, int(remoteData.(float64)))
						}
					}
				}
			default:
				if localData != remoteData {
					return fmt.Errorf("user attempted to update an immutable property ('%s'): [user input: %s; actual: %s]", property.Name, localData, remoteData)
				}
			}
		}
	}
	return nil
}

// createPayloadFromLocalStateData is in charge of translating the values saved in the local state into a payload that can be posted/put
// to the API. Note that when reading the properties from the schema definition, there's a conversion to a compliant
// will automatically translate names into terraform compatible names that can be saved in the state file; otherwise
// terraform name so the look up in the local state operation works properly. The property names saved in the local state
// are always converted to terraform compatible names
// Note the readonly properties will not be posted/put to the API. The payload will always contain the desired state as far
// as the input is concerned.
func (r resourceFactory) createPayloadFromLocalStateData(resourceLocalData *schema.ResourceData) map[string]interface{} {
	input := map[string]interface{}{}
	resourceSchema, _ := r.openAPIResource.GetResourceSchema()
	for _, property := range resourceSchema.Properties {
		propertyName := property.Name
		// ReadOnly properties are not considered for the payload data (including the id if it's computed)
		if property.isReadOnly() {
			continue
		}
		if !property.IsParentProperty {
			if dataValue, ok := r.getResourceDataOKExists(*property, resourceLocalData); ok {
				err := r.populatePayload(input, property, dataValue)
				if err != nil {
					log.Printf("[ERROR] [resource='%s'] error when creating the property payload for property '%s': %s", r.openAPIResource.GetResourceName(), propertyName, err)
				}
			}
			log.Printf("[DEBUG] [resource='%s'] property payload [propertyName: %s; propertyValue: %+v]", r.openAPIResource.GetResourceName(), propertyName, input[propertyName])
		}
	}
	log.Printf("[DEBUG] [resource='%s'] createPayloadFromLocalStateData: %s", r.openAPIResource.GetResourceName(), sPrettyPrint(input))
	return input
}

// Similar to createPayloadFromLocalStateData but uses the current terraform configuration to create the request payload
func (r resourceFactory) createPayloadFromTerraformConfig(resourceLocalData *schema.ResourceData) map[string]interface{} {
	terraformConfigObject := getTerraformConfigObject(resourceLocalData.GetRawConfig()).(map[string]interface{})

	input := map[string]interface{}{}
	resourceSchema, _ := r.openAPIResource.GetResourceSchema()
	for _, property := range resourceSchema.Properties {
		propertyName := property.Name
		// ReadOnly properties are not considered for the payload data (including the id if it's computed)
		if property.isReadOnly() {
			continue
		}
		if !property.IsParentProperty {
			if dataValue, ok := terraformConfigObject[property.GetTerraformCompliantPropertyName()]; ok {
				err := r.populatePayload(input, property, dataValue)
				if err != nil {
					log.Printf("[ERROR] [resource='%s'] error when creating the property payload for property '%s': %s", r.openAPIResource.GetResourceName(), propertyName, err)
				}
			}
			log.Printf("[DEBUG] [resource='%s'] property payload [propertyName: %s; propertyValue: %+v]", r.openAPIResource.GetResourceName(), propertyName, input[propertyName])
		}
	}
	log.Printf("[DEBUG] [resource='%s'] createPayloadFromTerraformConfig: %s", r.openAPIResource.GetResourceName(), sPrettyPrint(input))
	return input
}

func (r resourceFactory) populatePayload(input map[string]interface{}, property *SpecSchemaDefinitionProperty, dataValue interface{}) error {
	if property == nil {
		return errors.New("populatePayload must receive a non nil property")
	}
	if dataValue == nil {
		return fmt.Errorf("property '%s' has a nil state dataValue", property.Name)
	}
	if property.isReadOnly() {
		return nil
	}
	dataValueKind := reflect.TypeOf(dataValue).Kind()
	switch dataValueKind {
	case reflect.Map:
		objectInput := map[string]interface{}{}
		mapValue := dataValue.(map[string]interface{})
		for propertyName, propertyValue := range mapValue {
			schemaDefinitionProperty, err := property.SpecSchemaDefinition.getPropertyBasedOnTerraformName(propertyName)
			if err != nil {
				return err
			}
			if err := r.populatePayload(objectInput, schemaDefinitionProperty, propertyValue); err != nil {
				return err
			}
		}
		input[property.Name] = objectInput
	case reflect.Slice, reflect.Array:
		if isListOfPrimitives, _ := property.isTerraformListOfSimpleValues(); isListOfPrimitives {
			input[property.Name] = dataValue.([]interface{})
		} else {
			// This is the work around put in place to have support for complex objects. In this case, because the
			// state representation of nested objects is an array, we need to make sure we don't end up constructing an
			// array but rather just a json object
			if property.shouldUseLegacyTerraformSDKBlockApproachForComplexObjects() {
				arrayValue := dataValue.([]interface{})
				if len(arrayValue) != 1 {
					return fmt.Errorf("something is really wrong here...an object property with nested objects should have exactly one elem in the terraform state list")
				}
				if err := r.populatePayload(input, property, arrayValue[0]); err != nil {
					return err
				}
			} else {
				arrayInput := []interface{}{}
				arrayValue := dataValue.([]interface{})
				for _, arrayItem := range arrayValue {
					objectInput := map[string]interface{}{}
					if err := r.populatePayload(objectInput, property, arrayItem); err != nil {
						return err
					}
					// Only assign the value of the object, otherwise a dup key will be assigned which will cause problems. Example
					// [propertyName: listeners; propertyValue: [map[options:[] origin_ingress_port:80 protocol:http shield_ingress_port:80]]]
					// Here we just want to assign as value: map[options:[] origin_ingress_port:80 protocol:http shield_ingress_port:80]
					arrayInput = append(arrayInput, objectInput[property.Name])
				}
				input[property.Name] = arrayInput
			}
		}
	case reflect.String:
		input[property.Name] = dataValue.(string)
	case reflect.Int:
		input[property.Name] = dataValue.(int)
	case reflect.Float64:
		input[property.Name] = dataValue.(float64)
	case reflect.Bool:
		input[property.Name] = dataValue.(bool)
	default:
		return fmt.Errorf("'%s' type not supported", property.Type)
	}
	return nil
}

func (r resourceFactory) getStatusValueFromPayload(payload map[string]interface{}) (string, error) {
	resourceSchema, err := r.openAPIResource.GetResourceSchema()
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
func (r resourceFactory) getResourceDataOKExists(schemaDefinitionProperty SpecSchemaDefinitionProperty, resourceLocalData *schema.ResourceData) (interface{}, bool) {
	return resourceLocalData.GetOkExists(schemaDefinitionProperty.GetTerraformCompliantPropertyName())
}
