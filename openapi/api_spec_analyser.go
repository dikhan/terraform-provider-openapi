package openapi

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

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
		resourceRootPath, resourceRoot, resourcePayloadSchemaDef, err := asa.isEndPointFullyTerraformResourceCompliant(resourcePath)
		if err != nil {
			log.Printf("[DEBUG] resource path '%s' not terraform compliant: %s", resourcePath, err)
			continue
		}

		r := resourceInfo{
			basePath:         asa.d.BasePath(),
			path:             resourceRootPath,
			host:             asa.d.Spec().Host,
			httpSchemes:      asa.d.Spec().Schemes,
			schemaDefinition: *resourcePayloadSchemaDef,
			createPathInfo:   *resourceRoot,
			pathInfo:         pathItem,
		}

		if r.shouldIgnoreResource() {
			continue
		}

		resourceName, err := r.getResourceName()
		if err != nil {
			log.Printf("[DEBUG] could not figure out the resource name for '%s': %s", resourcePath, err)
			continue
		}

		isMultiRegion, regions := r.isMultiRegionResource(asa.d.Spec().Extensions)
		if isMultiRegion {
			log.Printf("[INFO] resource '%s' is configured with host override AND multi region; creating reasource per region", r.path)
			for regionName, regionHost := range regions {
				resourceRegionName := fmt.Sprintf("%s_%s", resourceName, regionName)
				regionResource := resourceInfo{}
				regionResource = r
				regionResource.host = regionHost
				log.Printf("[INFO] multi region resource: name = %s, region = %s, host = %s", regionName, resourceRegionName, regionHost)
				resources[resourceRegionName] = regionResource
			}
			continue
		}

		hostOverride := r.getResourceOverrideHost()
		// if the override host is multi region then something must be wrong with the multi region configuration, failing to let the user know so they can fix the configuration
		if isMultiRegionHost, _ := r.isMultiRegionHost(hostOverride); isMultiRegionHost {
			return nil, fmt.Errorf("multi region configuration for resource '%s' is wrong, please check the multi region configuration in the swagger file is right for that resource", resourceName)
		}
		// Fall back to override the host if value is not empty; otherwise global host will be used as usual
		if hostOverride != "" {
			log.Printf("[INFO] resource '%s' is configured with host override, API calls will be made against '%s' instead of '%s'", r.path, hostOverride, asa.d.Spec().Host)
			r.host = hostOverride
		}
		resources[resourceName] = r
	}
	return resources, nil
}

// isEndPointFullyTerraformResourceCompliant returns true only if:
// - The path given 'resourcePath' is an instance path (e,g: "/users/{username}")
// - The path given has GET operation defined (required). PUT and DELETE are optional
// - The root path for the given path 'resourcePath' is found (e,g: "/users")
// - The root path for the given path 'resourcePath' has mandatory POST operation defined
// - The root path for the given path 'resourcePath' has a parameter of type 'body' with a schema property referencing to an existing definition object
// - The root path POST payload definition and the returned object in the response matches. Similarly, the GET operation should also have the same return object
// - The resource schema definition must contain a field that uniquelly identifies the resource or have a field with the 'x-terraform-id' extension set to true
// For instance, if resourcePath was "/users/{id}" and paths contained the following entries and implementations:
// paths:
//   /v1/users:
//     post:
//		 parameters:
//		 - in: "body"
//		   name: "body"
//		   description: "user to create"
//		   required: true
//		   schema:
//		     $ref: "#/definitions/User"
//		 responses:
//		   201:
//		     description: "successful operation"
//		     schema:
//		       $ref: "#/definitions/User"
//   /v1/users/{id}:
//	   get:
//	     parameters:
//	       - name: "id"
//	         in: "path"
//	         description: "The user id that needs to be fetched"
//	         required: true
//	         type: "string"
//	     responses:
//	       200:
//	      	 description: "successful operation"
//	         schema:
//	           $ref: "#/definitions/User"
// definitions:
//   Users:
//     type: "object"
//     required:
//       - name
//     properties:
//       id:
//         type: "string"
//         readOnly: true
//       name:
//         type: "string"
// then the expected returned value is true. Otherwise if the above criteria is not met, it is considered that
// the resourcePath provided is not terraform resource compliant.
func (asa apiSpecAnalyser) isEndPointFullyTerraformResourceCompliant(resourcePath string) (string, *spec.PathItem, *spec.Schema, error) {
	err := asa.validateInstancePath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}
	resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, err := asa.validateRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}
	err = asa.validateResourceSchemaDefinition(resourceRootPostSchemaDef)
	if err != nil {
		return "", nil, nil, err
	}
	return resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, nil
}

func (asa apiSpecAnalyser) validateInstancePath(path string) error {
	isResourceInstance, err := asa.isResourceInstanceEndPoint(path)
	if err != nil {
		return fmt.Errorf("error occurred while checking if path '%s' is a resource instance path", path)
	}
	if !isResourceInstance {
		return fmt.Errorf("path '%s' is not a resource instance path", path)
	}
	endPoint := asa.d.Spec().Paths.Paths[path]
	if endPoint.Get == nil {
		return fmt.Errorf("resource instance path '%s' missing required GET operation", path)
	}
	return nil
}

func (asa apiSpecAnalyser) validateRootPath(resourcePath string) (string, *spec.PathItem, *spec.Schema, error) {
	resourceRootPath, err := asa.findMatchingResourceRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}

	postExist := asa.postDefined(resourceRootPath)
	if !postExist {
		return "", nil, nil, fmt.Errorf("resource root path '%s' missing required POST operation", resourceRootPath)
	}

	resourceRootPathItem, _ := asa.d.Spec().Paths.Paths[resourceRootPath]
	resourceRootPostOperation := resourceRootPathItem.Post

	resourceRootPostSchemaDef, err := asa.getResourcePayloadSchemaDef(resourceRootPostOperation)
	if err != nil {
		return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation validation error: %s", resourceRootPath, err)
	}

	return resourceRootPath, &resourceRootPathItem, resourceRootPostSchemaDef, nil
}

func (asa apiSpecAnalyser) validateResourceSchemaDefinition(schema *spec.Schema) error {
	identifier := ""
	for propertyName, property := range schema.Properties {
		if propertyName == "id" {
			identifier = propertyName
			continue
		}
		if exists, useAsIdentifier := property.Extensions.GetBool(extTfID); exists && useAsIdentifier {
			identifier = propertyName
			break
		}
	}
	if identifier == "" {
		return fmt.Errorf("resource schema is missing a property that uniquely identifies the resource, either a property named 'id' or a property with the extension '%s' set to true", extTfID)
	}
	return nil
}

// postIsPresent checks if the given resource has a POST implementation returning true if the path is found
// in paths and the path exposes a POST operation
func (asa apiSpecAnalyser) postDefined(resourceRootPath string) bool {
	b, exists := asa.d.Spec().Paths.Paths[resourceRootPath]
	if !exists || b.Post == nil {
		return false
	}
	return true
}

func (asa apiSpecAnalyser) getResourcePayloadSchemaDef(resourceRootPostOperation *spec.Operation) (*spec.Schema, error) {
	ref, err := asa.getResourcePayloadSchemaRef(resourceRootPostOperation)
	if err != nil {
		return nil, err
	}
	payloadDefName, err := asa.getPayloadDefName(ref)
	if err != nil {
		return nil, err
	}
	payloadDefinition, exists := asa.d.Spec().Definitions[payloadDefName]
	if !exists {
		return nil, fmt.Errorf("missing schema definition in the swagger file with the supplied ref '%s'", ref)
	}
	return &payloadDefinition, nil
}

func (asa apiSpecAnalyser) getResourcePayloadSchemaRef(resourceRootPostOperation *spec.Operation) (string, error) {
	if len(resourceRootPostOperation.Parameters) <= 0 {
		return "", fmt.Errorf("operation does not have parameters defined")
	}

	// A given operation might have multiple parameters, looking for required 'body' parameter type
	var bodyParameter spec.Parameter
	var bodyParamCounter int
	for _, parameter := range resourceRootPostOperation.Parameters {
		if parameter.In == "body" {
			bodyParamCounter++
			bodyParameter = parameter
		}
	}
	if bodyParamCounter == 0 {
		return "", fmt.Errorf("operation is missing required 'body' type parameter")
	}
	if bodyParamCounter > 1 {
		return "", fmt.Errorf("operation contains multiple 'body' parameters")
	}
	payloadDefinitionSchemaRef := bodyParameter.Schema
	if payloadDefinitionSchemaRef == nil {
		return "", fmt.Errorf("operation is missing the ref to the schema definition")
	}
	if payloadDefinitionSchemaRef.Ref.String() == "" {
		return "", fmt.Errorf("operation has an invalid schema definition ref empty")
	}
	return payloadDefinitionSchemaRef.Ref.String(), nil
}

// getPayloadDefName only supports references to the same document. External references like URLs is not supported at the moment
func (asa apiSpecAnalyser) getPayloadDefName(ref string) (string, error) {
	reg, err := regexp.Compile(swaggerResourcePayloadDefinitionRegex)
	if err != nil {
		return "", fmt.Errorf("an error occurred while compiling the swaggerResourcePayloadDefinitionRegex regex '%s': %s", swaggerResourcePayloadDefinitionRegex, err)
	}
	payloadDefName := reg.FindStringSubmatch(ref)[0]
	if payloadDefName == "" {
		return "", fmt.Errorf("could not find a valid definition name for '%s'", ref)
	}
	return payloadDefName, nil
}

// resourceInstanceRegex loads up the regex specified in const resourceInstanceRegex
// If the regex is not able to compile the regular expression the function exists calling os.Exit(1) as
// there is the regex is completely busted
func (asa apiSpecAnalyser) resourceInstanceRegex() (*regexp.Regexp, error) {
	r, err := regexp.Compile(resourceInstanceRegex)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while compiling the resourceInstanceRegex regex '%s': %s", resourceInstanceRegex, err)
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

// findMatchingResourceRootPath returns the corresponding POST root and path for a given end point
// Example: Given 'resourcePath' being "/users/{username}" the result could be "/users" or "/users/" depending on
// how the POST operation (resourceRootPath) of the given resource is defined in swagger.
// If there is no match the returned string will be empty
func (asa apiSpecAnalyser) findMatchingResourceRootPath(resourcePath string) (string, error) {
	r, err := asa.resourceInstanceRegex()
	if err != nil {
		return "", err
	}
	result := r.FindStringSubmatch(resourcePath)
	log.Printf("[DEBUG] resource root path match result - %s", result)
	if len(result) != 2 {
		return "", fmt.Errorf("resource instance path '%s' missing valid resource root path, more than two results returned from match '%s'", resourcePath, result)
	}

	resourceRootPath := result[1] // e,g: /v1/cdns/{id} /v1/cdns/

	if _, exists := asa.d.Spec().Paths.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path with trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	// Handles the case where the swagger file root path does not have a trailing slash in the path
	resourceRootPath = strings.TrimRight(resourceRootPath, "/")
	if _, exists := asa.d.Spec().Paths.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path without trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	return "", fmt.Errorf("resource instance path '%s' missing resource root path", resourcePath)
}
