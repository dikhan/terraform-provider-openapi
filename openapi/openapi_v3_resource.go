package openapi

import (
	"encoding/json"
	"fmt"
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

func getExtensionAsJsonString(ext map[string]interface{}, name string) (string, error) {
	ifaceVal, found := ext[name]
	jsonVal, ok := ifaceVal.(json.RawMessage)
	if !found || !ok {
		return "", fmt.Errorf("extension '%s' is not a json string", name)
	}
	var val string
	if err := json.Unmarshal(jsonVal, &val); err != nil {
		return "", fmt.Errorf("extension '%s' is not a json string", name)
	}
	return val, nil
}

func (o *SpecV3Resource) getHost() (string, error) {
	panic("implement me - getHost")
}

func (o *SpecV3Resource) getResourcePath(parentIDs []string) (string, error) {
	panic("implement me - getResourcePath")
}

func (o *SpecV3Resource) GetResourceSchema() (*SpecSchemaDefinition, error) {
	// TODO: replace stub with core impl -- this is the heart. <3
	return &SpecSchemaDefinition{
		Properties: SpecSchemaDefinitionProperties{
			{
				Name:        "prop1",
				Description: "fake desc",
				Required:    true,
				Type:        TypeString,
			},
		},
	}, nil
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
	panic("implement me - GetParentResourceInfo")
}
