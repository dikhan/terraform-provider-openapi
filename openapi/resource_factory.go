package openapi

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

type resourceFactory struct {
	openAPIResource SpecResource
}

func (r resourceFactory) createTerraformResource() (*schema.Resource, error) {
	s, err := r.createResourceSchema()
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

func (r resourceFactory) createResourceSchema() (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}
	schemaDefinition, err := r.openAPIResource.getResourceSchema()
	if err != nil {
		return nil, err
	}
	for _, property := range schemaDefinition.Properties {
		// Terraform already has a field ID reserved, hence the schema does not need to include an explicit ID property
		if property.isPropertyNamedID() {
			continue
		}
		tfSchema, err := r.createTerraformPropertySchema(property)
		if err != nil {
			return nil, err
		}
		s[property.getTerraformCompliantPropertyName()] = tfSchema
	}
	return s, nil
}

func (r resourceFactory) createTerraformPropertySchema(property SchemaDefinitionProperty) (*schema.Schema, error) {
	propertySchema, err := r.createTerraformPropertyBasicSchema(property)
	if err != nil {
		return nil, err
	}
	// ValidateFunc is not yet supported on lists or sets
	if !property.isArrayProperty() {
		propertySchema.ValidateFunc = r.validateFunc(property)
	}
	return propertySchema, nil
}

func (r resourceFactory) createTerraformPropertyBasicSchema(property SchemaDefinitionProperty) (*schema.Schema, error) {
	var propertySchema *schema.Schema

	switch property.Type {
	case typeList:
		// Arrays only support 'string' items at the moment
		propertySchema = &schema.Schema{
			Type: schema.TypeList,
			Elem: &schema.Schema{Type: schema.TypeString},
		}
	case typeString:
		propertySchema = &schema.Schema{
			Type: schema.TypeString,
		}
	case typeInt:
		propertySchema = &schema.Schema{
			Type: schema.TypeInt,
		}
	case typeFloat:
		propertySchema = &schema.Schema{
			Type: schema.TypeFloat,
		}
	case typeBool:
		propertySchema = &schema.Schema{
			Type: schema.TypeBool,
		}
	}

	// Set the property as required or optional
	if property.isRequired() {
		propertySchema.Required = true
	} else {
		propertySchema.Optional = true
	}

	// If the value of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new value will be created
	propertySchema.ForceNew = property.ForceNew

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	propertySchema.Computed = property.ReadOnly

	// A sensitive property means that the value will not be disclosed in the state file, preventing secrets from
	// being leaked
	propertySchema.Sensitive = property.Sensitive

	if property.Default != nil {
		if property.ReadOnly {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error
			log.Printf("[WARN] '%s.%s' is readOnly and can not have a default value. The value is expected to be computed by the API. Terraform will fail on runtime when performing the property validation check", r.openAPIResource.getResourceName(), property.Name)
		} else {
			propertySchema.Default = property.Default
		}
	}
	return propertySchema, nil
}

func (r resourceFactory) validateFunc(property SchemaDefinitionProperty) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if property.Default != nil {
			if property.ReadOnly {
				err := fmt.Errorf(
					"'%s.%s' is configured as 'readOnly' and can not have a default value. The value is expected to be computed by the API. To fix the issue, pick one of the following options:\n"+
						"1. Remove the 'readOnly' attribute from %s in the swagger file so the default value '%v' can be applied. Default must be nil if computed\n"+
						"OR\n"+
						"2. Remove the 'default' attribute from %s in the swagger file, this means that the API will compute the value as specified by the 'readOnly' attribute\n", r.openAPIResource.getResourceName(), k, k, property.Default, k)
				errors = append(errors, err)
			}
		}
		return
	}
}

func (r resourceFactory) create(data *schema.ResourceData, i interface{}) error {
	openAPIClient := i.(ProviderClient)
	requestPayload := r.createPayloadFromLocalStateData(data)
	responsePayload := map[string]interface{}{}
	res, err := openAPIClient.Post(r.openAPIResource, requestPayload, responsePayload)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusCreated, http.StatusAccepted}); err != nil {
		return fmt.Errorf("[resource='%s'] POST %s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), err)
	}
	return r.updateLocalState(data, responsePayload)
}

func (r resourceFactory) read(data *schema.ResourceData, i interface{}) error {
	openAPIClient := i.(ProviderClient)
	remoteData, err := r.readRemote(data.Id(), openAPIClient)
	if err != nil {
		return err
	}
	return r.updateStateWithPayloadData(remoteData, data)
}

func (r resourceFactory) readRemote(id string, openAPIClient ProviderClient) (map[string]interface{}, error) {
	var err error
	responsePayload := map[string]interface{}{}
	resp, err := openAPIClient.Get(r.openAPIResource, id, responsePayload)
	if err != nil {
		return nil, err
	}
	if err := r.checkHTTPStatusCode(resp, []int{http.StatusOK}); err != nil {
		return nil, fmt.Errorf("[resource='%s'] GET %s/%s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), id, err)
	}
	return responsePayload, nil
}

func (r resourceFactory) update(data *schema.ResourceData, i interface{}) error {
	openAPIClient := i.(ProviderClient)
	operation := r.openAPIResource.getResourcePutOperation()
	if operation == nil {
		return fmt.Errorf("[resource='%s'] resource does not support PUT opperation, check the swagger file exposed on '%s'", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath())
	}
	requestPayload := r.createPayloadFromLocalStateData(data)
	responsePayload := map[string]interface{}{}
	if err := r.checkImmutableFields(data, openAPIClient); err != nil {
		return err
	}
	res, err := openAPIClient.Put(r.openAPIResource, data.Id(), requestPayload, responsePayload)
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusOK}); err != nil {
		return fmt.Errorf("[resource='%s'] UPDATE %s/%s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), data.Id(), err)
	}
	return r.updateStateWithPayloadData(responsePayload, data)
}

func (r resourceFactory) delete(data *schema.ResourceData, i interface{}) error {
	openAPIClient := i.(ProviderClient)
	operation := r.openAPIResource.getResourceDeleteOperation()
	if operation == nil {
		return fmt.Errorf("[resource='%s'] resource does not support DELETE opperation, check the swagger file exposed on '%s'", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath())
	}
	res, err := openAPIClient.Delete(r.openAPIResource, data.Id())
	if err != nil {
		return err
	}
	if err := r.checkHTTPStatusCode(res, []int{http.StatusNoContent, http.StatusOK}); err != nil {
		return fmt.Errorf("[resource='%s'] DELETE %s/%s failed: %s", r.openAPIResource.getResourceName(), r.openAPIResource.getResourcePath(), data.Id(), err)
	}
	return nil
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
			return fmt.Errorf("[resource='%s'] HTTP Reponse Status Code %d - Error '%s' occurred while reading the response body", r.openAPIResource.getResourceName(), res.StatusCode, err)
		}
		if len(b) > 0 {
			resBody = string(b)
		}
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("[resource='%s'] HTTP Reponse Status Code %d - Unauthorized: API access is denied due to invalid credentials (%s)", r.openAPIResource.getResourceName(), res.StatusCode, resBody)
		default:
			return fmt.Errorf("[resource='%s'] HTTP Reponse Status Code %d not matching expected one %v (%s)", r.openAPIResource.getResourceName(), res.StatusCode, expectedHTTPStatusCodes, resBody)
		}
	}
	return nil
}

func (r resourceFactory) checkImmutableFields(updatedResourceLocalData *schema.ResourceData, openAPIClient ProviderClient) error {
	var remoteData map[string]interface{}
	var err error
	if remoteData, err = r.readRemote(updatedResourceLocalData.Id(), openAPIClient); err != nil {
		return err
	}
	resourceSchema, err := r.openAPIResource.getResourceSchema()
	if err != nil {
		return err
	}
	for _, immutablePropertyName := range resourceSchema.getImmutableProperties() {
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
		// TODO: validate that the data returned by the API matches the data configured by the user. This is a edge case sceneario but can likely happen with inconsistent APIs
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
	resourceSchema, _ := r.openAPIResource.getResourceSchema()
	for propertyName, property := range resourceSchema.Properties {
		// ReadOnly properties are not considered for the payload data
		if property.isPropertyNamedID() || property.ReadOnly {
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
		log.Printf("[DEBUG] [resource='%s'] createPayloadFromLocalStateData [%s] - newValue[%+v]", r.openAPIResource.getResourceName(), propertyName, input[propertyName])
	}
	return input
}

// getResourceDataOK returns the data for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) getResourceDataOKExists(schemaDefinitionPropertyName string, resourceLocalData *schema.ResourceData) (interface{}, bool) {
	resourceSchema, _ := r.openAPIResource.getResourceSchema()
	schemaDefinitionProperty := resourceSchema.Properties[schemaDefinitionPropertyName]
	return resourceLocalData.GetOkExists(schemaDefinitionProperty.getTerraformCompliantPropertyName())
}

// setResourceDataProperty sets the value for the given schemaDefinitionPropertyName using the terraform compliant property name
func (r resourceFactory) setResourceDataProperty(schemaDefinitionPropertyName string, value interface{}, resourceLocalData *schema.ResourceData) error {
	resourceSchema, _ := r.openAPIResource.getResourceSchema()
	schemaDefinitionProperty := resourceSchema.Properties[schemaDefinitionPropertyName]
	return resourceLocalData.Set(schemaDefinitionProperty.getTerraformCompliantPropertyName(), value)
}
