package openapi

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dikhan/terraform-provider-openapi/v2/openapi/openapiutils"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

const extTfResourceRegionsFmt = "x-terraform-resource-regions-%s"

// specV2Analyser defines an SpecAnalyser implementation for OpenAPI v2 specification
// Forcing creation of this object via constructor so proper input validation is performed before creating the struct
// instance
type specV2Analyser struct {
	openAPIDocumentURL string
	d                  *loads.Document
}

// newSpecAnalyserV2 creates an instance of specV2Analyser which implements the SpecAnalyser interface
// This implementation provides an analyser that understands an OpenAPI v2 document
func newSpecAnalyserV2(openAPIDocumentFilename string) (*specV2Analyser, error) {
	if openAPIDocumentFilename == "" {
		return nil, errors.New("open api document filename argument empty, please provide the url of the OpenAPI document")
	}
	apiSpec, err := loads.JSONSpec(openAPIDocumentFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the OpenAPI document from '%s' - error = %s", openAPIDocumentFilename, err)
	}
	apiSpec, err = apiSpec.Expanded()
	if err != nil {
		return nil, fmt.Errorf("failed to expand the OpenAPI document from '%s' - error = %s", openAPIDocumentFilename, err)
	}
	return &specV2Analyser{
		d:                  apiSpec,
		openAPIDocumentURL: openAPIDocumentFilename,
	}, nil
}

func (specAnalyser *specV2Analyser) createMultiRegionResources(regions []string, resourceRootPath string, resourceRoot, pathItem spec.PathItem, resourcePayloadSchemaDef *spec.Schema) ([]SpecResource, error) {
	var resources []SpecResource
	for _, regionName := range regions {
		r, err := newSpecV2ResourceWithRegion(regionName, resourceRootPath, *resourcePayloadSchemaDef, resourceRoot, pathItem, specAnalyser.d.Spec().Definitions, specAnalyser.d.Spec().Paths.Paths)
		if err != nil {
			return nil, fmt.Errorf("failed to create a resource with region: %s", err)
		}
		log.Printf("[INFO] multi region resource name = %s, region = '%s'", r.GetResourceName(), regionName)
		resources = append(resources, r)
	}
	return resources, nil
}

func (specAnalyser *specV2Analyser) GetTerraformCompliantDataSources() []SpecResource {
	var dataSources []SpecResource
	spec := specAnalyser.d.Spec()
	paths := spec.Paths
	for resourcePath, pathItem := range paths.Paths {
		schemaDefinition, err := specAnalyser.isEndPointTerraformDataSourceCompliant(pathItem)
		if err != nil {
			log.Printf("[DEBUG] resource path '%s' not terraform data source compliant: %s", resourcePath, err)
			continue
		}

		d, err := newSpecV2DataSource(resourcePath, *schemaDefinition, pathItem, specAnalyser.d.Spec().Paths.Paths)
		if err != nil {
			log.Printf("[WARN] ignoring data source '%s' due to an error while creating a creating the SpecV2Resource: %s", resourcePath, err)
			continue
		}

		log.Printf("[INFO] found terraform compliant data source [name='%s', rootPath='%s']", d.GetResourceName(), resourcePath)
		dataSources = append(dataSources, d)
	}
	return dataSources
}

func (specAnalyser *specV2Analyser) GetTerraformCompliantResources() ([]SpecResource, error) {
	var resources []SpecResource
	start := time.Now()
	spec := specAnalyser.d.Spec()
	paths := spec.Paths
	for resourcePath, pathItem := range paths.Paths {
		resourceRootPath, resourceRoot, resourcePayloadSchemaDef, err := specAnalyser.isEndPointFullyTerraformResourceCompliant(resourcePath)
		if err != nil {
			log.Printf("[DEBUG] resource path '%s' not terraform compliant: %s", resourcePath, err)
			continue
		}

		isMultiRegion, regions, err := specAnalyser.isMultiRegionResource(resourceRoot, specAnalyser.d.Spec().Extensions)
		if err != nil {
			log.Printf("multi region configuration for resource '%s' is not valid: ", err)
			continue
		}
		if isMultiRegion {
			log.Printf("[INFO] resource '%s' is configured with host override AND multi region; creating one reasource per region", resourceRootPath)
			multiRegionResources, err := specAnalyser.createMultiRegionResources(regions, resourceRootPath, *resourceRoot, pathItem, resourcePayloadSchemaDef)
			if err != nil {
				log.Printf("[WARN] ignoring multiregion resource '%s' due to an error: %s", resourceRootPath, err)
				continue
			}
			resources = append(resources, multiRegionResources...)
			continue
		}

		r, err := newSpecV2Resource(resourceRootPath, *resourcePayloadSchemaDef, *resourceRoot, pathItem, specAnalyser.d.Spec().Definitions, specAnalyser.d.Spec().Paths.Paths)
		if err != nil {
			log.Printf("[WARN] ignoring resource '%s' due to an error while creating a creating the SpecV2Resource: %s", resourceRootPath, err)
			continue
		}

		err = specAnalyser.validateSubResourceTerraformCompliance(*r)
		if err != nil {
			log.Printf("[WARN] ignoring subresource name='%s' with rootPath='%s' due to not meeting validation requirements: %s", r.GetResourceName(), resourceRootPath, err)
			continue
		}

		log.Printf("[INFO] found terraform compliant resource [name='%s', rootPath='%s', instancePath='%s']", r.GetResourceName(), resourceRootPath, resourcePath)
		resources = append(resources, r)
	}
	log.Printf("[INFO] found %d terraform compliant resources (time: %s)", len(resources), time.Since(start))
	return resources, nil
}

func (specAnalyser *specV2Analyser) validateSubResourceTerraformCompliance(r SpecV2Resource) error {
	parentResourceInfo := r.GetParentResourceInfo()
	if parentResourceInfo != nil {
		resourcePath := r.Path
		for _, parentInstanceURIs := range parentResourceInfo.parentInstanceURIs {
			if pathExists, _ := specAnalyser.pathExists(parentInstanceURIs); !pathExists {
				return fmt.Errorf("subresource with path '%s' is missing parent path instance definition '%s'", resourcePath, parentInstanceURIs)
			}
		}
		for _, parentURI := range parentResourceInfo.parentURIs {
			parentPathExists, parentPathItem := specAnalyser.pathExists(parentURI)
			if !parentPathExists {
				return fmt.Errorf("subresource with path '%s' is missing parent root path definition '%s'", resourcePath, parentURI)
			}
			parentResource := SpecV2Resource{RootPathItem: parentPathItem}
			if parentResource.ShouldIgnoreResource() {
				return fmt.Errorf("subresource with path '%s' contains a parent %s that is marked as ignored, therefore ignoring the subresource too", resourcePath, parentURI)
			}
		}
	}
	return nil
}

func (specAnalyser *specV2Analyser) pathExists(path string) (bool, spec.PathItem) {
	p, exists := specAnalyser.d.Spec().Paths.Paths[path]
	if !exists {
		log.Printf("[WARN] path %s not found, falling back to checking if the path with trailing slash %s/ exists", path, path)
		p, exists = specAnalyser.d.Spec().Paths.Paths[path+"/"]
		if !exists {
			return false, spec.PathItem{}
		}
	}
	return true, p
}

// isMultiRegionResource returns true on ly if:
// - the value is parametrized following the pattern: some.subdomain.${keyword}.domain.com, where ${keyword} must be present in the string, otherwise the resource will not be considered multi region
// - there is a matching 'x-terraform-resource-regions-${keyword}' extension defined in the swagger root level (extensions passed in), where ${keyword} will be the value of the parameter in the above URL
// - and finally the value of the extension is an array of strings containing the different regions where the resource can be created
func (specAnalyser *specV2Analyser) isMultiRegionResource(resourceRoot *spec.PathItem, extensions spec.Extensions) (bool, []string, error) {
	overrideHost := getResourceOverrideHost(resourceRoot.Post)
	if overrideHost == "" {
		return false, nil, nil
	}
	isMultiRegionHost, regex := openapiutils.IsMultiRegionHost(overrideHost)
	if !isMultiRegionHost {
		return false, nil, nil
	}
	region := regex.FindStringSubmatch(overrideHost)
	if len(region) != 5 {
		return false, nil, fmt.Errorf("override host %s provided does not comply with expected regex format", overrideHost)
	}
	regionIdentifier := region[3]
	regionExtensionName := specAnalyser.getResourceRegionExtensionName(regionIdentifier)
	if resourceRegions, exists := openapiutils.StringExtensionExists(extensions, regionExtensionName); exists {
		resourceRegions = strings.Replace(resourceRegions, " ", "", -1)
		regions := strings.Split(resourceRegions, ",")
		if len(regions) < 1 {
			return false, nil, fmt.Errorf("could not find any region for '%s' matching region extension %s: '%s'", regionIdentifier, regionExtensionName, resourceRegions)
		}
		apiRegions := []string{}
		for _, region := range regions {
			apiRegions = append(apiRegions, region)
		}
		if len(apiRegions) < 1 {
			return false, nil, fmt.Errorf("could not build properly the resource region map for '%s' matching region extension %s: '%s'", regionIdentifier, regionExtensionName, resourceRegions)
		}
		return true, apiRegions, nil
	}
	return false, nil, fmt.Errorf("missing matching '%s' root level region extension '%s'", regionIdentifier, regionExtensionName)
}

func (specAnalyser *specV2Analyser) getResourceRegionExtensionName(regionIdentifier string) string {
	return fmt.Sprintf(extTfResourceRegionsFmt, regionIdentifier)
}

func (specAnalyser *specV2Analyser) GetSecurity() SpecSecurity {
	return &specV2Security{
		SecurityDefinitions: specAnalyser.d.Spec().SecurityDefinitions,
		GlobalSecurity:      specAnalyser.d.Spec().Security,
	}
}

// GetAllHeaderParameters gets all the parameters of type headers present in the swagger file and returns the header
// configurations. Currently only the following parameters are supported:
// - root level parameters (not supported)
// - path level parameters (not supported)
// - operation level parameters (supported)
func (specAnalyser *specV2Analyser) GetAllHeaderParameters() SpecHeaderParameters {
	return getAllHeaderParameters(specAnalyser.d.Spec().Paths.Paths)
}

func (specAnalyser *specV2Analyser) GetAPIBackendConfiguration() (SpecBackendConfiguration, error) {
	return newOpenAPIBackendConfigurationV2(specAnalyser.d.Spec(), specAnalyser.openAPIDocumentURL)
}

// isEndPointFullyTerraformResourceCompliant returns true only if:
// - The path given 'resourcePath' is an instance path (e,g: "/users/{username}")
// - The path given has GET operation defined (required). PUT and DELETE are optional
// - The root path for the given path 'resourcePath' is found (e,g: "/users")
// - The root path for the given path 'resourcePath' has mandatory POST operation defined
// - The root path POST operation for the given path 'resourcePath' has a parameter of type 'body' with a schema property referencing to an existing definition object or defining the schema inline (in which case the properties must be all input either required or optional properties)
// - The root path POST operation request payload definition and the response schema definitions may be the same (eg: they share the same definition model that declares both inputs like required/optional properties and the outputs as readOnly properties).
// - The root path POST operation request payload definition and the response schema definitions may be different (eg: the request schema contains only the inputs as required/optional properties and the response schema contains the inputs in the form of readOnly properties plus any other property that might be auth-generated by the API also configured as readOnly).
// - The path given GET operation schema must match the root path POST response schema
// - The resource schema definition must contain a field that uniquely identifies the resource or have a field with the 'x-terraform-id' extension set to true
// For more info about the requirements: https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#terraform-compliant-resource-requirements
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
func (specAnalyser *specV2Analyser) isEndPointFullyTerraformResourceCompliant(resourcePath string) (string, *spec.PathItem, *spec.Schema, error) {
	log.Printf("[DEBUG] validating end point terraform compatibility %s", resourcePath)
	err := specAnalyser.validateInstancePath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}
	resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, err := specAnalyser.validateRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}
	err = specAnalyser.validateResourceSchemaDefinition(resourceRootPostSchemaDef)
	if err != nil {
		return "", nil, nil, err
	}
	return resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, nil
}

func (specAnalyser *specV2Analyser) isEndPointTerraformDataSourceCompliant(path spec.PathItem) (*spec.Schema, error) {
	if path.Get == nil {
		return nil, errors.New("missing get operation")
	}
	if path.Get.Responses != nil {
		response, responseStatusOK := path.Get.Responses.ResponsesProps.StatusCodeResponses[http.StatusOK]
		if !responseStatusOK {
			return nil, errors.New("missing get 200 OK response specification")
		}
		if response.Schema == nil {
			return nil, errors.New("missing response schema")
		}
		if len(response.Schema.Type) > 0 && !response.Schema.Type.Contains("array") {
			return nil, errors.New("response does not return an array of items")
		}
		if response.Schema.Items == nil || response.Schema.Items.Schema == nil || !response.Schema.Items.Schema.Type.Contains("object") || len(response.Schema.Items.Schema.Properties) == 0 {
			return nil, errors.New("the response schema is missing the items schema specification or the items schema is not properly defined as object with properties configured")
		}
		return response.Schema.Items.Schema, nil
	}
	return nil, errors.New("missing get responses")
}

func (specAnalyser *specV2Analyser) validateInstancePath(path string) error {
	isResourceInstance, err := specAnalyser.isResourceInstanceEndPoint(path)
	if err != nil {
		return fmt.Errorf("error occurred while checking if path '%s' is a resource instance path", path)
	}
	if !isResourceInstance {
		return fmt.Errorf("path '%s' is not a resource instance path", path)
	}
	endPoint := specAnalyser.d.Spec().Paths.Paths[path]
	if endPoint.Get == nil {
		return fmt.Errorf("resource instance path '%s' missing required GET operation", path)
	}
	return nil
}

func (specAnalyser *specV2Analyser) validateRootPath(resourcePath string) (string, *spec.PathItem, *spec.Schema, error) {
	resourceRootPath, err := specAnalyser.findMatchingResourceRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}

	postExist := specAnalyser.postDefined(resourceRootPath)
	if !postExist {
		return "", nil, nil, fmt.Errorf("resource root path '%s' missing required POST operation", resourceRootPath)
	}

	resourceRootPathItem, _ := specAnalyser.d.Spec().Paths.Paths[resourceRootPath]
	resourceRootPostOperation := resourceRootPathItem.Post

	resourceRootPostRequestSchemaDef, err := specAnalyser.getBodyParameterBodySchema(resourceRootPostOperation)
	if err != nil {
		bodyParam := specAnalyser.bodyParameterExists(resourceRootPostOperation)
		// Use case where resource does not expect any input as part of the POST root operation, and only produces computed properties
		if bodyParam == nil {
			resourceSchema, err := specAnalyser.getSuccessfulResponseDefinition(resourceRootPostOperation)
			if err != nil {
				return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation (without body parameter) error: %s", resourceRootPath, err)
			}
			err = specAnalyser.validateResourceSchemaDefWithOptions(resourceSchema, true)
			if err != nil {
				return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation (without body parameter) validation error: %s", resourceRootPath, err)
			}
			return resourceRootPath, &resourceRootPathItem, resourceSchema, nil
		}
		return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation validation error: %s", resourceRootPath, err)
	}

	resourceRootPostResponseSchemaDef, err := specAnalyser.getSuccessfulResponseDefinition(resourceRootPostOperation)
	if err != nil {
		log.Printf("[DEBUG] failed to get the resource '%s' root path POST successful response configuration: %s", resourceRootPath, err)
		return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation is missing a successful response definition: %s", resourceRootPath, err)
	}

	if specAnalyser.schemaIsEqual(resourceRootPostRequestSchemaDef, resourceRootPostResponseSchemaDef) {
		log.Printf("[DEBUG] resource '%s' root path POST's req and resp schema definitions are the same", resourceRootPath)
		return resourceRootPath, &resourceRootPathItem, resourceRootPostRequestSchemaDef, nil
	}

	// Use case where resource POST's request payload model is different than the response payload (eg: request payload does not contain the id property (or any computed properties) and the response payload contains the inputs (as computed props already) and any other computed property that might be returned by the POST operation
	log.Printf("[DEBUG] resource '%s' root path POST's req and resp schemas not matching, checking if request schema is contained in the response schema and attemping to merge into one schema containing both the request and response schemas that contain both the required/optional inputs as well as all the computed properties", resourceRootPath)
	// if response payload contains the request properties but readOnly then that's a valid use case too
	mergedPostReqAndRespPayloadSchemas, err := specAnalyser.mergeRequestAndResponseSchemas(resourceRootPostRequestSchemaDef, resourceRootPostResponseSchemaDef)
	if err != nil {
		log.Printf("[DEBUG] failed to merge resource '%s' root path POST request and response schemas: %s", resourceRootPath, err)
		return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation does not meet any of the supported use cases", resourceRootPath)
	}
	log.Printf("[INFO] resource '%s' root path POST's req and resp merged into one: %+v", resourceRootPath, mergedPostReqAndRespPayloadSchemas)
	return resourceRootPath, &resourceRootPathItem, mergedPostReqAndRespPayloadSchemas, nil
}

// mergeRequestAndResponseSchemas attempts to merge the request schema and response schema and validates whether they compliant
// with the following specification:
// - the request schema must contain only properties that are either required or optional, not readOnly. If there are readOnly properties in the request schema, they will be ignored and not considered in the final schema.
// - the response schema must contain only properties that are readOnly, if they are not they will be converted automatically as readOnly in the final schema
// - the response schema must contain all the properties from the request schema but configured as readOnly
// If the above requirements are met, the resulted merged schema will be configured as follows:
// - All the properties from the request schema will be kept as is and integrated in the merged schema
// - All the properties from the response schema will be kept as is and integrated in the merged schema. The properties that are also in the request schema will not be integrated as the configuraiton in the request schema will be kept for those
// - The resulted merged properties will contain the extensions from both the request and response schema properties. However, if there
// is a matching property on both schema's but with a different value, the value in the response schema extension would take preference and will be the one kept in the final merged schema
func (specAnalyser *specV2Analyser) mergeRequestAndResponseSchemas(requestSchema *spec.Schema, responseSchema *spec.Schema) (*spec.Schema, error) {
	if requestSchema == nil {
		return nil, fmt.Errorf("resource missing request schema")
	}
	if responseSchema == nil {
		return nil, fmt.Errorf("resource missing response schema")
	}
	responseSchemaProps := responseSchema.Properties
	requestSchemaProps := requestSchema.Properties
	// responseSchema must contain at least 1 more property than requestSchema since it is expected that responseSchema will have the id readOnly property
	if len(responseSchemaProps) < len(requestSchemaProps) {
		return nil, fmt.Errorf("resource response schema contains less properties than the request schema, response schema must contain the request schema properties to be able to merge both schemas")
	}

	// Init merged schema to empty
	mergedSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Properties: map[string]spec.Schema{},
		},
	}

	// Copy response schema props into the merge schema. This avoids potential issues with pointers where when overriding
	// the response schema with the requests it would override the original response schema property too
	for responsePropName, responseProp := range responseSchemaProps {
		if !responseProp.ReadOnly {
			log.Printf("[WARN] resource's response schema property '%s' must be readOnly as response properties are considered computed (returned by the API). Therefore, the provider will automatically convert it to readOnly in the final resource schema", responsePropName)
			responseProp.ReadOnly = true
		}
		mergedSchema.Properties[responsePropName] = responseProp
	}
	// Ensure only the request's required properties are kept as required too in the final merged schema
	mergedSchema.Required = requestSchema.Required
	for requestSchemaPropName, requestSchemaProp := range requestSchemaProps {

		//
		// Ignoring the request property if it's readOnly. If the property is to be returned by the POST response it should be
		// defined as part of the response schema and therefore it will be added to the final schema accordingly.
		// This decision is made to ensure compliance with OpenAPI spec 2.0 but instead of failing is gracefully handling the 'badly' documented document.
		// More info on readOnly here: https://swagger.io/specification/v2/#fixed-fields-13 (readOnly section)
		// A "read only" property means that it MAY be sent as part of a response but MUST NOT be sent as part of the request.
		// Properties marked as readOnly being true SHOULD NOT be in the required list of the defined schema.
		if requestSchemaProp.ReadOnly {
			continue
		}

		_, exists := responseSchemaProps[requestSchemaPropName]
		if !exists {
			return nil, fmt.Errorf("resource's request schema property '%s' not contained in the response schema", requestSchemaPropName)
		}

		// Override response property with request property so the property input configuration is kept as is
		mergedSchema.Properties[requestSchemaPropName] = requestSchemaProp

		// Ensure the extensions from both the request and response schemas are kept.
		// If the same extension is present in both the request and response but with different values, the extension value in the response schema takes preference
		for extensionName, extensionValue := range responseSchemaProps[requestSchemaPropName].Extensions {
			mergedProp := mergedSchema.Properties[requestSchemaPropName]
			if mergedProp.Extensions == nil {
				mergedProp.Extensions = map[string]interface{}{}
				mergedSchema.Properties[requestSchemaPropName] = mergedProp
			}
			mergedProp.Extensions[extensionName] = extensionValue
		}
	}
	return mergedSchema, nil
}

func (specAnalyser *specV2Analyser) schemaIsEqual(requestSchema *spec.Schema, responseSchema *spec.Schema) bool {
	if requestSchema == responseSchema {
		return true
	}
	if requestSchema != nil && responseSchema != nil {
		requestSchemaJSON, err := requestSchema.MarshalJSON()
		if err == nil {
			responseSchemaJSON, err := responseSchema.MarshalJSON()
			if err == nil {
				if string(requestSchemaJSON) == string(responseSchemaJSON) {
					return true
				}
			}
		}
	}
	return false
}

// getSuccessfulResponseDefinition is responsible for getting the model definition from the response that matches a successful
// response (either 200, 201 or 202 whichever is found first). It is assumed that the the responses will only include one of the
// aforementioned successful responses, if multiple are present the first one found will be selected and its corresponding schema
// will be returned
func (specAnalyser *specV2Analyser) getSuccessfulResponseDefinition(operation *spec.Operation) (*spec.Schema, error) {
	if operation == nil || operation.Responses == nil {
		return nil, fmt.Errorf("operation is missing responses")
	}
	for responseStatusCode, response := range operation.Responses.ResponsesProps.StatusCodeResponses {
		if responseStatusCode == http.StatusOK || responseStatusCode == http.StatusCreated || responseStatusCode == http.StatusAccepted {
			if response.Schema == nil {
				return nil, fmt.Errorf("operation response '%d' is missing the schema definition", responseStatusCode)
			}
			return response.Schema, nil
		}
	}
	return nil, fmt.Errorf("operation is missing successful response")
}

func (specAnalyser *specV2Analyser) validateResourceSchemaDefWithOptions(schema *spec.Schema, shouldPropBeReadOnly bool) error {
	containsIdentifier := false
	for propertyName, property := range schema.Properties {
		if propertyName == "id" {
			containsIdentifier = true
		} else if exists, useAsIdentifier := property.Extensions.GetBool(extTfID); exists && useAsIdentifier {
			containsIdentifier = true
		}
		if shouldPropBeReadOnly {
			if property.ReadOnly == false {
				return fmt.Errorf("resource schema contains properties that are not just read only")
			}
		}
	}
	if containsIdentifier == false {
		return fmt.Errorf("resource schema is missing a property that uniquely identifies the resource, either a property named 'id' or a property with the extension '%s' set to true", extTfID)
	}
	return nil
}

func (specAnalyser *specV2Analyser) validateResourceSchemaDefinition(schema *spec.Schema) error {
	return specAnalyser.validateResourceSchemaDefWithOptions(schema, false)
}

// postIsPresent checks if the given resource has a POST implementation returning true if the path is found
// in paths and the path exposes a POST operation
func (specAnalyser *specV2Analyser) postDefined(resourceRootPath string) bool {
	b, exists := specAnalyser.d.Spec().Paths.Paths[resourceRootPath]
	if !exists || b.Post == nil {
		return false
	}
	return true
}

func (specAnalyser *specV2Analyser) bodyParameterExists(resourceRootPostOperation *spec.Operation) *spec.Parameter {
	if resourceRootPostOperation == nil {
		return nil
	}
	for _, parameter := range resourceRootPostOperation.Parameters {
		if parameter.In == "body" {
			return &parameter
		}
	}
	return nil
}

func (specAnalyser *specV2Analyser) getBodyParameterBodySchema(resourceRootPostOperation *spec.Operation) (*spec.Schema, error) {
	bodyParameter := specAnalyser.bodyParameterExists(resourceRootPostOperation)
	if bodyParameter == nil {
		return nil, fmt.Errorf("resource root operation missing body parameter")
	}

	if bodyParameter.Schema == nil {
		return nil, fmt.Errorf("resource root operation missing the schema for the POST operation body parameter")
	}

	if bodyParameter.Schema.Ref.String() != "" {
		return nil, fmt.Errorf("the operation ref was not expanded properly, check that the ref is valid (no cycles, bogus, etc)")
	}

	if len(bodyParameter.Schema.Properties) > 0 {
		return bodyParameter.Schema, nil
	}
	return nil, fmt.Errorf("POST operation contains an schema with no properties")
}

// isResourceInstanceEndPoint checks if the given path is of form /resource/{id}
func (specAnalyser *specV2Analyser) isResourceInstanceEndPoint(p string) (bool, error) {
	r, _ := regexp.Compile("^.*{.+}[\\/]?$")
	return r.MatchString(p), nil
}

// findMatchingResourceRootPath returns the corresponding POST root and path for a given end point
// Example: Given 'resourcePath' being "/users/{username}" the result could be "/users" or "/users/" depending on
// how the POST operation (resourceRootPath) of the given resource is defined in swagger.
// If there is no match the returned string will be empty
func (specAnalyser *specV2Analyser) findMatchingResourceRootPath(resourceInstancePath string) (string, error) {
	r, _ := regexp.Compile(resourceInstanceRegex)
	result := r.FindStringSubmatch(resourceInstancePath)
	log.Printf("[DEBUG] resource '%s' root path match: %s", resourceInstancePath, result)
	if len(result) != 2 {
		return "", fmt.Errorf("resource instance path '%s' missing valid resource root path, more than two results returned from match '%s'", resourceInstancePath, result)
	}

	resourceRootPath := result[1] // e,g: /v1/cdns/{id} /v1/cdns/

	if _, exists := specAnalyser.d.Spec().Paths.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path with trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	// Handles the case where the swagger file root path does not have a trailing slash in the path
	resourceRootPath = strings.TrimRight(resourceRootPath, "/")
	if _, exists := specAnalyser.d.Spec().Paths.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path without trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	return "", fmt.Errorf("resource instance path '%s' missing resource root path", resourceInstancePath)
}
