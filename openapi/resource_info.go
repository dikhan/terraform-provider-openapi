package openapi

import (
	"fmt"
	"reflect"

	"log"
	"strings"

	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	"regexp"
	"time"
)

const resourceVersionRegex = "(/v[0-9]*/)"
const resourceNameRegex = "((/\\w*[/]?))+$"

// Definition level extensions
const extTfImmutable = "x-terraform-immutable"
const extTfForceNew = "x-terraform-force-new"
const extTfSensitive = "x-terraform-sensitive"
const extTfFieldName = "x-terraform-field-name"
const extTfFieldStatus = "x-terraform-field-status"
const extTfID = "x-terraform-id"

// Operation level extensions
const extTfExcludeResource = "x-terraform-exclude-resource"
const extTfResourcePollEnabled = "x-terraform-resource-poll-enabled"
const extTfResourcePollTargetStatuses = "x-terraform-resource-poll-completed-statuses"
const extTfResourcePollPendingStatuses = "x-terraform-resource-poll-pending-statuses"
const extTfResourceTimeout = "x-terraform-resource-timeout"
const extTfResourceName = "x-terraform-resource-name"
const extTfResourceURL = "x-terraform-resource-host"
const extTfResourceRegionsFmt = "x-terraform-resource-regions-%s"

const idDefaultPropertyName = "id"
const statusDefaultPropertyName = "status"

type resourcesInfo map[string]resourceInfo

// resourceInfo serves as translator between swagger definitions and terraform schemas
type resourceInfo struct {
	basePath string
	// path contains relative path to the resource e,g: /v1/resource
	path             string
	host             string
	httpSchemes      []string
	schemaDefinition *spec.Schema
	// createPathInfo contains info about /resource, including the POST operation
	createPathInfo spec.PathItem
	// pathInfo contains info about /resource/{id}, including GET, PUT and REMOVE operations if applicable
	pathInfo spec.PathItem

	// schemaDefinitions contains all the definitions which might be needed in case the resource schema contains properties
	// of type object which in turn refer to other definitions
	schemaDefinitions map[string]spec.Schema
}

func (r resourceInfo) createTerraformResourceSchema() (map[string]*schema.Schema, error) {
	return r.terraformSchema(r.schemaDefinition, true)
}

// terraformSchema returns the terraform schema for the given schema definition. if ignoreID is true then properties named
// id will be ignored, this is because terraform already has an ID field reserved that identifies uniquely the resource and
// root level schema can not contain a property named ID. For other levels, in case there are properties of type object
// id named properties is allowed as there won't be a  conflict with terraform in that case.
func (r resourceInfo) terraformSchema(schemaDefinition *spec.Schema, ignoreID bool) (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}
	for propertyName, property := range schemaDefinition.Properties {
		// ID should only be ignored when looping through the root level properties of the schema definition
		if r.isIDProperty(propertyName) && ignoreID {
			continue
		}
		tfSchema, err := r.createTerraformPropertyBasicSchema(propertyName, property, schemaDefinition.Required)
		if err != nil {
			return nil, err
		}
		s[r.convertToTerraformCompliantFieldName(propertyName, property)] = tfSchema
	}
	return s, nil
}

func (r resourceInfo) convertToTerraformCompliantFieldName(propertyName string, property spec.Schema) string {
	if preferredPropertyName, exists := property.Extensions.GetString(extTfFieldName); exists {
		return terraformutils.ConvertToTerraformCompliantName(preferredPropertyName)
	}
	return terraformutils.ConvertToTerraformCompliantName(propertyName)
}

func (r resourceInfo) validateFunc(propertyName string, property spec.Schema) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if property.Default != nil {
			if property.ReadOnly {
				err := fmt.Errorf(
					"'%s' is configured as 'readOnly' and can not have a default value. The value is expected to be computed by the API. To fix the issue, pick one of the following options:\n"+
						"1. Remove the 'readOnly' attribute from %s in the swagger file so the default value '%v' can be applied. Default must be nil if computed\n"+
						"OR\n"+
						"2. Remove the 'default' attribute from %s in the swagger file, this means that the API will compute the value as specified by the 'readOnly' attribute\n", k, k, property.Default, k)
				errors = append(errors, err)
			}
		}
		return
	}
}

func (r resourceInfo) isRequired(propertyName string, requiredProps []string) bool {
	var required = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (r resourceInfo) createTerraformPropertyBasicSchema(propertyName string, property spec.Schema, requiredProperties []string) (*schema.Schema, error) {
	var propertySchema *schema.Schema
	if isObject, schemaDefinition, err := r.isObjectProperty(property); isObject {
		if err != nil {
			return nil, err
		}
		s, err := r.terraformSchema(schemaDefinition, false)
		if err != nil {
			return nil, err
		}
		propertySchema = &schema.Schema{
			Type: schema.TypeMap,
			Elem: &schema.Resource{
				Schema: s,
			},
		}
	} else if r.isArrayProperty(property) { // Arrays only support 'string' items at the moment
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
	required := r.isRequired(propertyName, requiredProperties)
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
			log.Printf("[WARN] '%s' is readOnly and can not have a default value. The value is expected to be computed by the API. Terraform will fail on runtime when performing the property validation check", propertyName)
		} else {
			propertySchema.Default = property.Default
		}
	}

	// ValidateFunc is not yet supported on lists or sets
	if !r.isArrayProperty(property) {
		propertySchema.ValidateFunc = r.validateFunc(propertyName, property)
	}

	return propertySchema, nil
}

func (r resourceInfo) isObjectProperty(property spec.Schema) (bool, *spec.Schema, error) {

	if r.isObjectTypeProperty(property) && len(property.Properties) != 0 {
		return true, &property, nil
	}

	if property.Ref.Ref.GetURL() != nil {
		schema, err := openapiutils.GetSchemaDefinition(r.schemaDefinitions, property.Ref.String())
		if err != nil {
			return true, nil, err
		}
		return true, schema, nil
	}
	return false, nil, nil
}

func (r resourceInfo) isArrayProperty(property spec.Schema) bool {
	return r.isOfType(property, "array")
}

func (r resourceInfo) isObjectTypeProperty(property spec.Schema) bool {
	return r.isOfType(property, "object")
}

func (r resourceInfo) isOfType(property spec.Schema, propertyType string) bool {
	return property.Type.Contains(propertyType)
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

func (r resourceInfo) getResourceURL() (string, error) {
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

func (r resourceInfo) getResourceIDURL(id string) (string, error) {
	url, err := r.getResourceURL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", url, id), nil
}

// getResourceIdentifier returns the property name that is supposed to be used as the identifier. The resource id
// is selected as follows:
// 1.If the given schema definition contains a property configured with metadata 'x-terraform-id' set to true, that property
// will be used to set the state ID of the resource. Additionally, the value will be used when performing GET/PUT/DELETE requests to
// identify the resource in question.
// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// will have a property named 'id'
// 3. If none of the above requirements is met, an error will be returned
func (r resourceInfo) getResourceIdentifier() (string, error) {
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

// getStatusIdentifier loops through the schema definition given and tries to find the status id. This method supports both simple structures
// where the status field is at the schema definition root level or complex structures where status field is meant to be
// a sub-property of an object type property
// 1.If the given schema definition contains a property configured with metadata 'x-terraform-field-status' set to true, that property
// will be used to check the different statues for the asynchronous pooling mechanism.
// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// will have a property named 'status'
// 3. If the status field is NOT an object, then the array returned will contain one element with the property name that identifies
// the status field.
// 3. If the schema definition contains a deemed status field (as described above) and the property is of object type, the same logic
// as above will be applied to identify the status field to be used within the object property. In this case the result will
// be an array containing the property hierarchy, starting from the root and ending with the actual status field. This is needed
// so the correct status field can be extracted from payloads.
// 4. If none of the above requirements is met, an error will be returned
func (r resourceInfo) getStatusIdentifier(schemaDefinition *spec.Schema, shouldIgnoreID, shouldEnforceReadOnly bool) ([]string, error) {
	var statusProperty string
	var statusHierarchy []string
	for propertyName, property := range schemaDefinition.Properties {
		if r.isIDProperty(propertyName) && shouldIgnoreID {
			continue
		}
		if r.isStatusProperty(propertyName) {
			statusProperty = propertyName
			continue
		}
		// field with extTfFieldStatus metadata takes preference over 'status' fields as the service provider is the one acknowledging
		// the fact that this field should be used as identifier of the resource
		if terraformID, ok := property.Extensions.GetBool(extTfFieldStatus); ok && terraformID {
			statusProperty = propertyName
			break
		}
	}
	// if the id field is missing and there isn't any properties set with extTfFieldStatus, there is not way for the resource
	// to be identified and therefore an error is returned
	if statusProperty == "" {
		return nil, fmt.Errorf("could not find any status property in the resource swagger definition. Please make sure the resource definition has either one property named 'status' or one property that contains %s metadata", extTfFieldStatus)
	}
	if !schemaDefinition.Properties[statusProperty].ReadOnly && shouldEnforceReadOnly {
		return nil, fmt.Errorf("schema definition status property '%s' must be readOnly", statusProperty)
	}

	statusHierarchy = append(statusHierarchy, statusProperty)
	if isObject, propertySchemaDefinition, err := r.isObjectProperty(schemaDefinition.Properties[statusProperty]); isObject {
		if err != nil {
			return nil, err
		}
		statusIdentifier, err := r.getStatusIdentifier(propertySchemaDefinition, false, false)
		if err != nil {
			return nil, err
		}
		statusHierarchy = append(statusHierarchy, statusIdentifier...)
	}

	return statusHierarchy, nil
}

func (r resourceInfo) getStatusValueFromPayload(payload map[string]interface{}) (string, error) {
	statuses, err := r.getStatusIdentifier(r.schemaDefinition, true, true)
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

// shouldIgnoreResource checks whether the POST operation for a given resource as the 'x-terraform-exclude-resource' extension
// defined with true value. If so, the resource will not be exposed to the OpenAPI Terraform provder; otherwise it will
// be exposed and users will be able to manage such resource via terraform.
func (r resourceInfo) shouldIgnoreResource() bool {
	if extensionExists, ignoreResource := r.createPathInfo.Post.Extensions.GetBool(extTfExcludeResource); extensionExists && ignoreResource {
		return true
	}
	return false
}

// getResourceOverrideHost checks if the x-terraform-resource-host extension is present and if so returns its value. This
// value will override the global host value, and the API calls for this resource will be made agasint the value returned
func (r resourceInfo) getResourceOverrideHost() string {
	if resourceURL, exists := r.createPathInfo.Post.Extensions.GetString(extTfResourceURL); exists && resourceURL != "" {
		return resourceURL
	}
	return ""
}

func (r resourceInfo) isMultiRegionHost(overrideHost string) (bool, *regexp.Regexp) {
	regex, err := regexp.Compile("(\\S+)(\\$\\{(\\S+)\\})(\\S+)")
	if err != nil {
		log.Printf("[DEBUG] failed to compile region identifier regex: %s", err)
		return false, nil
	}
	return len(regex.FindStringSubmatch(overrideHost)) != 0, regex
}

// isMultiRegionResource returns true on ly if:
// - the value is parametrized following the pattern: some.subdomain.${keyword}.domain.com, where ${keyword} must be present in the string, otherwise the resource will not be considered multi region
// - there is a matching 'x-terraform-resource-regions-${keyword}' extension defined in the swagger root level (extensions passed in), where ${keyword} will be the value of the parameter in the above URL
// - and finally the value of the extension is an array of strings containing the different regions where the resource can be created
func (r resourceInfo) isMultiRegionResource(extensions spec.Extensions) (bool, map[string]string) {
	overrideHost := r.getResourceOverrideHost()
	if overrideHost == "" {
		return false, nil
	}
	isMultiRegionHost, regex := r.isMultiRegionHost(overrideHost)
	if !isMultiRegionHost {
		return false, nil
	}
	region := regex.FindStringSubmatch(overrideHost)
	if len(region) != 5 {
		log.Printf("[DEBUG] override host %s provided does not comply with expected regex format", overrideHost)
		return false, nil
	}
	regionIdentifier := region[3]
	regionExtensionValue := fmt.Sprintf(extTfResourceRegionsFmt, regionIdentifier)
	if resourceRegions, exists := openapiutils.StringExtensionExists(extensions, regionExtensionValue); exists {
		resourceRegions = strings.Replace(resourceRegions, " ", "", -1)
		regions := strings.Split(resourceRegions, ",")
		if len(regions) < 1 {
			log.Printf("[DEBUG] could not find any region for '%s' matching region extension %s: '%s'", regionIdentifier, regionExtensionValue, resourceRegions)
			return false, nil
		}
		apiRegionsMap := map[string]string{}
		for _, region := range regions {
			repStr := fmt.Sprintf("${1}%s$4", region)
			apiRegionsMap[region] = regex.ReplaceAllString(overrideHost, repStr)
		}
		if len(apiRegionsMap) < 1 {
			log.Printf("[DEBUG] could not build properly the resource region map for '%s' matching region extension %s: '%s'", regionIdentifier, regionExtensionValue, resourceRegions)
			return false, nil

		}
		return true, apiRegionsMap
	}
	log.Printf("missing matching '%s' root level region extension '%s'", regionIdentifier, regionExtensionValue)
	return false, nil
}

// isResourcePollingEnabled checks whether there is any response code defined for the given responseStatusCode and if so
// whether that response contains the extension 'x-terraform-resource-poll-enabled' set to true returning true;
// otherwise false is returned
func (r resourceInfo) isResourcePollingEnabled(responses *spec.Responses, responseStatusCode int) (bool, *spec.Response) {
	response, exists := responses.StatusCodeResponses[responseStatusCode]
	if !exists {
		return false, nil
	}
	if isResourcePollEnabled, ok := response.Extensions.GetBool(extTfResourcePollEnabled); ok && isResourcePollEnabled {
		return true, &response
	}
	return false, nil
}

func (r resourceInfo) getResourcePollTargetStatuses(response spec.Response) ([]string, error) {
	return r.getPollingStatuses(response, extTfResourcePollTargetStatuses)
}

func (r resourceInfo) getResourcePollPendingStatuses(response spec.Response) ([]string, error) {
	return r.getPollingStatuses(response, extTfResourcePollPendingStatuses)
}

func (r resourceInfo) getPollingStatuses(response spec.Response, extension string) ([]string, error) {
	statuses := []string{}
	if resourcePollTargets, exists := response.Extensions.GetString(extension); exists {
		spaceTrimmedTargets := strings.Replace(resourcePollTargets, " ", "", -1)
		statuses = strings.Split(spaceTrimmedTargets, ",")
	} else {
		return nil, fmt.Errorf("response missing required extension '%s' for the polling mechanism to work", extension)
	}
	return statuses, nil
}

func (r resourceInfo) getResourceTimeout(operation *spec.Operation) (*time.Duration, error) {
	if operation == nil {
		return nil, nil
	}
	return r.getTimeDuration(operation.Extensions, extTfResourceTimeout)
}

func (r resourceInfo) getTimeDuration(extensions spec.Extensions, extension string) (*time.Duration, error) {
	if value, exists := extensions.GetString(extension); exists {
		regex, err := regexp.Compile("^\\d+(\\.\\d+)?[smh]{1}$")
		if err != nil {
			return nil, err
		}
		if !regex.Match([]byte(value)) {
			return nil, fmt.Errorf("invalid duration value: '%s'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)", value)
		}
		return r.getDuration(value)
	}
	return nil, nil
}

func (r resourceInfo) getDuration(t string) (*time.Duration, error) {
	duration, err := time.ParseDuration(t)
	return &duration, err
}

func (r resourceInfo) getResourceTerraformName() string {
	return r.getExtensionStringValue(r.createPathInfo.Post.Extensions, extTfResourceName)
}

func (r resourceInfo) getExtensionStringValue(extensions spec.Extensions, key string) string {
	if value, exists := extensions.GetString(key); exists && value != "" {
		return value
	}
	return ""
}

func (r resourceInfo) isIDProperty(propertyName string) bool {
	return r.propertyNameMatchesDefaultName(propertyName, idDefaultPropertyName)
}

func (r resourceInfo) isStatusProperty(propertyName string) bool {
	return r.propertyNameMatchesDefaultName(propertyName, statusDefaultPropertyName)
}

func (r resourceInfo) propertyNameMatchesDefaultName(propertyName, expectedPropertyName string) bool {
	return terraformutils.ConvertToTerraformCompliantName(propertyName) == expectedPropertyName
}

// getResourceName gets the name of the resource from a path /resource/{id}
func (r resourceInfo) getResourceName() (string, error) {
	nameRegex, err := regexp.Compile(resourceNameRegex)
	if err != nil {
		return "", fmt.Errorf("an error occurred while compiling the resourceNameRegex regex '%s': %s", resourceNameRegex, err)
	}
	var resourceName string
	resourcePath := r.path
	matches := nameRegex.FindStringSubmatch(resourcePath)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find a valid name for resource instance path '%s'", resourcePath)
	}
	resourceName = strings.Replace(matches[len(matches)-1], "/", "", -1)

	if preferredName := r.getResourceTerraformName(); preferredName != "" {
		resourceName = preferredName
	}

	versionRegex, err := regexp.Compile(resourceVersionRegex)
	if err != nil {
		return "", fmt.Errorf("an error occurred while compiling the resourceVersionRegex regex '%s': %s", resourceVersionRegex, err)
	}
	versionMatches := versionRegex.FindStringSubmatch(resourcePath)
	if len(versionMatches) != 0 {
		version := strings.Replace(versionRegex.FindStringSubmatch(resourcePath)[1], "/", "", -1)
		resourceNameWithVersion := fmt.Sprintf("%s_%s", resourceName, version)
		return resourceNameWithVersion, nil
	}
	return resourceName, nil
}
