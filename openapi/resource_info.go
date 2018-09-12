package openapi

import (
	"fmt"

	"log"
	"strings"

	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	"regexp"
	"time"
)

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
const extTfResourceURL = "x-terraform-resource-host"
const extTfResourceRegionsFmt = "x-terraform-resource-regions-%s"

const idDefaultPropertyName = "id"
const statusDefaultPropertyName = "status"

type resourcesInfo map[string]resourceInfo

// resourceInfo serves as translator between swagger definitions and terraform schemas
type resourceInfo struct {
	name     string
	basePath string
	// path contains relative path to the resource e,g: /v1/resource
	path             string
	host             string
	httpSchemes      []string
	schemaDefinition spec.Schema
	// createPathInfo contains info about /resource, including the POST operation
	createPathInfo spec.PathItem
	// pathInfo contains info about /resource/{id}, including GET, PUT and REMOVE operations if applicable
	pathInfo spec.PathItem
}

func (r resourceInfo) createTerraformResourceSchema() (map[string]*schema.Schema, error) {
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

func (r resourceInfo) convertToTerraformCompliantFieldName(propertyName string, property spec.Schema) string {
	if preferredPropertyName, exists := property.Extensions.GetString(extTfFieldName); exists {
		return terraformutils.ConvertToTerraformCompliantName(preferredPropertyName)
	}
	return terraformutils.ConvertToTerraformCompliantName(propertyName)
}

func (r resourceInfo) createTerraformPropertySchema(propertyName string, property spec.Schema) (*schema.Schema, error) {
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

func (r resourceInfo) validateFunc(propertyName string, property spec.Schema) schema.SchemaValidateFunc {
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

func (r resourceInfo) isRequired(propertyName string, requiredProps []string) bool {
	var required = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (r resourceInfo) createTerraformPropertyBasicSchema(propertyName string, property spec.Schema) (*schema.Schema, error) {
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

func (r resourceInfo) isArrayProperty(property spec.Schema) bool {
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

// getStatusIdentifier returns the property name that is supposed to be used as the status field. The status field
// is selected as follows:
// 1.If the given schema definition contains a property configured with metadata 'x-terraform-field-status' set to true, that property
// will be used to check the different statues for the asynchronous pooling mechanism.
// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// will have a property named 'status'
// 3. If none of the above requirements is met, an error will be returned
func (r resourceInfo) getStatusIdentifier() (string, error) {
	statusProperty := ""
	for propertyName, property := range r.schemaDefinition.Properties {
		if r.isIDProperty(propertyName) {
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
		return "", fmt.Errorf("could not find any status property in the resource swagger definition. Please make sure the resource definition has either one property named 'status' or one property that contains %s metadata", extTfFieldStatus)
	}
	if !r.schemaDefinition.Properties[statusProperty].ReadOnly {
		return "", fmt.Errorf("schema definition status property '%s' must be readOnly", statusProperty)
	}
	return statusProperty, nil
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
	log.Printf("[DEBUG] failed to compile region identifier regex: %s", err)
	if err != nil {
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
	log.Printf("missing '%s' root level region extension %s", regionIdentifier, regionExtensionValue)
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

func (r resourceInfo) isIDProperty(propertyName string) bool {
	return r.propertyNameMatchesDefaultName(propertyName, idDefaultPropertyName)
}

func (r resourceInfo) isStatusProperty(propertyName string) bool {
	return r.propertyNameMatchesDefaultName(propertyName, statusDefaultPropertyName)
}

func (r resourceInfo) propertyNameMatchesDefaultName(propertyName, expectedPropertyName string) bool {
	return terraformutils.ConvertToTerraformCompliantName(propertyName) == expectedPropertyName
}
