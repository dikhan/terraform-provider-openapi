package openapi

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/dikhan/terraform-provider-openapi/v3/openapi/openapiutils"
	"github.com/go-openapi/spec"
)

const pathParameterRegex = "/({[\\w]*})*/"

// resourceVersionRegexTemplate is used to identify the version attached to the given resource. The parameter in the
// template will be replaced with the actual resource name so if there is a match the version grabbed is assured to belong
// to the resource in question and not any other version showing in the path before the resource name
const resourceVersionRegexTemplate = "/(v[\\d]*)/%s"

const resourceNameRegex = "((/[\\w-]*[/]?))+$"

// resourceParentNameRegex is the regex used to identify the different parents from a path that is a sub-resource. If used
// calling FindStringSubmatch, any match will contain the following groups in the corresponding array index:
// Index 0: This value will represent the full match containing also the path parameter (e,g: /v1/cdns/{id})
// Index 1: This value will represent the resource path (without the instance path parameter) - e,g: /v1/cdns
// Index 2: This value will represent version if it exists in the path (e,g: v1)
// Index 3: This value will represent the resource path name (e,g: cdns)
//
// - Example calling FindAllStringSubmatch with '/v1/cdns/{id}/v1/firewalls' path:
// matches, _ := resourceParentRegex.FindAllStringSubmatch("/v1/cdns/{id}/v1/firewalls", -1)
// matches[0][0]: Full match /v1/cdns/{id}
// matches[0][1]: Group 1. /v1/cdns
// matches[0][2]: Group 2. v1
// matches[0][3]: Group 3. cdns

// - Example calling FindAllStringSubmatch with '/v1/cdns/{id}/v2/firewalls/{id}/v3/rules' path
// matches, _ := resourceParentRegex.FindAllStringSubmatch("/v1/cdns/{id}/v2/firewalls/{id}/v3/rules", -1)
// matches[0][0]: Full match /v1/cdns/{id}
// matches[0][1]: Group 1. /v1/cdns
// matches[0][2]: Group 2. v1
// matches[0][3]: Group 3. cdns
// matches[1][0]: Full match /v2/firewalls/{id}
// matches[1][1]: Group 1. /v2/firewalls
// matches[1][2]: Group 2. v2
// matches[1][3]: Group 3. firewalls
const resourceParentNameRegex = `(\/(?:\w+\/)?(?:v\d+\/)?\w+)\/{\w+}`

const resourceInstanceRegex = "((?:.*)){.*}"

// Definition level extensions
const extTfImmutable = "x-terraform-immutable"
const extTfForceNew = "x-terraform-force-new"
const extTfSensitive = "x-terraform-sensitive"
const extTfFieldName = "x-terraform-field-name"
const extTfFieldStatus = "x-terraform-field-status"
const extTfID = "x-terraform-id"
const extTfComputed = "x-terraform-computed"
const extTfIgnoreOrder = "x-terraform-ignore-order"
const extIgnoreOrder = "x-ignore-order"
const extTfWriteOnly = "x-terraform-write-only"

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
	Name string
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

	Paths map[string]spec.PathItem

	// Cached objects that are loaded once (when the corresponding function that loads the object is called the first time) and
	// on subsequent method calls the cached object is returned instead saving executing time.

	// specSchemaDefinitionCached is cached in GetResourceSchema() method
	specSchemaDefinitionCached *SpecSchemaDefinition
	// parentResourceInfoCached is cached in GetParentResourceInfo() method
	parentResourceInfoCached *ParentResourceInfo
	// resolvedPathCached is cached in getResourcePath() method
	resolvedPathCached string
}

// newSpecV2Resource creates a SpecV2Resource with no region and default host
func newSpecV2Resource(path string, schemaDefinition spec.Schema, rootPathItem, instancePathItem spec.PathItem, schemaDefinitions map[string]spec.Schema, paths map[string]spec.PathItem) (*SpecV2Resource, error) {
	return newSpecV2ResourceWithConfig(path, schemaDefinition, rootPathItem, instancePathItem, schemaDefinitions, paths)
}

func newSpecV2DataSource(path string, schemaDefinition spec.Schema, rootPathItem spec.PathItem, paths map[string]spec.PathItem) (*SpecV2Resource, error) {
	resource := &SpecV2Resource{
		Path:              path,
		SchemaDefinition:  schemaDefinition,
		RootPathItem:      rootPathItem,
		InstancePathItem:  spec.PathItem{},
		SchemaDefinitions: nil,
		Paths:             paths,
	}
	name, err := resource.buildResourceName()
	if err != nil {
		return nil, fmt.Errorf("could not build resource name for '%s': %s", path, err)
	}
	resource.Name = name
	return resource, nil
}

func newSpecV2ResourceWithConfig(path string, schemaDefinition spec.Schema, rootPathItem, instancePathItem spec.PathItem, schemaDefinitions map[string]spec.Schema, paths map[string]spec.PathItem) (*SpecV2Resource, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if paths == nil {
		return nil, fmt.Errorf("paths must not be nil")
	}
	resource := &SpecV2Resource{
		Path:              path,
		SchemaDefinition:  schemaDefinition,
		RootPathItem:      rootPathItem,
		InstancePathItem:  instancePathItem,
		SchemaDefinitions: schemaDefinitions,
		Paths:             paths,
	}
	name, err := resource.buildResourceName()
	if err != nil {
		return nil, fmt.Errorf("could not build resource name for '%s': %s", path, err)
	}
	resource.Name = name
	return resource, nil
}

// GetResourceName returns the resource name including the region at the end of the resource name if applicable
func (o *SpecV2Resource) GetResourceName() string {
	return o.Name
}

// GetResourceName returns the name of the resource (including the version if applicable). The name is build from the resource
// root path /resource/{id} or if specified the value set in the x-terraform-resource-name extension is used instead along
// with the version (if applicable)
func (o *SpecV2Resource) buildResourceName() (string, error) {
	preferredName := ""
	if preferred := o.getResourceTerraformName(); preferred != "" {
		preferredName = preferred
	}
	fullResourceName, err := o.buildResourceNameFromPath(o.Path, preferredName)
	if err != nil {
		return "", err
	}
	parentResourceInfo := o.GetParentResourceInfo()
	if parentResourceInfo != nil {
		fullResourceName = parentResourceInfo.fullParentResourceName + "_" + fullResourceName
	}
	return fullResourceName, nil
}

// buildResourceNameFromPath returns the name of the resource (including the version if applicable and using the preferred name
// if provided). The name will be calculated using the last part of the path which is meant to be the resource name that the URI
// refers to (e,g: /resource/{id}). If the path is versioned /v1/resource/{id} then the corresponding returned name will
// be either the built name from the path or the preferred name with the version appended at the end.
// For instance, given the following input the output will be:
// /cdns/{id} -> cdns
// /cdns/{id} and preferred name being cdn -> cdn
// /v1/cdns/{id} -> cdns_v1
// /v1/cdns/{id} and preferred name being cdn -> cdn_v1
func (o *SpecV2Resource) buildResourceNameFromPath(resourcePath, preferredName string) (string, error) {
	nameRegex, _ := regexp.Compile(resourceNameRegex)
	var resourceName string
	matches := nameRegex.FindStringSubmatch(resourcePath)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find a valid name for resource instance path '%s'", resourcePath)
	}
	resourceName = strings.Replace(matches[len(matches)-1], "/", "", -1)
	resourceName = strings.ReplaceAll(resourceName, "-", "_")

	versionRegex, _ := regexp.Compile(fmt.Sprintf(resourceVersionRegexTemplate, resourceName))

	if preferredName != "" {
		resourceName = preferredName
	}

	fullResourceName := resourceName
	v := versionRegex.FindAllStringSubmatch(resourcePath, -1)
	if len(v) > 0 {
		version := v[0][1]
		fullResourceName = fmt.Sprintf("%s_%s", resourceName, version)
	}

	return fullResourceName, nil
}

// getResourcePath returns the root path of the resource. If the resource is a subresource and therefore the path contains
// path parameters these will be resolved accordingly based on the ids provided. For instance, considering the given
// resource path "/v1/cdns/{cdn_id}/v1/firewalls" and the []strin{"cdnID"} the returned path will be "/v1/cdns/cdnID/v1/firewalls".
// If the resource path is not parameterised, then regular path will be returned accordingly
func (o *SpecV2Resource) getResourcePath(parentIDs []string) (string, error) {
	if o.resolvedPathCached != "" {
		log.Printf("[DEBUG] getResourcePath hit the cache for '%s'", o.Name)
		return o.resolvedPathCached, nil
	}
	resolvedPath := o.Path

	pathParameterRegex, _ := regexp.Compile(pathParameterRegex)
	pathParamsMatches := pathParameterRegex.FindAllStringSubmatch(resolvedPath, -1)

	switch {
	case len(pathParamsMatches) == 0:
		o.resolvedPathCached = resolvedPath
		log.Printf("[DEBUG] getResourcePath cache loaded for '%s'", o.Name)
		return resolvedPath, nil

	case len(parentIDs) > len(pathParamsMatches):
		return "", fmt.Errorf("could not resolve sub-resource path correctly '%s' with the given ids - more ids than path params: %s", resolvedPath, parentIDs)

	case len(parentIDs) < len(pathParamsMatches):
		return "", fmt.Errorf("could not resolve sub-resource path correctly '%s' with the given ids - missing ids to resolve the path params properly: %s", resolvedPath, parentIDs)
	}

	// At this point it's assured that there is an equal number of parameters to resolved and their corresponding ID values
	for idx, parentID := range parentIDs {
		if strings.Contains(parentID, "/") {
			return "", fmt.Errorf("could not resolve sub-resource path correctly '%s' due to parent IDs (%s) containing not supported characters (forward slashes)", resolvedPath, parentIDs)
		}
		resolvedPath = strings.Replace(resolvedPath, pathParamsMatches[idx][1], parentIDs[idx], 1)
	}

	o.resolvedPathCached = resolvedPath
	log.Printf("[DEBUG] getResourcePath cache loaded for '%s'", o.Name)
	return resolvedPath, nil
}

// getHost can return an empty host in which case the expectation is that the host used will be the one specified in the
// swagger host attribute or if not present the host used will be the host where the swagger file was served
func (o *SpecV2Resource) getHost() (string, error) {
	overrideHost := getResourceOverrideHost(o.RootPathItem.Post)
	if overrideHost == "" {
		return "", nil
	}
	return overrideHost, nil
}

func (o *SpecV2Resource) getResourceOperations() specResourceOperations {
	return specResourceOperations{
		List:   o.createResourceOperation(o.RootPathItem.Get),
		Post:   o.createResourceOperation(o.RootPathItem.Post),
		Get:    o.createResourceOperation(o.InstancePathItem.Get),
		Put:    o.createResourceOperation(o.InstancePathItem.Put),
		Delete: o.createResourceOperation(o.InstancePathItem.Delete),
	}
}

// ShouldIgnoreResource checks whether the POST operation for a given resource as the 'x-terraform-exclude-resource' extension
// defined with true value. If so, the resource will not be exposed to the OpenAPI Terraform provider; otherwise it will
// be exposed and users will be able to manage such resource via terraform.
func (o *SpecV2Resource) ShouldIgnoreResource() bool {
	postOperation := o.RootPathItem.Post
	if postOperation != nil {
		if postOperation.Extensions != nil {
			if o.isBoolExtensionEnabled(postOperation.Extensions, extTfExcludeResource) {
				return true
			}
		}
	}
	return false
}

// GetParentResourceInfo returns the information about the parent resources
func (o *SpecV2Resource) GetParentResourceInfo() *ParentResourceInfo {
	if o.parentResourceInfoCached != nil {
		log.Printf("[DEBUG] GetParentResourceInfo hit the cache for '%s'", o.Name)
		return o.parentResourceInfoCached
	}
	resourceParentRegex, _ := regexp.Compile(resourceParentNameRegex)
	parentMatches := resourceParentRegex.FindAllStringSubmatch(o.Path, -1)
	if len(parentMatches) > 0 {
		var parentURI string
		var parentInstanceURI string

		var parentResourceNames, parentURIs, parentInstanceURIs []string
		for _, match := range parentMatches {
			fullMatch := match[0]
			rootPath := match[1]
			parentURI = parentInstanceURI + rootPath
			parentInstanceURI = parentInstanceURI + fullMatch
			parentURIs = append(parentURIs, parentURI)
			parentInstanceURIs = append(parentInstanceURIs, parentInstanceURI)
		}

		fullParentResourceName := ""
		preferredParentName := ""
		for _, parentURI := range parentURIs {
			// `o.Paths` is used to read the preferred name over that resource if `x-terraform-preferred-name` is set
			if o.Paths != nil {
				if parent, ok := o.Paths[parentURI]; ok {
					preferredParentName = o.getPreferredName(parent)
				} else {
					// Falling back to checking path with trailing slash
					if parent, ok := o.Paths[parentURI+"/"]; ok {
						preferredParentName = o.getPreferredName(parent)
					}
				}
			}
			parentResourceName, err := o.buildResourceNameFromPath(parentURI, preferredParentName)
			if err != nil {
				log.Printf("[ERROR] could not build parent resource info due to the following error: %s", err)
				return nil //untested
			}
			parentResourceNames = append(parentResourceNames, parentResourceName)
			fullParentResourceName = fullParentResourceName + parentResourceName + "_"
		}
		fullParentResourceName = strings.TrimRight(fullParentResourceName, "_")

		sub := &ParentResourceInfo{
			parentResourceNames:    parentResourceNames,
			fullParentResourceName: fullParentResourceName,
			parentURIs:             parentURIs,
			parentInstanceURIs:     parentInstanceURIs,
		}
		o.parentResourceInfoCached = sub
		log.Printf("[DEBUG] GetParentResourceInfo cache loaded for '%s'", o.Name)
		return sub
	}
	return nil
}

// GetResourceSchema returns the resource schema
func (o *SpecV2Resource) GetResourceSchema() (*SpecSchemaDefinition, error) {
	if o.specSchemaDefinitionCached != nil {
		log.Printf("[DEBUG] GetResourceSchema hit the cache for '%s'", o.Name)
		return o.specSchemaDefinitionCached, nil
	}
	specSchemaDefinition, err := o.getSchemaDefinitionWithOptions(&o.SchemaDefinition, true)
	if err != nil {
		return nil, err
	}
	o.specSchemaDefinitionCached = specSchemaDefinition
	log.Printf("[DEBUG] GetResourceSchema cache loaded for '%s'", o.Name)
	return o.specSchemaDefinitionCached, nil
}

func (o *SpecV2Resource) getSchemaDefinition(schema *spec.Schema) (*SpecSchemaDefinition, error) {
	return o.getSchemaDefinitionWithOptions(schema, false)
}

func (o *SpecV2Resource) getSchemaDefinitionWithOptions(schema *spec.Schema, addParentProps bool) (*SpecSchemaDefinition, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema argument must not be nil")
	}
	schemaDefinition := &SpecSchemaDefinition{}
	schemaDefinition.Properties = SpecSchemaDefinitionProperties{}

	// This map ensures no duplicates will happen if the schema happens to have a parent id property. if so, it will be overridden with the expected parent property configuration (e,g: making the prop required)
	schemaProps := map[string]*SpecSchemaDefinitionProperty{}
	for propertyName, property := range schema.Properties {
		schemaDefinitionProperty, err := o.createSchemaDefinitionProperty(propertyName, property, schema.Required)
		if err != nil {
			return nil, err
		}
		schemaProps[propertyName] = schemaDefinitionProperty
	}
	if addParentProps {
		parentResourceInfo := o.GetParentResourceInfo()
		if parentResourceInfo != nil {
			parentPropertyNames := parentResourceInfo.GetParentPropertiesNames()
			for _, parentPropertyName := range parentPropertyNames {
				pr, _ := o.createSchemaDefinitionProperty(parentPropertyName, spec.Schema{SchemaProps: spec.SchemaProps{Type: spec.StringOrArray{"string"}}}, []string{parentPropertyName})
				pr.IsParentProperty = true
				schemaProps[parentPropertyName] = pr
			}
		}
	}

	for _, property := range schemaProps {
		schemaDefinition.Properties = append(schemaDefinition.Properties, property)
	}
	return schemaDefinition, nil
}

func (o *SpecV2Resource) createSchemaDefinitionProperty(propertyName string, property spec.Schema, requiredProperties []string) (*SpecSchemaDefinitionProperty, error) {
	schemaDefinitionProperty := &SpecSchemaDefinitionProperty{}

	schemaDefinitionProperty.Name = propertyName
	propertyType, err := o.getPropertyType(property)
	if err != nil {
		return nil, fmt.Errorf("failed to process property '%s': %s", propertyName, err)
	}
	schemaDefinitionProperty.Type = propertyType
	schemaDefinitionProperty.Description = property.Description

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

		// Edge case where the description of the property is set under the items property instead of the root level
		//       array_property:
		//        items:
		//          description: Groups allowed to manage this identity
		//          type: string
		//        type: array
		if schemaDefinitionProperty.Description == "" {
			if property.Items != nil && property.Items.Schema != nil {
				schemaDefinitionProperty.Description = property.Items.Schema.Description
			}
		}

		schemaDefinitionProperty.ArrayItemsType = itemsType
		schemaDefinitionProperty.SpecSchemaDefinition = itemsSchema // only diff than nil if type is object

		if o.isBoolExtensionEnabled(property.Extensions, extTfIgnoreOrder) || o.isBoolExtensionEnabled(property.Extensions, extIgnoreOrder) {
			schemaDefinitionProperty.IgnoreItemsOrder = true
		}

		log.Printf("[DEBUG] found array type property '%s' with items of type '%s'", propertyName, itemsType)
	}

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
	if o.isBoolExtensionEnabled(property.Extensions, extTfForceNew) {
		schemaDefinitionProperty.ForceNew = true
	}

	// A sensitive property means that the value will not be disclosed in the state file, preventing secrets from
	// being leaked
	if o.isBoolExtensionEnabled(property.Extensions, extTfSensitive) {
		schemaDefinitionProperty.Sensitive = true
	}

	// field with extTfID metadata takes preference over 'id' fields as the service provider is the one acknowledging
	// the fact that this field should be used as identifier of the resource
	if o.isBoolExtensionEnabled(property.Extensions, extTfID) {
		schemaDefinitionProperty.IsIdentifier = true
	}

	if o.isBoolExtensionEnabled(property.Extensions, extTfImmutable) {
		schemaDefinitionProperty.Immutable = true
	}

	if o.isBoolExtensionEnabled(property.Extensions, extTfFieldStatus) {
		schemaDefinitionProperty.IsStatusIdentifier = true
	}

	if o.isBoolExtensionEnabled(property.Extensions, extTfWriteOnly) {
		schemaDefinitionProperty.WriteOnly = true
	}

	// Use the default keyword in the parameter schema to specify the default value for an optional parameter. The default
	// value is the one that the server uses if the client does not supply the parameter value in the request.
	// Link: https://swagger.io/docs/specification/describing-parameters#default
	schemaDefinitionProperty.Default = property.Default

	return schemaDefinitionProperty, nil
}

func (o *SpecV2Resource) isBoolExtensionEnabled(extensions spec.Extensions, extension string) bool {
	if extensions != nil {
		if enabled, ok := extensions.GetBool(extension); ok && enabled {
			return true
		}
	}
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
		if o.isBoolExtensionEnabled(property.Extensions, extTfComputed) {
			return false, fmt.Errorf("optional computed property validation failed for property '%s': optional computed properties with default attributes should not have '%s' extension too", propertyName, extTfComputed)
		}
		return true, nil
	}
	return false, nil
}

// IsOptionalComputed returns true if the property is marked with the extension 'x-terraform-computed'
// This covers the use case where a property is not marked as readOnly but still is optional value that can come from the user or if not provided will be computed by the API. Example
//
// optional_computed: # optional property that the default value is NOT known at runtime
//  type: "string"
//  x-terraform-computed: true
func (o *SpecV2Resource) isOptionalComputed(propertyName string, property spec.Schema) (bool, error) {
	if o.isBoolExtensionEnabled(property.Extensions, extTfComputed) {
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
	return propertyType == TypeString || propertyType == TypeInt || propertyType == TypeFloat || propertyType == TypeBool
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
	if !o.isArrayItemPrimitiveType(itemsType) && !(itemsType == TypeObject) {
		return "", fmt.Errorf("array item type '%s' not supported", itemsType)
	}
	return itemsType, nil
}

func (o *SpecV2Resource) getPropertyType(property spec.Schema) (schemaDefinitionPropertyType, error) {
	if o.isArrayTypeProperty(property) {
		return TypeList, nil
	} else if isObject, _, err := o.isObjectProperty(property); isObject || err != nil {
		return TypeObject, err
	} else if property.Type.Contains("string") {
		return TypeString, nil
	} else if property.Type.Contains("integer") {
		return TypeInt, nil
	} else if property.Type.Contains("number") {
		return TypeFloat, nil
	} else if property.Type.Contains("boolean") {
		return TypeBool, nil
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
		return true, nil, fmt.Errorf("object is missing the nested schema definition or the ref is pointing to a non existing schema definition")
	}
	return false, nil, nil
}

func (o *SpecV2Resource) isArrayProperty(property spec.Schema) (bool, schemaDefinitionPropertyType, *SpecSchemaDefinition, error) {
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
	return o.getPreferredName(o.RootPathItem)
}

func (o *SpecV2Resource) getPreferredName(path spec.PathItem) string {
	preferredName, _ := path.Extensions.GetString(extTfResourceName)
	if preferredName == "" && path.Post != nil {
		preferredName, _ = path.Post.Extensions.GetString(extTfResourceName)
	}
	return preferredName
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
	for statusCode, response := range operation.Responses.StatusCodeResponses { //panics on ImportState if the swagger doesn't define status code responses
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
	if o.isBoolExtensionEnabled(response.Extensions, extTfResourcePollEnabled) {
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
func getResourceOverrideHost(rootPathItemPost *spec.Operation) string {
	if rootPathItemPost == nil {
		return ""
	}
	if resourceURL, exists := rootPathItemPost.Extensions.GetString(extTfResourceURL); exists && resourceURL != "" {
		return resourceURL
	}
	return ""
}
