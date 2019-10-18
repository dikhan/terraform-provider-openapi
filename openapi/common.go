package openapi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/dikhan/terraform-provider-openapi/openapi/openapierr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func checkHTTPStatusCode(openAPIResource SpecResource, res *http.Response, expectedHTTPStatusCodes []int) error {
	if !responseContainsExpectedStatus(expectedHTTPStatusCodes, res.StatusCode) {
		var resBody string
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d - Error '%s' occurred while reading the response body", openAPIResource.getResourceName(), res.StatusCode, err)
		}
		if b != nil && len(b) > 0 {
			resBody = string(b)
		}
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d - Unauthorized: API access is denied due to invalid credentials (%s)", openAPIResource.getResourceName(), res.StatusCode, resBody)
		case http.StatusNotFound:
			return &openapierr.NotFoundError{OriginalError: fmt.Errorf("HTTP Response Status Code %d - Not Found. Could not find resource instance: %s", res.StatusCode, resBody)}
		default:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d not matching expected one %v (%s)", openAPIResource.getResourceName(), res.StatusCode, expectedHTTPStatusCodes, resBody)
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
		return []string{}, errors.New("can't get parent ids from a resourceFactory with no openAPIResource")
	}

	parentResourceInfo := openAPIResource.getParentResourceInfo()
	if parentResourceInfo != nil {
		parentResourceNames := parentResourceInfo.getParentPropertiesNames()

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

// updateStateWithPayloadData is in charge of saving the given payload into the state file. The property names are
// converted into compliant terraform names if needed.
func updateStateWithPayloadData(openAPIResource SpecResource, remoteData map[string]interface{}, resourceLocalData *schema.ResourceData) error {
	resourceSchema, err := openAPIResource.getResourceSchema()
	if err != nil {
		return err
	}
	for propertyName, propertyValue := range remoteData {
		property, err := resourceSchema.getProperty(propertyName)
		if err != nil {
			return fmt.Errorf("failed to update state with remote data. This usually happens when the API returns properties that are not specified in the resource's schema definition in the OpenAPI document - error = %s", err)
		}
		if property.isPropertyNamedID() {
			continue
		}
		value, err := convertPayloadToLocalStateDataValue(property, propertyValue, false)
		if err != nil {
			return err
		}
		if value != nil {
			if err := setResourceDataProperty(openAPIResource, propertyName, value, resourceLocalData); err != nil {
				return err
			}
		}
	}
	return nil
}

func convertPayloadToLocalStateDataValue(property *specSchemaDefinitionProperty, propertyValue interface{}, useString bool) (interface{}, error) {
	if propertyValue == nil {
		return nil, nil
	}
	dataValueKind := reflect.TypeOf(propertyValue).Kind()
	switch dataValueKind {
	case reflect.Map:
		objectInput := map[string]interface{}{}
		mapValue := propertyValue.(map[string]interface{})
		for propertyName, propertyValue := range mapValue {
			schemaDefinitionProperty, err := property.SpecSchemaDefinition.getProperty(propertyName)
			if err != nil {
				return nil, err
			}
			var propValue interface{}
			// Here we are processing the items of the list which are objects. In this case we need to keep the original
			// types as Terraform honors property types for resource schemas attached to typeList properties
			if property.isArrayOfObjectsProperty() {
				propValue, err = convertPayloadToLocalStateDataValue(schemaDefinitionProperty, propertyValue, false)
			} else { // Here we need to use strings as values as terraform typeMap only supports string items
				propValue, err = convertPayloadToLocalStateDataValue(schemaDefinitionProperty, propertyValue, true)
			}
			if err != nil {
				return nil, err
			}
			objectInput[schemaDefinitionProperty.getTerraformCompliantPropertyName()] = propValue
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
	case reflect.Slice, reflect.Array:
		if isListOfPrimitives, _ := property.isTerraformListOfSimpleValues(); isListOfPrimitives {
			return propertyValue, nil
		}
		if property.isArrayOfObjectsProperty() {
			arrayInput := []interface{}{}
			arrayValue := propertyValue.([]interface{})
			for _, arrayItem := range arrayValue {
				objectValue, err := convertPayloadToLocalStateDataValue(property, arrayItem, false)
				if err != nil {
					return err, nil
				}
				arrayInput = append(arrayInput, objectValue)
			}
			return arrayInput, nil
		}
		return nil, fmt.Errorf("property '%s' is supposed to be an array objects", property.Name)
	case reflect.String:
		return propertyValue.(string), nil
	case reflect.Int:
		if useString {
			return fmt.Sprintf("%d", propertyValue.(int)), nil
		}
		return propertyValue.(int), nil
	case reflect.Float64:
		// In golang, a number in JSON message is always parsed into float64. Hence, checking here if the property value is
		// an actual int or if not then casting to float64
		if property.Type == typeInt {
			if useString {
				return fmt.Sprintf("%d", int(propertyValue.(float64))), nil
			}
			return int(propertyValue.(float64)), nil
		}
		if useString {
			// For some reason after apply the state for object configurations is saved with values "0.00" but subsequent plans find diffs saying that the value changed from "0.00" to "0"
			// Adding this check for the time being to avoid the above diffs
			if propertyValue.(float64) == 0 {
				return "0", nil
			}
			return fmt.Sprintf("%.2f", propertyValue.(float64)), nil
		}
		return propertyValue.(float64), nil
	case reflect.Bool:
		if useString {
			// this is only applicable to objects
			if propertyValue.(bool) {
				return fmt.Sprintf("true"), nil
			}
			return fmt.Sprintf("false"), nil
		}
		return propertyValue.(bool), nil
	default:
		return nil, fmt.Errorf("'%s' type not supported", dataValueKind)
	}
}

// setResourceDataProperty sets the expectedValue for the given schemaDefinitionPropertyName using the terraform compliant property name
func setResourceDataProperty(openAPIResource SpecResource, schemaDefinitionPropertyName string, value interface{}, resourceLocalData *schema.ResourceData) error {
	resourceSchema, _ := openAPIResource.getResourceSchema()
	schemaDefinitionProperty, err := resourceSchema.getProperty(schemaDefinitionPropertyName)
	if err != nil {
		return fmt.Errorf("could not find schema definition property name %s in the resource data: %s", schemaDefinitionPropertyName, err)
	}
	return resourceLocalData.Set(schemaDefinitionProperty.getTerraformCompliantPropertyName(), value)
}

// setStateID sets the local resource's data ID with the newly identifier created in the POST API request. Refer to
// r.resourceInfo.getResourceIdentifier() for more info regarding what property is selected as the identifier.
func setStateID(openAPIres SpecResource, resourceLocalData *schema.ResourceData, payload map[string]interface{}) error {
	resourceSchema, err := openAPIres.getResourceSchema()
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
