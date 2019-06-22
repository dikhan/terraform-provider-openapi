package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
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
const extTfComputed = "x-terraform-computed"

// Operation level extensions
const extTfResourceTimeout = "x-terraform-resource-timeout"
const extTfResourcePollEnabled = "x-terraform-resource-poll-enabled"
const extTfResourcePollTargetStatuses = "x-terraform-resource-poll-completed-statuses"
const extTfResourcePollPendingStatuses = "x-terraform-resource-poll-pending-statuses"
const extTfExcludeResource = "x-terraform-exclude-resource"
const extTfResourceName = "x-terraform-resource-name"
const extTfResourceURL = "x-terraform-resource-host"

// SpecV2Resource defines a struct that implements the SpecResource interface and it's based on OpenAPI v2 specification
type SpecV2Resource struct {
	Name   string
	Region string
	// Path contains the full relative path to the resource e,g: /v1/resource
	Path string
	// SpecSchemaDefinition definition represents the representational state (aka model) of the resource
	SchemaDefinition spec.Schema
	// RootPathItem contains info about the resource root path e,g: /resource, including the POST operation used to create instances of this resource
	RootPathItem spec.PathItem
	// InstancePathItem contains info about the resource's instance /resource/{id}, including GET, PUT and REMOVE operations if applicable
	InstancePathItem spec.PathItem

	// SchemaDefinitions contains all the definitions which might be needed in case the resource schema contains properties
	// of type object which in turn refer to other definitions
	SchemaDefinitions map[string]spec.Schema
}

// newSpecV2Resource creates a SpecV2Resource with no region and default host
func newSpecV2Resource(path string, schemaDefinition spec.Schema, rootPathItem, instancePathItem spec.PathItem, schemaDefinitions map[string]spec.Schema) (*SpecV2Resource, error) {
	return newSpecV2ResourceWithRegion("", path, schemaDefinition, rootPathItem, instancePathItem, schemaDefinitions)
}

func newSpecV2ResourceWithRegion(region, path string, schemaDefinition spec.Schema, rootPathItem, instancePathItem spec.PathItem, schemaDefinitions map[string]spec.Schema) (*SpecV2Resource, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	resource := &SpecV2Resource{
		Path:              path,
		Region:            region,
		SchemaDefinition:  schemaDefinition,
		RootPathItem:      rootPathItem,
		InstancePathItem:  instancePathItem,
		SchemaDefinitions: schemaDefinitions,
	}
	name, err := resource.buildResourceName()
	if err != nil {
		return nil, fmt.Errorf("could not build resource name for '%s': %s", path, err)
	}
	resource.Name = name
	return resource, nil
}

func (o *SpecV2Resource) getResourceName() string {
	if o.Region != "" {
		return fmt.Sprintf("%s_%s", o.Name, o.Region)
	}
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

// getHost can return an empty host in which case the expectation is that the host used will be the one specified in the
// swagger host attribute or if not present the host used will be the host where the swagger file was served
func (o *SpecV2Resource) getHost() (string, error) {
	overrideHost := getResourceOverrideHost(o.RootPathItem.Post)
	if overrideHost == "" {
		return "", nil
	}
	multiRegionHost, err := openapiutils.GetMultiRegionHost(overrideHost, o.Region)
	if err != nil {
		return "", nil
	}
	if multiRegionHost != "" {
		return multiRegionHost, nil
	}
	return overrideHost, nil
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
// defined with true value. If so, the resource will not be exposed to the OpenAPI Terraform provider; otherwise it will
// be exposed and users will be able to manage such resource via terraform.
func (o *SpecV2Resource) shouldIgnoreResource() bool {
	if extensionExists, ignoreResource := o.RootPathItem.Post.Extensions.GetBool(extTfExcludeResource); extensionExists && ignoreResource {
		return true
	}
	return false
}

func (o *SpecV2Resource) getResourceSchema() (*specSchemaDefinition, error) {
	return o.getSchemaDefinition(&o.SchemaDefinition)
}

func (o *SpecV2Resource) getSchemaDefinition(schema *spec.Schema) (*specSchemaDefinition, error) {
	schemaDefinition := &specSchemaDefinition{}
	schemaDefinition.Properties = specSchemaDefinitionProperties{}
	for propertyName, property := range schema.Properties {
		schemaDefinitionProperty, err := o.createSchemaDefinitionProperty(propertyName, property, schema.Required)
		if err != nil {
			return nil, err
		}
		schemaDefinition.Properties = append(schemaDefinition.Properties, schemaDefinitionProperty)
	}
	return schemaDefinition, nil
}

func (o *SpecV2Resource) createSchemaDefinitionProperty(propertyName string, property spec.Schema, requiredProperties []string) (*specSchemaDefinitionProperty, error) {
	schemaDefinitionProperty := &specSchemaDefinitionProperty{}

	if isObject, schemaDefinition, err := o.isObjectProperty(property); isObject || err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to process object type property '%s': %s", propertyName, err)
		}
		objectSchemaDefinition, err := o.getSchemaDefinition(schemaDefinition)
		if err != nil {
			return nil, err
		}
		schemaDefinitionProperty.SpecSchemaDefinition = objectSchemaDefinition
		log.Printf("[DEBUG] found object type property '%s'", propertyName)
	} else if isArray, itemsType, itemsSchema, err := o.isArrayProperty(property); isArray || err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to process array type property '%s': %s", propertyName, err)
		}
		schemaDefinitionProperty.ArrayItemsType = itemsType
		schemaDefinitionProperty.SpecSchemaDefinition = itemsSchema // only diff than nil if type is object
		log.Printf("[DEBUG] found array type property '%s' with items of type '%s'", propertyName, itemsType)
	}

	propertyType, err := o.getPropertyType(property)
	if err != nil {
		return nil, err
	}
	schemaDefinitionProperty.Type = propertyType

	schemaDefinitionProperty.Name = propertyName

	if preferredPropertyName, exists := property.Extensions.GetString(extTfFieldName); exists {
		schemaDefinitionProperty.PreferredName = preferredPropertyName
	}

	// Set the property as required (if not required the property will be considered optional)
	required := o.isRequired(propertyName, requiredProperties)
	if required {
		schemaDefinitionProperty.Required = true
		if property.ReadOnly {
			return nil, fmt.Errorf("failed to process property '%s': a required property cannot be readOnly too", propertyName)
		}
		schemaDefinitionProperty.Computed = false
	} else {
		schemaDefinitionProperty.Required = false

		optionalComputed, err := o.isOptionalComputedProperty(propertyName, property, requiredProperties)
		if err != nil {
			return nil, err
		}

		// Only set to true if property is computed OR optional-computed, purely optional properties are not computed since
		// API is not expected to auto-generate any value by default if value is not provided
		schemaDefinitionProperty.Computed = property.ReadOnly || optionalComputed
	}

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back in the response from the api and it is stored in the state.
	// Link: https://swagger.io/docs/specification/data-models/data-types#readonly-writeonly
	// schemaDefinitionProperty.ReadOnly is set to true if the property is explicitly readOnly OR if it's not readOnly but still considered optional computed
	schemaDefinitionProperty.ReadOnly = property.ReadOnly

	// If the value of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new value will be created
	if forceNew, ok := property.Extensions.GetBool(extTfForceNew); ok && forceNew {
		schemaDefinitionProperty.ForceNew = true
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
		schemaDefinitionProperty.IsStatusIdentifier = true
	}

	if o.isPropertyWithNestedObjects(property) {
		schemaDefinitionProperty.IsNestedObject = true
	}

	// Use the default keyword in the parameter schema to specify the default value for an optional parameter. The default
	// value is the one that the server uses if the client does not supply the parameter value in the request.
	// Link: https://swagger.io/docs/specification/describing-parameters#default
	schemaDefinitionProperty.Default = property.Default

	return schemaDefinitionProperty, nil
}

// TODO: fill this out and add unit tests
func (o *SpecV2Resource) isPropertyWithNestedObjects(property spec.Schema) bool {
	// Here we do the following:
	// - check if the type is object, if not return false right away
	// - loop through the property.Properties and if any of the properties is in turn of type object then we return true; false otherwise
	return false
}

func (o *SpecV2Resource) isOptionalComputedProperty(propertyName string, property spec.Schema, requiredProperties []string) (bool, error) {
	required := o.isRequired(propertyName, requiredProperties)
	if required {
		return false, nil
	}

	isOptionalComputedWithDefault, err := o.isOptionalComputedWithDefault(propertyName, property)
	if err != nil {
		return false, err
	}
	if isOptionalComputedWithDefault {
		return true, nil
	}

	isOptionalComputed, err := o.isOptionalComputed(propertyName, property)
	if err != nil {
		return false, err
	}
	if isOptionalComputed {
		return true, nil
	}

	return false, nil
}

// isOptionalComputedWithDefault returns true if the property matches the OpenAPI spec to mark a property as optional
// and computed
// If the property does not have explicitly the 'x-terraform-computed', it could also be a optional computed property
// if it meets the OpenAPI spec for properties that are optional and still can be computed. This can be done
// by specifying the default attribute. Example:
//
// optional_computed_with_default:  # optional property that the default value is known at runtime, hence service provider documents it
//  type: "string"
//  default: “some known default value”
func (o *SpecV2Resource) isOptionalComputedWithDefault(propertyName string, property spec.Schema) (bool, error) {
	if !property.ReadOnly && property.Default != nil {
		if optionalComputed, ok := property.Extensions.GetBool(extTfComputed); ok && optionalComputed {
			return false, fmt.Errorf("optional computed property validation failed for property '%s': optional computed properties with default attributes should not have '%s' extension too", propertyName, extTfComputed)
		}
		return true, nil
	}
	return false, nil
}

// isOptionalComputed returns true if the property is marked with the extension 'x-terraform-computed'
// This covers the use case where a property is not marked as readOnly but still is optional value that can come from the user or if not provided will be computed by the API. Example
//
// optional_computed: # optional property that the default value is NOT known at runtime
//  type: "string"
//  x-terraform-computed: true
func (o *SpecV2Resource) isOptionalComputed(propertyName string, property spec.Schema) (bool, error) {
	if optionalComputed, ok := property.Extensions.GetBool(extTfComputed); ok && optionalComputed {
		if property.ReadOnly {
			return false, fmt.Errorf("optional computed property validation failed for property '%s': optional computed properties marked with '%s' can not be readOnly", propertyName, extTfComputed)
		}
		if property.Default != nil {
			return false, fmt.Errorf("optional computed property validation failed for property '%s': optional computed properties marked with '%s' can not have the default value as the value is not known at plan time. If the value is known, then this extension should not be used, and rather the 'default' attribute should be populated", propertyName, extTfComputed)
		}
		return true, nil
	}
	return false, nil
}

func (o *SpecV2Resource) isArrayItemPrimitiveType(propertyType schemaDefinitionPropertyType) bool {
	return propertyType == typeString || propertyType == typeInt || propertyType == typeFloat || propertyType == typeBool
}

func (o *SpecV2Resource) validateArrayItems(property spec.Schema) (schemaDefinitionPropertyType, error) {
	if property.Items == nil || property.Items.Schema == nil {
		return "", fmt.Errorf("array property is missing items schema definition")
	}
	if o.isArrayTypeProperty(*property.Items.Schema) {
		return "", fmt.Errorf("array property can not have items of type 'array'")
	}
	itemsType, err := o.getPropertyType(*property.Items.Schema)
	if err != nil {
		return "", err
	}
	if !o.isArrayItemPrimitiveType(itemsType) && !(itemsType == typeObject) {
		return "", fmt.Errorf("array item type '%s' not supported", itemsType)
	}
	return itemsType, nil
}

func (o *SpecV2Resource) getPropertyType(property spec.Schema) (schemaDefinitionPropertyType, error) {
	if o.isArrayTypeProperty(property) {
		return typeList, nil
	} else if isObject, _, err := o.isObjectProperty(property); isObject || err != nil {
		return typeObject, err
	} else if property.Type.Contains("string") {
		return typeString, nil
	} else if property.Type.Contains("integer") {
		return typeInt, nil
	} else if property.Type.Contains("number") {
		return typeFloat, nil
	} else if property.Type.Contains("boolean") {
		return typeBool, nil
	}
	return "", fmt.Errorf("non supported '%+v' type", property.Type)
}

func (o *SpecV2Resource) isObjectProperty(property spec.Schema) (bool, *spec.Schema, error) {
	if o.isObjectTypeProperty(property) || property.Ref.Ref.GetURL() != nil {
		// Case of nested object schema
		if len(property.Properties) != 0 {
			return true, &property, nil
		}
		// Case of external ref - in this case the type could be populated or not
		if property.Ref.Ref.GetURL() != nil {
			schema, err := openapiutils.GetSchemaDefinition(o.SchemaDefinitions, property.Ref.String())
			if err != nil {
				return true, nil, fmt.Errorf("object ref is poitning to a non existing schema definition: %s", err)
			}
			return true, schema, nil
		}
		return true, nil, fmt.Errorf("object is missing the nested schema definition or the ref is poitning to a non existing schema definition")
	}
	return false, nil, nil
}

func (o *SpecV2Resource) isArrayProperty(property spec.Schema) (bool, schemaDefinitionPropertyType, *specSchemaDefinition, error) {
	if o.isArrayTypeProperty(property) {
		itemsType, err := o.validateArrayItems(property)
		if err != nil {
			return false, "", nil, err
		}
		if o.isArrayItemPrimitiveType(itemsType) {
			return true, itemsType, nil, nil
		}
		// This is the case where items must be object
		if isObject, schemaDefinition, err := o.isObjectProperty(*property.Items.Schema); isObject || err != nil {
			if err != nil {
				return true, itemsType, nil, err
			}
			objectSchemaDefinition, err := o.getSchemaDefinition(schemaDefinition)
			if err != nil {
				return true, itemsType, nil, err
			}
			return true, itemsType, objectSchemaDefinition, nil
		}
	}
	return false, "", nil, nil
}

func (o *SpecV2Resource) isArrayTypeProperty(property spec.Schema) bool {
	return o.isOfType(property, "array")
}

func (o *SpecV2Resource) isObjectTypeProperty(property spec.Schema) bool {
	return o.isOfType(property, "object")
}

func (o *SpecV2Resource) isOfType(property spec.Schema, propertyType string) bool {
	return property.Type.Contains(propertyType)
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

func (o *SpecV2Resource) getResourceTerraformName() string {
	if o.RootPathItem.Post == nil {
		return ""
	}
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

// getResourceOverrideHost checks if the x-terraform-resource-host extension is present and if so returns its value. This
// value will override the global host value, and the API calls for this resource will be made against the value returned
func getResourceOverrideHost(rootPathItem *spec.Operation) string {
	if resourceURL, exists := rootPathItem.Extensions.GetString(extTfResourceURL); exists && resourceURL != "" {
		return resourceURL
	}
	return ""
}
