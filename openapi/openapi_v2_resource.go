package openapi

import (
	"fmt"
	"github.com/go-openapi/spec"
	"log"
	"regexp"
	"strings"
	"time"
)

const resourceVersionRegex = "(/v[0-9]*/)"
const resourceNameRegex = "((/\\w*[/]?))+$"
const resourceInstanceRegex = "((?:.*)){.*}"
const swaggerResourcePayloadDefinitionRegex = "(\\w+)[^//]*$"

// Definition level extensions
const extTfImmutable = "x-terraform-immutable"
const extTfForceNew = "x-terraform-force-new"
const extTfSensitive = "x-terraform-sensitive"
const extTfFieldName = "x-terraform-field-name"
const extTfFieldStatus = "x-terraform-field-status"
const extTfID = "x-terraform-id"

// Operation level extensions
const extTfResourceTimeout = "x-terraform-resource-timeout"
const extTfResourcePollEnabled = "x-terraform-resource-poll-enabled"
const extTfResourcePollTargetStatuses = "x-terraform-resource-poll-completed-statuses"
const extTfResourcePollPendingStatuses = "x-terraform-resource-poll-pending-statuses"
const extTfExcludeResource = "x-terraform-exclude-resource"
const extTfResourceName = "x-terraform-resource-name"

// SpecV2Resource defines a struct that implements the SpecResource interface and it's based on OpenAPI v2 specification
type SpecV2Resource struct {
	Name string
	// Path contains the full relative path to the resource e,g: /v1/resource
	Path string
	// specSchemaDefinition definition represents the representational state (aka model) of the resource
	SchemaDefinition spec.Schema
	// RootPathItem contains info about the resource root path e,g: /resource, including the POST operation used to create instances of this resource
	RootPathItem spec.PathItem
	// InstancePathItem contains info about the resource's instance /resource/{id}, including GET, PUT and REMOVE operations if applicable
	InstancePathItem spec.PathItem
}

func newSpecV2Resource(path string, schemaDefinition spec.Schema, rootPathItem, instancePathItem spec.PathItem) (*SpecV2Resource, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	resource := &SpecV2Resource{
		Path:             path,
		SchemaDefinition: schemaDefinition,
		RootPathItem:     rootPathItem,
		InstancePathItem: instancePathItem,
	}
	name, err := resource.buildResourceName()
	if err != nil {
		return nil, fmt.Errorf("could not build resource name for '%s': %s", path, err)
	}
	resource.Name = name
	return resource, nil
}

func (o *SpecV2Resource) getResourceName() string {
	return o.Name
}

// getResourceName gets the name of the resource from a path /resource/{id}
// getResourceName returns the name of the resource (including the version if applicable). The name is build from the resource
// root path /resource/{id} or if specified the value set in the x-terraform-resource-name extension is used instead along
// along with the version (if applicable)
// the provider name to build the terraform resource name that will be used in the terraform configuration file
func (o *SpecV2Resource) buildResourceName() (string, error) {
	nameRegex, err := regexp.Compile(resourceNameRegex)
	if err != nil {
		return "", fmt.Errorf("an error occurred while compiling the resourceNameRegex regex '%s': %s", resourceNameRegex, err)
	}
	var resourceName string
	resourcePath := o.Path
	matches := nameRegex.FindStringSubmatch(resourcePath)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find a valid name for resource instance path '%s'", resourcePath)
	}
	resourceName = strings.Replace(matches[len(matches)-1], "/", "", -1)

	if preferredName := o.getResourceTerraformName(); preferredName != "" {
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

func (o *SpecV2Resource) getResourcePath() string {
	return o.Path
}

func (o *SpecV2Resource) getResourceOperations() specResourceOperations {
	return specResourceOperations{
		Post:   o.createResourceOperation(o.RootPathItem.Post),
		Get:    o.createResourceOperation(o.InstancePathItem.Get),
		Put:    o.createResourceOperation(o.InstancePathItem.Put),
		Delete: o.createResourceOperation(o.InstancePathItem.Delete),
	}
}

// shouldIgnoreResource checks whether the POST operation for a given resource as the 'x-terraform-exclude-resource' extension
// defined with true value. If so, the resource will not be exposed to the OpenAPI Terraform provder; otherwise it will
// be exposed and users will be able to manage such resource via terraform.
func (o *SpecV2Resource) shouldIgnoreResource() bool {
	if extensionExists, ignoreResource := o.RootPathItem.Post.Extensions.GetBool(extTfExcludeResource); extensionExists && ignoreResource {
		return true
	}
	return false
}

func (o *SpecV2Resource) getResourceSchema() (*specSchemaDefinition, error) {
	schemaDefinition := &specSchemaDefinition{}
	schemaDefinition.Properties = specSchemaDefinitionProperties{}
	for propertyName, property := range o.SchemaDefinition.Properties {
		schemaDefinitionProperty, err := o.createSchemaDefinitionProperty(propertyName, property)
		if err != nil {
			return nil, err
		}
		schemaDefinition.Properties[schemaDefinitionProperty.Name] = schemaDefinitionProperty
	}
	return schemaDefinition, nil
}

func (o *SpecV2Resource) createSchemaDefinitionProperty(propertyName string, property spec.Schema) (*specSchemaDefinitionProperty, error) {
	schemaDefinitionProperty := &specSchemaDefinitionProperty{}

	schemaDefinitionProperty.Name = propertyName

	if preferredPropertyName, exists := property.Extensions.GetString(extTfFieldName); exists {
		schemaDefinitionProperty.PreferredName = preferredPropertyName
	}

	if o.isArrayProperty(property) {
		schemaDefinitionProperty.Type = typeList
	} else if property.Type.Contains("string") {
		schemaDefinitionProperty.Type = typeString
	} else if property.Type.Contains("integer") {
		schemaDefinitionProperty.Type = typeInt
	} else if property.Type.Contains("number") {
		schemaDefinitionProperty.Type = typeFloat
	} else if property.Type.Contains("boolean") {
		schemaDefinitionProperty.Type = typeBool
	}

	// Set the property as required or optional
	required := o.isRequired(propertyName, o.SchemaDefinition.Required)
	if required {
		schemaDefinitionProperty.Required = true
	}

	// If the value of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new value will be created
	if forceNew, ok := property.Extensions.GetBool(extTfForceNew); ok && forceNew {
		schemaDefinitionProperty.ForceNew = true
	}

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	if property.ReadOnly {
		schemaDefinitionProperty.ReadOnly = true
	}

	// A sensitive property means that the value will not be disclosed in the state file, preventing secrets from
	// being leaked
	if sensitive, ok := property.Extensions.GetBool(extTfSensitive); ok && sensitive {
		schemaDefinitionProperty.Sensitive = true
	}

	// field with extTfID metadata takes preference over 'id' fields as the service provider is the one acknowledging
	// the fact that this field should be used as identifier of the resource
	if terraformID, ok := property.Extensions.GetBool(extTfID); ok && terraformID {
		schemaDefinitionProperty.IsIdentifier = true
	}

	if immutable, ok := property.Extensions.GetBool(extTfImmutable); ok && immutable {
		schemaDefinitionProperty.Immutable = true
	}

	if isStatusIdentifier, ok := property.Extensions.GetBool(extTfFieldStatus); ok && isStatusIdentifier {
		schemaDefinitionProperty.Immutable = true
	}

	if property.Default != nil {
		if property.ReadOnly {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error
			log.Printf("[WARN] '%s' is readOnly and can not have a default value. The value is expected to be computed by the API. Terraform will fail on runtime when performing the property validation check", propertyName)
		} else {
			schemaDefinitionProperty.Default = property.Default
		}
	}
	return schemaDefinitionProperty, nil
}

func (o *SpecV2Resource) isRequired(propertyName string, requiredProps []string) bool {
	var required = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (o *SpecV2Resource) isArrayProperty(property spec.Schema) bool {
	return property.Type.Contains("array")
}

func (o *SpecV2Resource) getResourceTerraformName() string {
	return o.getExtensionStringValue(o.RootPathItem.Post.Extensions, extTfResourceName)
}

func (o *SpecV2Resource) getExtensionStringValue(extensions spec.Extensions, key string) string {
	if value, exists := extensions.GetString(key); exists && value != "" {
		return value
	}
	return ""
}

func (o *SpecV2Resource) createResourceOperation(operation *spec.Operation) *specResourceOperation {
	if operation == nil {
		return nil
	}
	headerParameters := getHeaderConfigurations(operation.Parameters)
	securitySchemes := createSecuritySchemes(operation.Security)
	return &specResourceOperation{
		HeaderParameters: headerParameters,
		SecuritySchemes:  securitySchemes,
		responses:        o.createResponses(operation),
	}
}

func (o *SpecV2Resource) createResponses(operation *spec.Operation) specResponses {
	responses := specResponses{}
	for statusCode, response := range operation.Responses.StatusCodeResponses {
		responses[statusCode] = &specResponse{
			isPollingEnabled:    o.isResourcePollingEnabled(response),
			pollTargetStatuses:  o.getResourcePollTargetStatuses(response),
			pollPendingStatuses: o.getResourcePollPendingStatuses(response),
		}
	}
	return responses
}

// isResourcePollingEnabled checks whether there is any response code defined for the given responseStatusCode and if so
// whether that response contains the extension 'x-terraform-resource-poll-enabled' set to true returning true;
// otherwise false is returned
func (o *SpecV2Resource) isResourcePollingEnabled(response spec.Response) bool {
	if isResourcePollEnabled, ok := response.Extensions.GetBool(extTfResourcePollEnabled); ok && isResourcePollEnabled {
		return true
	}
	return false
}

func (o *SpecV2Resource) getResourcePollTargetStatuses(response spec.Response) []string {
	return o.getPollingStatuses(response, extTfResourcePollTargetStatuses)
}

func (o *SpecV2Resource) getResourcePollPendingStatuses(response spec.Response) []string {
	return o.getPollingStatuses(response, extTfResourcePollPendingStatuses)
}

func (o *SpecV2Resource) getPollingStatuses(response spec.Response, extension string) []string {
	var statuses []string
	if resourcePollTargets, exists := response.Extensions.GetString(extension); exists {
		spaceTrimmedTargets := strings.Replace(resourcePollTargets, " ", "", -1)
		statuses = strings.Split(spaceTrimmedTargets, ",")
	}
	return statuses
}

func (o *SpecV2Resource) getTimeouts() (*specTimeouts, error) {
	var postTimeout *time.Duration
	var getTimeout *time.Duration
	var putTimeout *time.Duration
	var deleteTimeout *time.Duration
	var err error
	if postTimeout, err = o.getResourceTimeout(o.RootPathItem.Post); err != nil {
		return nil, err
	}
	if getTimeout, err = o.getResourceTimeout(o.InstancePathItem.Get); err != nil {
		return nil, err
	}
	if putTimeout, err = o.getResourceTimeout(o.InstancePathItem.Put); err != nil {
		return nil, err
	}
	if deleteTimeout, err = o.getResourceTimeout(o.InstancePathItem.Delete); err != nil {
		return nil, err
	}
	return &specTimeouts{
		Post:   postTimeout,
		Get:    getTimeout,
		Put:    putTimeout,
		Delete: deleteTimeout,
	}, nil
}

func (o *SpecV2Resource) getResourceTimeout(operation *spec.Operation) (*time.Duration, error) {
	if operation == nil {
		return nil, nil
	}
	return o.getTimeDuration(operation.Extensions, extTfResourceTimeout)
}

func (o *SpecV2Resource) getTimeDuration(extensions spec.Extensions, extension string) (*time.Duration, error) {
	if value, exists := extensions.GetString(extension); exists {
		regex, err := regexp.Compile("^\\d+(\\.\\d+)?[smh]{1}$")
		if err != nil {
			return nil, err
		}
		if !regex.Match([]byte(value)) {
			return nil, fmt.Errorf("invalid duration value: '%s'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)", value)
		}
		return o.getDuration(value)
	}
	return nil, nil
}

func (o *SpecV2Resource) getDuration(t string) (*time.Duration, error) {
	duration, err := time.ParseDuration(t)
	return &duration, err
}
