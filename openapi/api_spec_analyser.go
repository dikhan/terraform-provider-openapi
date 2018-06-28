package openapi

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

const resourceVersionRegex = "(/v[0-9]*/)"
const resourceNameRegex = "(/\\w*/)+{.*}"
const resourceInstanceRegex = "((?:.*)){.*}"
const swaggerResourcePayloadDefinitionRegex = "(\\w+)[^//]*$"

// apiSpecAnalyser analyses the swagger doc and provides helper methods to retrieve all the end points that can
// be used as terraform resources. These endpoints have to meet certain criteria to be considered eligible resources
// as explained below:
// A resource is considered any end point that meets the following:
// 	- POST operation on the root path (e,g: api/users)
//	- GET operation on the instance path (e,g: api/users/{id}). Other operations like DELETE, PUT are optional
// In the example above, the resource name would be 'users'.
// Versioning is also supported, thus if the endpoint above had been api/v1/users the corresponding resouce name would
// have been 'users_v1'
type apiSpecAnalyser struct {
	d *loads.Document
}

func (asa apiSpecAnalyser) getResourcesInfo() (resourcesInfo, error) {
	resources := resourcesInfo{}
	for resourcePath, pathItem := range asa.d.Spec().Paths.Paths {
		isEndPointTerraformResourceCompliant, err := asa.isEndPointTerraformResourceCompliant(resourcePath, asa.d.Spec().Paths.Paths)
		if err != nil {
			return resources, err
		}
		if !isEndPointTerraformResourceCompliant {
			continue
		}
		resourceRootPath, err := asa.findMatchingResourceRootPath(resourcePath, asa.d.Spec().Paths.Paths)
		resourceName, err := asa.getResourceName(resourcePath)
		if err != nil {
			return nil, err
		}
		resourcePayloadSchemaDef, err := asa.getResourcePayloadSchemaDef(resourceRootPath)
		if err != nil {
			return nil, err
		}
		r := resourceInfo{
			name:             resourceName,
			basePath:         asa.d.BasePath(),
			path:             resourceRootPath,
			host:             asa.d.Spec().Host,
			httpSchemes:      asa.d.Spec().Schemes,
			schemaDefinition: *resourcePayloadSchemaDef,
			createPathInfo:   asa.d.Spec().Paths.Paths[resourceRootPath],
			pathInfo:         pathItem,
		}
		resources[resourceName] = r
	}
	return resources, nil
}

func (asa apiSpecAnalyser) getResourcePayloadSchemaDef(resourceRootPath string) (*spec.Schema, error) {
	path, exist := asa.d.Spec().Paths.Paths[resourceRootPath]
	if !exist {
		return nil, fmt.Errorf("path %s does not exists in the swagger file", resourceRootPath)
	}
	if path.Post == nil {
		return nil, fmt.Errorf("path %s POST operation missing", resourceRootPath)
	}
	if len(path.Post.Parameters) <= 0 {
		return nil, fmt.Errorf("path %s POST operation is missing paremeters", resourceRootPath)
	}
	payloadDefinitionSchemaRef := path.Post.Parameters[0].Schema
	if payloadDefinitionSchemaRef == nil {
		return nil, fmt.Errorf("resource %s POST operation is missing the ref to the schema definition", resourceRootPath)
	}
	ref := payloadDefinitionSchemaRef.Ref.String()
	reg, err := regexp.Compile(swaggerResourcePayloadDefinitionRegex)
	if err != nil {
		return nil, fmt.Errorf("something really wrong happened if the ref reg can't be compiled")
	}
	payloadDefName := reg.FindStringSubmatch(ref)[0]
	if payloadDefName == "" {
		return nil, fmt.Errorf("could not find a submatch in the ref regex %s for the reference supplied %s", ref, swaggerResourcePayloadDefinitionRegex)
	}
	payloadDefinition, exists := asa.d.Spec().Definitions[payloadDefName]
	if !exists {
		return nil, fmt.Errorf("could not find any schema definition in the swagger file with the supplied ref %s", ref)
	}
	return &payloadDefinition, nil
}

// isEndPointTerraformResourceCompliant returns true only if the path given 'resourcePath' exposes POST and GET operations.
// PUT and DELETE are optional operations.
// For instance, if resourcePath was "/users/{username}" and paths contained the following entries and implementations:
// - "/users"
// 		- POST
// - "/users/{username}"
// 		- GET
// 		- PUT (optional)
// 		- DELETE (optional)
// then the expected returned value is true. Otherwise if the above criteria is not met, it is considered that
// the resourcePath provided is not terraform resource compliant.
func (asa apiSpecAnalyser) isEndPointTerraformResourceCompliant(resourcePath string, paths map[string]spec.PathItem) (bool, error) {
	isResourceInstance, err := asa.isResourceInstanceEndPoint(resourcePath)
	if err != nil {
		return false, err
	}
	if isResourceInstance {
		endPoint := paths[resourcePath]
		if endPoint.Get == nil {
			return false, nil
		}
		resourceRootPath, err := asa.findMatchingResourceRootPath(resourcePath, paths)
		if err != nil {
			return false, err
		}
		if resourceRootPath == "" {
			log.Printf("[DEBUG] could not find root path for resource - %s", resourcePath)
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

// postIsPresent checks if the given resource has a POST implementation returning true if the path is found
// in paths and the path exposes a POST operation
func (asa apiSpecAnalyser) postIsPresent(resourceRootPath string, paths map[string]spec.PathItem) bool {
	b := paths[resourceRootPath]
	if &b == nil || b.Post == nil {
		return false
	}
	return true
}

// resourceInstanceRegex loads up the regex specified in const resourceInstanceRegex
// If the regex is not able to compile the regular expression the function exists calling os.Exit(1) as
// there is the regex is completely busted
func (asa apiSpecAnalyser) resourceInstanceRegex() (*regexp.Regexp, error) {
	r, err := regexp.Compile(resourceInstanceRegex)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] the resful endpoint regex does not seem to be configured properly [%s] %s", resourceInstanceRegex, err)
	}
	return r, nil
}

// isResourceInstanceEndPoint checks if the given path is of form /resource/{id}
func (asa apiSpecAnalyser) isResourceInstanceEndPoint(p string) (bool, error) {
	r, err := asa.resourceInstanceRegex()
	if err != nil {
		return false, err
	}
	return r.MatchString(p), nil
}

// getResourceName gets the name of the resource from a path /resource/{id}
func (asa apiSpecAnalyser) getResourceName(resourcePath string) (string, error) {
	isResourceInstance, err := asa.isResourceInstanceEndPoint(resourcePath)
	if err != nil {
		return "", err
	}
	if !isResourceInstance {
		return "", fmt.Errorf("resource names can only be extracted from valid resful resource instance paths, e,g: /resource_name/{id} - the path passed in %s was not valid", resourcePath)
	}

	nameRegex, err := regexp.Compile(resourceNameRegex)
	if err != nil {
		return "", fmt.Errorf("[ERROR] something is really wrong with the resource name regex [%s] %s", resourceNameRegex, err)
	}
	resourceName := strings.Replace(nameRegex.FindStringSubmatch(resourcePath)[1], "/", "", -1)

	versionRegex, err := regexp.Compile(resourceVersionRegex)
	if err != nil {
		return "", fmt.Errorf("[ERROR] something is really wrong with the resource version regex [%s] %s", resourceVersionRegex, err)
	}
	versionMatches := versionRegex.FindStringSubmatch(resourcePath)
	if len(versionMatches) != 0 {
		version := strings.Replace(versionRegex.FindStringSubmatch(resourcePath)[1], "/", "", -1)
		resourceNameWithVersion := fmt.Sprintf("%s_%s", resourceName, version)
		return resourceNameWithVersion, nil
	}
	return resourceName, nil
}

// findMatchingResourceRootPath returns the corresponding root path for a given end point
// Example: Given 'resourcePath' being "/users/{username}" the result could be "/users" or "/users/" depending on
// how the POST operation (resourceRootPath) of the given resource is defined in swagger.
// If there is no match the returned string will be empty
func (asa apiSpecAnalyser) findMatchingResourceRootPath(resourcePath string, paths map[string]spec.PathItem) (string, error) {
	r, err := asa.resourceInstanceRegex()
	if err != nil {
		return "", err
	}
	result := r.FindStringSubmatch(resourcePath)
	log.Printf("[DEBUG] findMatchingResourceRootPath result - %s", result)
	if len(result) != 2 {
		return "", nil
	}

	resourceRootPath := result[1] // e,g: /v1/cdns/{id} /v1/cdns/

	// Handles the case where the swagger file root path has a trailing slash in the path
	postExist := asa.postIsPresent(resourceRootPath, paths)
	if postExist {
		log.Printf("[DEBUG] found resource root path with trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	// Handles the case where the swagger file root path does not have a trailing slash in the path
	resourceRootPath = strings.TrimRight(resourceRootPath, "/")
	postExist = asa.postIsPresent(resourceRootPath, paths)
	if postExist {
		log.Printf("[DEBUG] found resource root path without trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	log.Printf("[DEBUG] end point %s missing POST operation", resourceRootPath)
	return "", nil
}
