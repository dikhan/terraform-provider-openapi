package openapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/dikhan/terraform-provider-openapi/v3/openapi/openapierr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func crudWithContext(crudFunc func(data *schema.ResourceData, i interface{}) error, timeoutFor string, resourceName string) func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
		errChan := make(chan error, 1)
		go func() { errChan <- crudFunc(data, i) }()
		select {
		case <-ctx.Done():
			return diag.Errorf("%s: '%s' %s timeout is %s", ctx.Err(), resourceName, timeoutFor, data.Timeout(timeoutFor))
		case err := <-errChan:
			if err != nil {
				return diag.FromErr(err)
			}
		}
		return nil
	}
}

func checkHTTPStatusCode(openAPIResource SpecResource, res *http.Response, expectedHTTPStatusCodes []int) error {
	if !responseContainsExpectedStatus(expectedHTTPStatusCodes, res.StatusCode) {
		var resBody string
		if res.Body != nil {
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d - Error '%s' occurred while reading the response body", openAPIResource.GetResourceName(), res.StatusCode, err)
			}
			if b != nil && len(b) > 0 {
				resBody = string(b)
			}
		}
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d - Unauthorized: API access is denied due to invalid credentials (%s)", openAPIResource.GetResourceName(), res.StatusCode, resBody)
		case http.StatusNotFound:
			return &openapierr.NotFoundError{OriginalError: fmt.Errorf("HTTP Response Status Code %d - Not Found. Could not find resource instance: %s", res.StatusCode, resBody)}
		default:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d not matching expected one %v (%s)", openAPIResource.GetResourceName(), res.StatusCode, expectedHTTPStatusCodes, resBody)
		}
	}
	return nil
}

func responseContainsExpectedStatus(expectedStatusCodes []int, responseStatusCode int) bool {
	for _, expectedStatusCode := range expectedStatusCodes {
		if expectedStatusCode == responseStatusCode {
			return true
		}
	}
	return false
}

func getParentIDsAndResourcePath(openAPIResource SpecResource, data *schema.ResourceData) (parentIDs []string, resourcePath string, err error) {
	parentIDs, err = getParentIDs(openAPIResource, data)
	if err != nil {
		return nil, "", err
	}
	resourcePath, err = openAPIResource.getResourcePath(parentIDs)
	if err != nil {
		return nil, "", err
	}
	return
}

func getParentIDs(openAPIResource SpecResource, data *schema.ResourceData) ([]string, error) {
	if openAPIResource == nil {
		return []string{}, errors.New("can't get parent ids from an empty SpecResource")
	}
	if data == nil {
		return []string{}, errors.New("can't get parent ids from a nil ResourceData")
	}
	parentResourceInfo := openAPIResource.GetParentResourceInfo()
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

// updateStateWithPayloadData is in charge of saving the given payload into the state file keeping for list properties the
// same order as the input (if the list property has the IgnoreItemsOrder set to true). The property names are converted into compliant terraform names if needed.
// The property names are converted into compliant terraform names if needed.
func updateStateWithPayloadData(openAPIResource SpecResource, remoteData map[string]interface{}, resourceLocalData *schema.ResourceData) error {
	return updateStateWithPayloadDataAndOptions(openAPIResource, remoteData, resourceLocalData, true)
}

// dataSourceUpdateStateWithPayloadData is in charge of saving the given payload into the state file keeping for list properties the
// same order received by the API. The property names are converted into compliant terraform names if needed.
func dataSourceUpdateStateWithPayloadData(openAPIResource SpecResource, remoteData map[string]interface{}, resourceLocalData *schema.ResourceData) error {
	return updateStateWithPayloadDataAndOptions(openAPIResource, remoteData, resourceLocalData, false)
}

// updateStateWithPayloadDataAndOptions is in charge of saving the given payload into the state file AND if the ignoreListOrder is enabled
// it will go ahead and compare the items in the list (input vs remote) for properties of type list and the flag 'IgnoreItemsOrder' set to true
// The property names are converted into compliant terraform names if needed.
func updateStateWithPayloadDataAndOptions(openAPIResource SpecResource, remoteData map[string]interface{}, resourceLocalData *schema.ResourceData, ignoreListOrderEnabled bool) error {
	resourceSchema, err := openAPIResource.GetResourceSchema()
	if err != nil {
		return err
	}
	var terraformConfigObject map[string]interface{}
	if resourceLocalData != nil {
		terraformConfigObject = getTerraformConfigObject(resourceLocalData.GetRawConfig()).(map[string]interface{})
	} else {
		return nil
	}
	for propertyName, propertyRemoteValue := range remoteData {
		property, err := resourceSchema.getProperty(propertyName)
		if err != nil {
			log.Printf("[WARN] The API returned a property that is not specified in the resource's schema definition in the OpenAPI document - error = %s", err)
			continue
		}
		if property.isPropertyNamedID() {
			continue
		}

		propValue := propertyRemoteValue
		var propertyLocalStateValue interface{}
		if len(terraformConfigObject) > 0 && !property.isReadOnly() {
			propertyLocalStateValue = terraformConfigObject[property.GetTerraformCompliantPropertyName()]
		} else {
			propertyLocalStateValue = resourceLocalData.Get(property.GetTerraformCompliantPropertyName())
		}

		if ignoreListOrderEnabled && property.shouldIgnoreOrder() {
			propValue = processIgnoreOrderIfEnabled(*property, propertyLocalStateValue, propertyRemoteValue)
		}

		value, err := convertPayloadToLocalStateDataValue(property, propValue, propertyLocalStateValue)
		if err != nil {
			return err
		}
		if value != nil {
			if err := setResourceDataProperty(*property, value, resourceLocalData); err != nil {
				return err
			}
		}
	}
	return nil
}

// processIgnoreOrderIfEnabled checks whether the property has enabled the `IgnoreItemsOrder` field and if so, goes ahead
// and returns a new list trying to match as much as possible the input order from the user (not remotes). The following use
// cases are supported:
// Use case 0: The desired state for an array property (input from user, inputPropertyValue) contains items in certain order AND the remote state (remoteValue) comes back with the same items in the same order.
// Use case 1: The desired state for an array property (input from user, inputPropertyValue) contains items in certain order BUT the remote state (remoteValue) comes back with the same items in different order.
// Use case 2: The desired state for an array property (input from user, inputPropertyValue) contains items in certain order BUT the remote state (remoteValue) comes back with the same items in different order PLUS new ones.
// Use case 3: The desired state for an array property (input from user, inputPropertyValue) contains items in certain order BUT the remote state (remoteValue) comes back with a shorter list where the remaining elems match the inputs.
// Use case 4: The desired state for an array property (input from user, inputPropertyValue) contains items in certain order BUT the remote state (remoteValue) some back with the list with the same size but some elems were updated
func processIgnoreOrderIfEnabled(property SpecSchemaDefinitionProperty, inputPropertyValue, remoteValue interface{}) interface{} {
	if inputPropertyValue == nil || remoteValue == nil { // treat remote as the final state if input value does not exists
		return remoteValue
	}
	if property.shouldIgnoreOrder() {
		newPropertyValue := []interface{}{}
		inputValueArray := inputPropertyValue.([]interface{})
		remoteValueArray := remoteValue.([]interface{})
		for _, inputItemValue := range inputValueArray {
			for _, remoteItemValue := range remoteValueArray {
				if property.equalItems(property.ArrayItemsType, inputItemValue, remoteItemValue) {
					// rearrange elements in remoteValue to follow order in inputValue, which is from the tf config
					// remoteValue is needed as it contains Optional (e.g., ReadOnly) attributes that tf config does not have
					// while retaining the order of elements in tf config to ensure consistency
					var sortedRemoteItemValue = property.syncOrderWhenEqual(property.ArrayItemsType, inputItemValue, remoteItemValue)
					newPropertyValue = append(newPropertyValue, sortedRemoteItemValue)
					break
				}
			}
		}
		modifiedItems := []interface{}{}
		for _, remoteItemValue := range remoteValueArray {
			match := false
			for _, inputItemValue := range inputValueArray {
				if property.equalItems(property.ArrayItemsType, inputItemValue, remoteItemValue) {
					match = true
					break
				}
			}
			if !match {
				modifiedItems = append(modifiedItems, remoteItemValue)
			}
		}
		for _, updatedItem := range modifiedItems {
			newPropertyValue = append(newPropertyValue, updatedItem)
		}
		return newPropertyValue
	}
	return remoteValue
}

func convertPayloadToLocalStateDataValue(property *SpecSchemaDefinitionProperty, propertyValue interface{}, propertyLocalStateValue interface{}) (interface{}, error) {
	if property.WriteOnly {
		return propertyLocalStateValue, nil
	}

	switch property.Type {
	case TypeObject:
		return convertObjectToLocalStateData(property, propertyValue, propertyLocalStateValue)
	case TypeList:
		if isListOfPrimitives, _ := property.isTerraformListOfSimpleValues(); isListOfPrimitives {
			return propertyValue, nil
		}
		if property.isArrayOfObjectsProperty() {
			arrayInput := []interface{}{}

			arrayValue := make([]interface{}, 0)
			if propertyValue != nil {
				arrayValue = propertyValue.([]interface{})
			}

			localStateArrayValue := make([]interface{}, 0)
			if propertyLocalStateValue != nil {
				localStateArrayValue = propertyLocalStateValue.([]interface{})
			}

			for arrayIdx := 0; arrayIdx < len(arrayValue); arrayIdx++ {
				var arrayItem interface{} = nil
				if arrayIdx < len(arrayValue) {
					arrayItem = arrayValue[arrayIdx]
				}
				var localStateArrayItem interface{} = nil
				if arrayIdx < len(localStateArrayValue) {
					localStateArrayItem = localStateArrayValue[arrayIdx]
				}
				objectValue, err := convertObjectToLocalStateData(property, arrayItem, localStateArrayItem)
				if err != nil {
					return err, nil
				}
				if objectValue != nil {
					arrayInput = append(arrayInput, objectValue)
				}
			}
			return arrayInput, nil
		}
		return nil, fmt.Errorf("property '%s' is supposed to be an array objects", property.Name)
	case TypeString:
		if propertyValue == nil {
			return propertyLocalStateValue, nil
		}
		return propertyValue.(string), nil
	case TypeInt:
		if propertyValue == nil {
			return propertyLocalStateValue, nil
		}
		// In golang, a number in JSON message is always parsed into float64, however testing/internal use can define the property value as a proper int.
		if reflect.TypeOf(propertyValue).Kind() == reflect.Int {
			return propertyValue.(int), nil
		}
		return int(propertyValue.(float64)), nil
	case TypeFloat:
		if propertyValue == nil {
			return propertyLocalStateValue, nil
		}
		return propertyValue.(float64), nil
	case TypeBool:
		if propertyValue == nil {
			return propertyLocalStateValue, nil
		}
		return propertyValue.(bool), nil
	default:
		return nil, fmt.Errorf("'%s' type not supported", property.Type)
	}
}

func convertObjectToLocalStateData(property *SpecSchemaDefinitionProperty, propertyValue interface{}, propertyLocalStateValue interface{}) (interface{}, error) {
	objectInput := map[string]interface{}{}

	mapValue := make(map[string]interface{})
	if propertyValue != nil {
		var castOk bool
		mapValue, castOk = propertyValue.(map[string]interface{})
		if !castOk {
			return nil, fmt.Errorf("invalid value '%s' for property '%s' of type '%s'", propertyValue, property.Name, property.Type)
		}
	}

	localStateMapValue := make(map[string]interface{})
	if propertyLocalStateValue != nil {
		if reflect.TypeOf(propertyLocalStateValue).Kind() == reflect.Map {
			localStateMapValue = propertyLocalStateValue.(map[string]interface{})
		} else if reflect.TypeOf(propertyLocalStateValue).Kind() == reflect.Slice && len(propertyLocalStateValue.([]interface{})) == 1 {
			localStateMapValue = propertyLocalStateValue.([]interface{})[0].(map[string]interface{}) // local state can store nested objects as a single item array
		}
	}

	for _, schemaDefinitionProperty := range property.SpecSchemaDefinition.Properties {
		propertyValue := schemaDefinitionProperty.getPropertyValueFromMap(mapValue)

		// Here we are processing the items of the list which are objects. In this case we need to keep the original
		// types as Terraform honors property types for resource schemas attached to TypeList properties
		propValue, err := convertPayloadToLocalStateDataValue(schemaDefinitionProperty, propertyValue, localStateMapValue[schemaDefinitionProperty.GetTerraformCompliantPropertyName()])
		if err != nil {
			return nil, err
		}
		if propValue != nil {
			objectInput[schemaDefinitionProperty.GetTerraformCompliantPropertyName()] = propValue
		}
	}

	// This is the work around put in place to have support for complex objects considering terraform sdk limitation to use
	// blocks only for TypeList and TypeSet . In this case, we need to make sure that the json (which reflects to a map)
	// gets translated to the expected array of one item that terraform expects.
	if property.shouldUseLegacyTerraformSDKBlockApproachForComplexObjects() {
		arrayInput := []interface{}{}
		arrayInput = append(arrayInput, objectInput)
		return arrayInput, nil
	}
	return objectInput, nil
}

// setResourceDataProperty sets the expectedValue for the given schemaDefinitionPropertyName using the terraform compliant property name
func setResourceDataProperty(schemaDefinitionProperty SpecSchemaDefinitionProperty, value interface{}, resourceLocalData *schema.ResourceData) error {
	return resourceLocalData.Set(schemaDefinitionProperty.GetTerraformCompliantPropertyName(), value)
}

// setStateID sets the local resource's data ID with the newly identifier created in the POST API request. Refer to
// r.resourceInfo.getResourceIdentifier() for more info regarding what property is selected as the identifier.
func setStateID(openAPIres SpecResource, resourceLocalData *schema.ResourceData, payload map[string]interface{}) error {
	resourceSchema, err := openAPIres.GetResourceSchema()
	if err != nil {
		return err
	}
	identifierProperty, err := resourceSchema.getResourceIdentifier()
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

// Gets the HCL code equivalent map object of the resource without caring about the state file.
// Only usable during certain operation lifecycles like updates, since read operations do not appear to
// expose the raw configuration during runtime. Useful to deal with scenarios where the local terraform state
// behaves unexpected, like the problem of computed properties within lists not "moving" as expected
// when the ordering changes
func getTerraformConfigObject(rawConfig cty.Value) interface{} {
	objectType := rawConfig.Type()
	if objectType.IsMapType() || objectType.IsObjectType() {
		output := map[string]interface{}{}
		if rawConfig.IsNull() {
			return output
		}
		mapValue := rawConfig.AsValueMap()
		for key, value := range mapValue {
			output[key] = getTerraformConfigObject(value)
		}
		return output
	}

	if objectType.IsListType() {
		output := []interface{}{}
		if rawConfig.IsNull() {
			return output
		}
		for _, listItemValue := range rawConfig.AsValueSlice() {
			output = append(output, getTerraformConfigObject(listItemValue))
		}
		return output
	}

	if objectType.Equals(cty.String) {
		if rawConfig.IsNull() {
			return ""
		}
		return rawConfig.AsString()
	}

	if objectType.Equals(cty.Number) {
		if rawConfig.IsNull() {
			return 0
		}
		number := rawConfig.AsBigFloat()
		if number.IsInt() {
			intVal, _ := number.Int64()
			return int(intVal)
		}
		floatVal, _ := number.Float64()
		return floatVal
	}

	if objectType.Equals(cty.Bool) {
		if rawConfig.IsNull() {
			return false
		}
		return rawConfig.True()
	}

	return nil // unknown type, default to nil
}
