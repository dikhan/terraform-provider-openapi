package openapi

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// SpecV3Resource defines a struct that implements the SpecResource interface and it's based on OpenAPI v3 specification
type SpecV3Resource struct {
	Name   string
	Region string
	// Path contains the full relative path to the resource e,g: /v1/resource
	Path string
	// SpecSchemaDefinition definition represents the representational state (aka model) of the resource
	SchemaDefinition *openapi3.Schema
	// RootPathItem contains info about the resource root path e,g: /resource, including the POST operation used to create instances of this resource
	RootPathItem *openapi3.PathItem
	// InstancePathItem contains info about the resource's instance /resource/{id}, including GET, PUT and REMOVE operations if applicable
	InstancePathItem *openapi3.PathItem

	// SchemaDefinitions contains all the definitions which might be needed in case the resource schema contains properties
	// of type object which in turn refer to other definitions
	SchemaDefinitions openapi3.Schemas

	Paths openapi3.Paths

	// Cached objects that are loaded once (when the corresponding function that loads the object is called the first time) and
	// on subsequent method calls the cached object is returned instead saving executing time.

	// specSchemaDefinitionCached is cached in GetResourceSchema() method
	specSchemaDefinitionCached *SpecSchemaDefinition
	// parentResourceInfoCached is cached in GetParentResourceInfo() method
	parentResourceInfoCached *ParentResourceInfo
	// resolvedPathCached is cached in getResourcePath() method
	resolvedPathCached string
}

var _ SpecResource = (*SpecV3Resource)(nil)

// newSpecV3Resource creates a SpecV3Resource with no region and default host
func newSpecV3Resource(path string, schemaDefinition *openapi3.Schema, rootPathItem, instancePathItem *openapi3.PathItem, schemaDefinitions openapi3.Schemas, paths map[string]*openapi3.PathItem) (*SpecV3Resource, error) {
	return newSpecV3ResourceWithConfig("", path, schemaDefinition, rootPathItem, instancePathItem, schemaDefinitions, paths)
}

func newSpecV3ResourceWithConfig(region, path string, schemaDefinition *openapi3.Schema, rootPathItem, instancePathItem *openapi3.PathItem, schemaDefinitions openapi3.Schemas, paths map[string]*openapi3.PathItem) (*SpecV3Resource, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if paths == nil {
		return nil, fmt.Errorf("paths must not be nil")
	}
	resource := &SpecV3Resource{
		Path:              path,
		Region:            region,
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

func (o *SpecV3Resource) GetResourceName() string {
	// TODO: implement multi-region support
	//if o.Region != "" {
	//	return fmt.Sprintf("%s_%s", o.Name, o.Region)
	//}
	return o.Name
}

// GetResourceName returns the name of the resource (including the version if applicable). The name is build from the resource
// root path /resource/{id} or if specified the value set in the x-terraform-resource-name extension is used instead along
// with the version (if applicable)
func (o *SpecV3Resource) buildResourceName() (string, error) {
	preferredName := ""
	if preferred := o.getResourceTerraformName(); preferred != "" {
		preferredName = preferred
	}
	fullResourceName, err := o.buildResourceNameFromPath(o.Path, preferredName)
	if err != nil {
		return "", err
	}
	// TODO: add support for subresources
	//parentResourceInfo := o.GetParentResourceInfo()
	//if parentResourceInfo != nil {
	//	fullResourceName = parentResourceInfo.fullParentResourceName + "_" + fullResourceName
	//}
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
func (o *SpecV3Resource) buildResourceNameFromPath(resourcePath, preferredName string) (string, error) {
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

func (o *SpecV3Resource) getResourceTerraformName() string {
	return o.getPreferredName(o.RootPathItem)
}

func (o *SpecV3Resource) getPreferredName(path *openapi3.PathItem) string {
	preferredName, _ := getExtensionAsJsonString(path.Extensions, extTfResourceName)
	if preferredName == "" && path.Post != nil {
		preferredName, _ = getExtensionAsJsonString(path.Post.Extensions, extTfResourceName)
	}
	return preferredName
}

func getExtensionAsJsonString(ext map[string]interface{}, name string) (string, bool) {
	ifaceVal, found := ext[name]
	if !found {
		return "", false
	}
	jsonVal, ok := ifaceVal.(json.RawMessage)
	if !ok {
		log.Printf("[DEBUG] extension '%s' is not a json string", name)
		return "", false
	}
	var val string
	if err := json.Unmarshal(jsonVal, &val); err != nil {
		log.Printf("[DEBUG] extension '%s' is not a json string - error: %v", name, err)
		return "", false
	}
	return val, true
}

func getExtensionAsJsonBool(ext map[string]interface{}, name string) (value bool, ok bool) {
	ifaceVal, found := ext[name]
	if !found {
		return false, false
	}
	jsonVal, ok := ifaceVal.(json.RawMessage)
	if !ok {
		log.Printf("[DEBUG] extension '%s' is not a json bool", name)
		return false, false
	}
	var val bool
	if err := json.Unmarshal(jsonVal, &val); err != nil {
		log.Printf("[DEBUG] extension '%s' is not a json bool - error: %v", name, err)
		return false, false
	}
	return val, true
}

func (o *SpecV3Resource) getHost() (string, error) {
	panic("implement me - getHost")
}

func (o *SpecV3Resource) getResourcePath(parentIDs []string) (string, error) {
	panic("implement me - getResourcePath")
}

func (o *SpecV3Resource) GetResourceSchema() (*SpecSchemaDefinition, error) {
	if o.specSchemaDefinitionCached != nil {
		log.Printf("[DEBUG] GetResourceSchema hit the cache for '%s'", o.Name)
		return o.specSchemaDefinitionCached, nil
	}
	specSchemaDefinition, err := o.getSchemaDefinitionWithOptions(o.SchemaDefinition, true)
	if err != nil {
		return nil, err
	}
	o.specSchemaDefinitionCached = specSchemaDefinition
	log.Printf("[DEBUG] GetResourceSchema cache loaded for '%s'", o.Name)
	return o.specSchemaDefinitionCached, nil
}

func (o *SpecV3Resource) getSchemaDefinition(schema *openapi3.Schema) (*SpecSchemaDefinition, error) {
	return o.getSchemaDefinitionWithOptions(schema, false)
}

func (o *SpecV3Resource) getSchemaDefinitionWithOptions(schema *openapi3.Schema, addParentProps bool) (*SpecSchemaDefinition, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema argument must not be nil")
	}
	schemaDefinition := &SpecSchemaDefinition{}
	schemaDefinition.Properties = SpecSchemaDefinitionProperties{}

	// This map ensures no duplicates will happen if the schema happens to have a parent id property. if so, it will be overridden with the expected parent property configuration (e,g: making the prop required)
	schemaProps := map[string]*SpecSchemaDefinitionProperty{}
	for propertyName, property := range schema.Properties {
		// TODO: support property.Ref
		schemaDefinitionProperty, err := o.createSchemaDefinitionProperty(propertyName, property.Value, schema.Required)
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
				pr, _ := o.createSchemaDefinitionProperty(parentPropertyName, &openapi3.Schema{Type: "string"}, []string{parentPropertyName})
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

func (o *SpecV3Resource) createSchemaDefinitionProperty(propertyName string, property *openapi3.Schema, requiredProperties []string) (*SpecSchemaDefinitionProperty, error) {
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
			// TODO: support property.Items.Ref
			if property.Items != nil && property.Items.Value != nil {
				schemaDefinitionProperty.Description = property.Items.Value.Description
			}
		}

		schemaDefinitionProperty.ArrayItemsType = itemsType
		schemaDefinitionProperty.SpecSchemaDefinition = itemsSchema // only diff than nil if type is object

		if o.isBoolExtensionEnabled(property.Extensions, extTfIgnoreOrder) || o.isBoolExtensionEnabled(property.Extensions, extIgnoreOrder) {
			schemaDefinitionProperty.IgnoreItemsOrder = true
		}

		log.Printf("[DEBUG] found array type property '%s' with items of type '%s'", propertyName, itemsType)
	}

	if preferredPropertyName, found := getExtensionAsJsonString(property.Extensions, extTfFieldName); found {
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

	// Use the default keyword in the parameter schema to specify the default value for an optional parameter. The default
	// value is the one that the server uses if the client does not supply the parameter value in the request.
	// Link: https://swagger.io/docs/specification/describing-parameters#default
	schemaDefinitionProperty.Default = property.Default

	return schemaDefinitionProperty, nil
}

func (o *SpecV3Resource) isBoolExtensionEnabled(extensions map[string]interface{}, extension string) bool {
	if extensions != nil {
		if enabled, ok := getExtensionAsJsonBool(extensions, extension); ok && enabled {
			return true
		}
	}
	return false
}

func (o *SpecV3Resource) isOptionalComputedProperty(propertyName string, property *openapi3.Schema, requiredProperties []string) (bool, error) {
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
func (o *SpecV3Resource) isOptionalComputedWithDefault(propertyName string, property *openapi3.Schema) (bool, error) {
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
func (o *SpecV3Resource) isOptionalComputed(propertyName string, property *openapi3.Schema) (bool, error) {
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

func (o *SpecV3Resource) isArrayItemPrimitiveType(propertyType schemaDefinitionPropertyType) bool {
	return propertyType == TypeString || propertyType == TypeInt || propertyType == TypeFloat || propertyType == TypeBool
}

func (o *SpecV3Resource) validateArrayItems(property *openapi3.Schema) (schemaDefinitionPropertyType, error) {
	// TODO: support .Ref
	if property.Items == nil || property.Items.Value == nil {
		return "", fmt.Errorf("array property is missing items schema definition")
	}
	// TODO: support .Ref
	if o.isArrayTypeProperty(property.Items.Value) {
		return "", fmt.Errorf("array property can not have items of type 'array'")
	}
	// TODO: support .Ref
	itemsType, err := o.getPropertyType(property.Items.Value)
	if err != nil {
		return "", err
	}
	if !o.isArrayItemPrimitiveType(itemsType) && !(itemsType == TypeObject) {
		return "", fmt.Errorf("array item type '%s' not supported", itemsType)
	}
	return itemsType, nil
}

func (o *SpecV3Resource) getPropertyType(property *openapi3.Schema) (schemaDefinitionPropertyType, error) {
	if o.isArrayTypeProperty(property) {
		return TypeList, nil
	} else if isObject, _, err := o.isObjectProperty(property); isObject || err != nil {
		return TypeObject, err
	} else if property.Type == "string" {
		return TypeString, nil
	} else if property.Type == "integer" {
		return TypeInt, nil
	} else if property.Type == "number" {
		return TypeFloat, nil
	} else if property.Type == "boolean" {
		return TypeBool, nil
	}
	return "", fmt.Errorf("non supported '%+v' type", property.Type)
}

func (o *SpecV3Resource) isObjectProperty(property *openapi3.Schema) (bool, *openapi3.Schema, error) {
	if o.isObjectTypeProperty(property) { // TODO: || property.Ref.Ref.GetURL() != nil {
		// Case of nested object schema
		if len(property.Properties) != 0 {
			return true, property, nil
		}
		// TODO: support external ref
		//// Case of external ref - in this case the type could be populated or not
		//if property.Ref.Ref.GetURL() != nil {
		//	schema, err := openapiutils.GetSchemaDefinition(o.SchemaDefinitions, property.Ref.String())
		//	if err != nil {
		//		return true, nil, fmt.Errorf("object ref is poitning to a non existing schema definition: %s", err)
		//	}
		//	return true, schema, nil
		//}
		return true, nil, fmt.Errorf("object is missing the nested schema definition or the ref is pointing to a non existing schema definition")
	}
	return false, nil, nil
}

func (o *SpecV3Resource) isArrayProperty(property *openapi3.Schema) (bool, schemaDefinitionPropertyType, *SpecSchemaDefinition, error) {
	if o.isArrayTypeProperty(property) {
		itemsType, err := o.validateArrayItems(property)
		if err != nil {
			return false, "", nil, err
		}
		if o.isArrayItemPrimitiveType(itemsType) {
			return true, itemsType, nil, nil
		}
		// This is the case where items must be object
		// TODO: support .Ref
		if isObject, schemaDefinition, err := o.isObjectProperty(property.Items.Value); isObject || err != nil {
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

func (o *SpecV3Resource) isArrayTypeProperty(property *openapi3.Schema) bool {
	return o.isOfType(property, "array")
}

func (o *SpecV3Resource) isObjectTypeProperty(property *openapi3.Schema) bool {
	return o.isOfType(property, "object")
}

func (o *SpecV3Resource) isOfType(property *openapi3.Schema, propertyType string) bool {
	return property.Type == propertyType
}

func (o *SpecV3Resource) isRequired(propertyName string, requiredProps []string) bool {
	var required = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (o *SpecV3Resource) ShouldIgnoreResource() bool {
	// TODO: support 'x-terraform-exclude-resource' extension
	return false
}

func (o *SpecV3Resource) getResourceOperations() specResourceOperations {
	panic("implement me - getResourceOperations")
}

func (o *SpecV3Resource) getTimeouts() (*specTimeouts, error) {
	// TODO: support "x-terraform-resource-timeout" extension
	timeout := 5 * time.Second
	return &specTimeouts{
		Post:   &timeout,
		Get:    &timeout,
		Put:    &timeout,
		Delete: &timeout,
	}, nil
}

func (o *SpecV3Resource) GetParentResourceInfo() *ParentResourceInfo {
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
