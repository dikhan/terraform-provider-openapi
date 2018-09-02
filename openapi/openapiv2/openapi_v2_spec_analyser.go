package openapiv2

import (
	"errors"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"log"
	"regexp"
	"strings"
)

// openApiV2SpecAnalyser defines an OpenAPISpecAnalyser implementation for OpenAPI v2 specification
// Forcing creation of this object via constructor so proper input validation is performed before creating the struct
// instance
type openApiV2SpecAnalyser struct {
	openAPIDocumentURL string
	d                  *loads.Document
}

// NewOpenApiV2SpecAnalyser creates an instance of openApiV2SpecAnalyser which implements the OpenAPISpecAnalyser interface
func NewOpenApiV2SpecAnalyser(openAPIDocumentURL string) (*openApiV2SpecAnalyser, error) {
	if openAPIDocumentURL == "" {
		return nil, errors.New("open api document url empty, please provide the url of the OpenAPI document")
	}
	apiSpec, err := loads.JSONSpec(openAPIDocumentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the OpenAPI document from '%s' - error = %s", openAPIDocumentURL, err)
	}
	return &openApiV2SpecAnalyser{
		d: apiSpec,
	}, nil
}

func (openApiSpec *openApiV2SpecAnalyser) GetHost() string {
	if openApiSpec.d.Spec().Host == "" {
		return openapiutils.GetHostFromURL(openApiSpec.openAPIDocumentURL)
	}
	return openApiSpec.d.Spec().Host
}

func (openApiSpec *openApiV2SpecAnalyser) GetTerraformCompliantResources() ([]openapi.OpenApiResource, error) {
	resources := []openapi.OpenApiResource{}
	for resourcePath, pathItem := range openApiSpec.d.Spec().Paths.Paths {
		resourceRootPath, resourceRoot, resourcePayloadSchemaDef, err := openApiSpec.isEndPointFullyTerraformResourceCompliant(resourcePath)
		if err != nil {
			log.Printf("[DEBUG] resource path '%s' not terraform compliant: %s", resourcePath, err)
			continue
		}
		resourceName, err := openApiSpec.getResourceName(resourcePath)
		if err != nil {
			log.Printf("[DEBUG] resource not figure out valid terraform resource name for '%s': %s", resourcePath, err)
			continue
		}
		r := resourceInfo{
			name:             resourceName,
			basePath:         asa.d.BasePath(),
			path:             resourceRootPath,
			host:             asa.d.Spec().Host,
			httpSchemes:      asa.d.Spec().Schemes,
			schemaDefinition: *resourcePayloadSchemaDef,
			createPathInfo:   *resourceRoot,
			pathInfo:         pathItem,
		}
		if !r.shouldIgnoreResource() {
			resources[resourceName] = r
		}
	}
	return resources, nil
}

func (openApiSpec *openApiV2SpecAnalyser) GetSecurity() openapi.OpenAPISecurity {
	return nil
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
func (openApiSpec *openApiV2SpecAnalyser) isEndPointFullyTerraformResourceCompliant(resourcePath string) (string, *spec.PathItem, *spec.Schema, error) {
	err := openApiSpec.validateInstancePath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}
	resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, err := openApiSpec.validateRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}
	err = openApiSpec.validateResourceSchemaDefinition(resourceRootPostSchemaDef)
	if err != nil {
		return "", nil, nil, err
	}
	return resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, nil
}

func (openApiSpec *openApiV2SpecAnalyser) validateInstancePath(path string) error {
	isResourceInstance, err := openApiSpec.isResourceInstanceEndPoint(path)
	if err != nil {
		return fmt.Errorf("error occurred while checking if path '%s' is a resource instance path", path)
	}
	if !isResourceInstance {
		return fmt.Errorf("path '%s' is not a resource instance path", path)
	}
	endPoint := openApiSpec.d.Spec().Paths.Paths[path]
	if endPoint.Get == nil {
		return fmt.Errorf("resource instance path '%s' missing required GET operation", path)
	}
	return nil
}

func (openApiSpec *openApiV2SpecAnalyser) validateRootPath(resourcePath string) (string, *spec.PathItem, *spec.Schema, error) {
	resourceRootPath, err := openApiSpec.findMatchingResourceRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}

	postExist := openApiSpec.postDefined(resourceRootPath)
	if !postExist {
		return "", nil, nil, fmt.Errorf("resource root path '%s' missing required POST operation", resourceRootPath)
	}

	resourceRootPathItem, _ := openApiSpec.d.Spec().Paths.Paths[resourceRootPath]
	resourceRootPostOperation := resourceRootPathItem.Post

	resourceRootPostSchemaDef, err := openApiSpec.getResourcePayloadSchemaDef(resourceRootPostOperation)
	if err != nil {
		return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation validation error: %s", resourceRootPath, err)
	}

	return resourceRootPath, &resourceRootPathItem, resourceRootPostSchemaDef, nil
}

func (openApiSpec *openApiV2SpecAnalyser) validateResourceSchemaDefinition(schema *spec.Schema) error {
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
func (openApiSpec *openApiV2SpecAnalyser) postDefined(resourceRootPath string) bool {
	b, exists := openApiSpec.d.Spec().Paths.Paths[resourceRootPath]
	if !exists || b.Post == nil {
		return false
	}
	return true
}

func (openApiSpec *openApiV2SpecAnalyser) getResourcePayloadSchemaDef(resourceRootPostOperation *spec.Operation) (*spec.Schema, error) {
	ref, err := openApiSpec.getResourcePayloadSchemaRef(resourceRootPostOperation)
	if err != nil {
		return nil, err
	}
	payloadDefName, err := openApiSpec.getPayloadDefName(ref)
	if err != nil {
		return nil, err
	}
	payloadDefinition, exists := openApiSpec.d.Spec().Definitions[payloadDefName]
	if !exists {
		return nil, fmt.Errorf("missing schema definition in the swagger file with the supplied ref '%s'", ref)
	}
	return &payloadDefinition, nil
}

func (openApiSpec *openApiV2SpecAnalyser) getResourcePayloadSchemaRef(resourceRootPostOperation *spec.Operation) (string, error) {
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
func (openApiSpec *openApiV2SpecAnalyser) getPayloadDefName(ref string) (string, error) {
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
func (openApiSpec *openApiV2SpecAnalyser) resourceInstanceRegex() (*regexp.Regexp, error) {
	r, err := regexp.Compile(resourceInstanceRegex)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while compiling the resourceInstanceRegex regex '%s': %s", resourceInstanceRegex, err)
	}
	return r, nil
}

// isResourceInstanceEndPoint checks if the given path is of form /resource/{id}
func (openApiSpec *openApiV2SpecAnalyser) isResourceInstanceEndPoint(p string) (bool, error) {
	r, err := openApiSpec.resourceInstanceRegex()
	if err != nil {
		return false, err
	}
	return r.MatchString(p), nil
}

// getResourceName gets the name of the resource from a path /resource/{id}
func (openApiSpec *openApiV2SpecAnalyser) getResourceName(resourcePath string) (string, error) {
	nameRegex, err := regexp.Compile(resourceNameRegex)
	if err != nil {
		return "", fmt.Errorf("an error occurred while compiling the resourceNameRegex regex '%s': %s", resourceNameRegex, err)
	}
	var resourceName string
	matches := nameRegex.FindStringSubmatch(resourcePath)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find a valid name for resource instance path '%s'", resourcePath)
	}
	resourceName = strings.Replace(matches[len(matches)-1], "/", "", -1)
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

// findMatchingResourceRootPath returns the corresponding POST root and path for a given end point
// Example: Given 'resourcePath' being "/users/{username}" the result could be "/users" or "/users/" depending on
// how the POST operation (resourceRootPath) of the given resource is defined in swagger.
// If there is no match the returned string will be empty
func (openApiSpec *openApiV2SpecAnalyser) findMatchingResourceRootPath(resourcePath string) (string, error) {
	r, err := openApiSpec.resourceInstanceRegex()
	if err != nil {
		return "", err
	}
	result := r.FindStringSubmatch(resourcePath)
	log.Printf("[DEBUG] resource root path match result - %s", result)
	if len(result) != 2 {
		return "", fmt.Errorf("resource instance path '%s' missing valid resource root path, more than two results returned from match '%s'", resourcePath, result)
	}

	resourceRootPath := result[1] // e,g: /v1/cdns/{id} /v1/cdns/

	if _, exists := openApiSpec.d.Spec().Paths.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path with trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	// Handles the case where the swagger file root path does not have a trailing slash in the path
	resourceRootPath = strings.TrimRight(resourceRootPath, "/")
	if _, exists := openApiSpec.d.Spec().Paths.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path without trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	return "", fmt.Errorf("resource instance path '%s' missing resource root path", resourcePath)
}
