package openapiv2

import (
	"github.com/hashicorp/terraform/helper/schema"
	"reflect"
	"github.com/go-openapi/spec"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"io/ioutil"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"strings"
)

const resourceVersionRegex = "(/v[0-9]*/)"
const resourceNameRegex = "((/\\w*/){\\w*})+$"
const resourceInstanceRegex = "((?:.*)){.*}"
const swaggerResourcePayloadDefinitionRegex = "(\\w+)[^//]*$"

const extTfImmutable = "x-terraform-immutable"
const extTfForceNew = "x-terraform-force-new"
const extTfSensitive = "x-terraform-sensitive"
const extTfExcludeResource = "x-terraform-exclude-resource"
const extTfFieldName = "x-terraform-field-name"
const extTfID = "x-terraform-id"

type OpenApiV2Resource struct {

}

func (o *OpenApiV2Resource) createResourceSchema() (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}
	for propertyName, property := range r.schemaDefinition.Properties {
		if r.isIDProperty(propertyName) {
			continue
		}
		tfSchema, err := r.createTerraformPropertySchema(propertyName, property)
		if err != nil {
			return nil, err
		}
		s[r.convertToTerraformCompliantFieldName(propertyName, property)] = tfSchema
	}
	return s, nil
}

func (o *OpenApiV2Resource) convertToTerraformCompliantFieldName(propertyName string, property spec.Schema) string {
	if preferredPropertyName, exists := property.Extensions.GetString(extTfFieldName); exists {
		return terraformutils.ConvertToTerraformCompliantName(preferredPropertyName)
	}
	return terraformutils.ConvertToTerraformCompliantName(propertyName)
}

func (o *OpenApiV2Resource) createTerraformPropertySchema(propertyName string, property spec.Schema) (*schema.Schema, error) {
	propertySchema, err := r.createTerraformPropertyBasicSchema(propertyName, property)
	if err != nil {
		return nil, err
	}
	// ValidateFunc is not yet supported on lists or sets
	if !r.isArrayProperty(property) {
		propertySchema.ValidateFunc = r.validateFunc(propertyName, property)
	}
	return propertySchema, nil
}

func (o *OpenApiV2Resource) validateFunc(propertyName string, property spec.Schema) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if property.Default != nil {
			if property.ReadOnly {
				err := fmt.Errorf(
					"'%s.%s' is configured as 'readOnly' and can not have a default value. The value is expected to be computed by the API. To fix the issue, pick one of the following options:\n"+
						"1. Remove the 'readOnly' attribute from %s in the swagger file so the default value '%v' can be applied. Default must be nil if computed\n"+
						"OR\n"+
						"2. Remove the 'default' attribute from %s in the swagger file, this means that the API will compute the value as specified by the 'readOnly' attribute\n", r.name, k, k, property.Default, k)
				errors = append(errors, err)
			}
		}
		return
	}
}

func (o *OpenApiV2Resource) isRequired(propertyName string, requiredProps []string) bool {
	var required = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (o *OpenApiV2Resource) createTerraformPropertyBasicSchema(propertyName string, property spec.Schema) (*schema.Schema, error) {
	var propertySchema *schema.Schema
	// Arrays only support 'string' items at the moment
	if r.isArrayProperty(property) {
		propertySchema = &schema.Schema{
			Type: schema.TypeList,
			Elem: &schema.Schema{Type: schema.TypeString},
		}
	} else if property.Type.Contains("string") {
		propertySchema = &schema.Schema{
			Type: schema.TypeString,
		}
	} else if property.Type.Contains("integer") {
		propertySchema = &schema.Schema{
			Type: schema.TypeInt,
		}
	} else if property.Type.Contains("number") {
		propertySchema = &schema.Schema{
			Type: schema.TypeFloat,
		}
	} else if property.Type.Contains("boolean") {
		propertySchema = &schema.Schema{
			Type: schema.TypeBool,
		}
	}

	// Set the property as required or optional
	required := r.isRequired(propertyName, r.schemaDefinition.Required)
	if required {
		propertySchema.Required = true
	} else {
		propertySchema.Optional = true
	}

	// If the value of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new value will be created
	if forceNew, ok := property.Extensions.GetBool(extTfForceNew); ok && forceNew {
		propertySchema.ForceNew = true
	}

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	if property.ReadOnly {
		propertySchema.Computed = true
	}

	// A sensitive property means that the value will not be disclosed in the state file, preventing secrets from
	// being leaked
	if sensitive, ok := property.Extensions.GetBool(extTfSensitive); ok && sensitive {
		propertySchema.Sensitive = true
	}

	if property.Default != nil {
		if property.ReadOnly {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error
			log.Printf("[WARN] '%s.%s' is readOnly and can not have a default value. The value is expected to be computed by the API. Terraform will fail on runtime when performing the property validation check", r.name, propertyName)
		} else {
			propertySchema.Default = property.Default
		}
	}
	return propertySchema, nil
}

func (o *OpenApiV2Resource) isArrayProperty(property spec.Schema) bool {
	return property.Type.Contains("array")
}

func (r resourceInfo) getImmutableProperties() []string {
	var immutableProperties []string
	for propertyName, property := range r.schemaDefinition.Properties {
		if r.isIDProperty(propertyName) {
			continue
		}
		if immutable, ok := property.Extensions.GetBool(extTfImmutable); ok && immutable {
			immutableProperties = append(immutableProperties, propertyName)
		}
	}
	return immutableProperties
}

func v getResourceURL() (string, error) {
	if r.host == "" || r.path == "" {
		return "", fmt.Errorf("host and path are mandatory attributes to get the resource URL - host['%s'], path['%s']", r.host, r.path)
	}
	defaultScheme := "http"
	for _, scheme := range r.httpSchemes {
		if scheme == "https" {
			defaultScheme = "https"
		}
	}
	path := r.path
	if strings.Index(r.path, "/") != 0 {
		path = fmt.Sprintf("/%s", r.path)
	}
	if r.basePath != "" && r.basePath != "/" {
		if strings.Index(r.basePath, "/") == 0 {
			return fmt.Sprintf("%s://%s%s%s", defaultScheme, r.host, r.basePath, path), nil
		}
		return fmt.Sprintf("%s://%s/%s%s", defaultScheme, r.host, r.basePath, path), nil
	}
	return fmt.Sprintf("%s://%s%s", defaultScheme, r.host, path), nil
}

func (o *OpenApiV2Resource) getResourceIDURL(id string) (string, error) {
	url, err := r.getResourceURL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", url, id), nil
}

// getResourceIdentifier returns the property name that is supposed to be used as the identifier. The resource id
// is selected as follows:
// 1.If the given schema definition contains a property configured with metadata 'x-terraform-id' set to true, that property value
// will be used to set the state ID of the resource. Additionally, the value will be used when performing GET/PUT/DELETE requests to
// identify the resource in question.
// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// will have a property named 'id'
// 3. If none of the above requirements is met, an error will be returned
func (o *OpenApiV2Resource) getResourceIdentifier() (string, error) {
	identifierProperty := ""
	for propertyName, property := range r.schemaDefinition.Properties {
		if r.isIDProperty(propertyName) {
			identifierProperty = propertyName
			continue
		}
		// field with extTfID metadata takes preference over 'id' fields as the service provider is the one acknowledging
		// the fact that this field should be used as identifier of the resource
		if terraformID, ok := property.Extensions.GetBool(extTfID); ok && terraformID {
			identifierProperty = propertyName
			break
		}
	}
	// if the id field is missing and there isn't any properties set with extTfID, there is not way for the resource
	// to be identified and therefore an error is returned
	if identifierProperty == "" {
		return "", fmt.Errorf("could not find any identifier property in the resource payload swagger definition. Please make sure the payload definition has either one property named 'id' or one property that contains %s metadata", extTfID)
	}
	return identifierProperty, nil
}

// shouldIgnoreResource checks whether the POST operation for a given resource as the 'x-terraform-exclude-resource' extension
// defined with true value. If so, the resource will not be exposed to the OpenAPI Terraform provder; otherwise it will
// be exposed and users will be able to manage such resource via terraform.
func (o *OpenApiV2Resource) shouldIgnoreResource() bool {
	if extensionExists, ignoreResource := r.createPathInfo.Post.Extensions.GetBool(extTfExcludeResource); extensionExists && ignoreResource {
		return true
	}
	return false
}

func (o *OpenApiV2Resource) isIDProperty(propertyName string) bool {
	return terraformutils.ConvertToTerraformCompliantName(propertyName) == "id"
}

func (o *OpenApiV2Resource) create(resourceLocalData *schema.ResourceData, i interface{}) error {
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

func (o *OpenApiV2Resource) read(resourceLocalData *schema.ResourceData, i interface{}) error {
	providerConfig := i.(providerConfig)
	remoteData, err := r.readRemote(resourceLocalData.Id(), providerConfig)
	if err != nil {
		return err
	}
	return r.updateStateWithPayloadData(remoteData, resourceLocalData)
}

func (o *OpenApiV2Resource) readRemote(id string, providerConfig providerConfig) (map[string]interface{}, error) {
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

func (o *OpenApiV2Resource) update(resourceLocalData *schema.ResourceData, i interface{}) error {
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

func (o *OpenApiV2Resource) delete(resourceLocalData *schema.ResourceData, i interface{}) error {
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
func (o *OpenApiV2Resource) appendOperationHeaders(operation *spec.Operation, providerConfig providerConfig, headers map[string]string) map[string]string {
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
func (o *OpenApiV2Resource) setStateID(resourceLocalData *schema.ResourceData, payload map[string]interface{}) error {
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
func (o *OpenApiV2Resource) updateLocalState(resourceLocalData *schema.ResourceData, payload map[string]interface{}) error {
	err := r.setStateID(resourceLocalData, payload)
	if err != nil {
		return err
	}
	return r.updateStateWithPayloadData(payload, resourceLocalData)
}

func (o *OpenApiV2Resource) checkHTTPStatusCode(res *http.Response, expectedHTTPStatusCodes []int) error {
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

func (o *OpenApiV2Resource) checkImmutableFields(updatedResourceLocalData *schema.ResourceData, providerConfig providerConfig) error {
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
func (o *OpenApiV2Resource) updateStateWithPayloadData(remoteData map[string]interface{}, resourceLocalData *schema.ResourceData) error {
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
func (o *OpenApiV2Resource) createPayloadFromLocalStateData(resourceLocalData *schema.ResourceData) map[string]interface{} {
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
func (o *OpenApiV2Resource) getResourceDataOKExists(schemaDefinitionPropertyName string, resourceLocalData *schema.ResourceData) (interface{}, bool) {
	schemaDefinitionProperty := r.resourceInfo.schemaDefinition.Properties[schemaDefinitionPropertyName]
	dataPropertyName := r.resourceInfo.convertToTerraformCompliantFieldName(schemaDefinitionPropertyName, schemaDefinitionProperty)
	return resourceLocalData.GetOkExists(dataPropertyName)
}

// setResourceDataProperty sets the value for the given schemaDefinitionPropertyName using the terraform compliant property name
func (o *OpenApiV2Resource) setResourceDataProperty(schemaDefinitionPropertyName string, value interface{}, resourceLocalData *schema.ResourceData) error {
	schemaDefinitionProperty := r.resourceInfo.schemaDefinition.Properties[schemaDefinitionPropertyName]
	dataPropertyName := r.resourceInfo.convertToTerraformCompliantFieldName(schemaDefinitionPropertyName, schemaDefinitionProperty)
	return resourceLocalData.Set(dataPropertyName, value)
}
