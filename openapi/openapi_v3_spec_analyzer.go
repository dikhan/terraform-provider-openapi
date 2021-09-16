package openapi

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// specV3Analyser defines an SpecAnalyser implementation for OpenAPI v3 specification
// Forcing creation of this object via constructor so proper input validation is performed before creating the struct
// instance
type specV3Analyser struct {
	openAPIDocumentURL string
	d                  *openapi3.T
}

var _ SpecAnalyser = (*specV3Analyser)(nil)

// newSpecAnalyserV3 creates an instance of specV2Analyser which implements the SpecAnalyser interface
// This implementation provides an analyser that understands an OpenAPI v2 document
func newSpecAnalyserV3(openAPIDocumentFilename string) (*specV3Analyser, error) {
	if openAPIDocumentFilename == "" {
		return nil, errors.New("open api document filename argument empty, please provide the url of the OpenAPI document")
	}
	openAPIDocumentURL, err := url.Parse(openAPIDocumentFilename)
	if err != nil {
		return nil, fmt.Errorf("invalid URL to retrieve OpenAPI document: '%s' - error = %s", openAPIDocumentFilename, err)
	}
	apiSpec, err := openapi3.NewLoader().LoadFromURI(openAPIDocumentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the OpenAPI document from '%s' - error = %s", openAPIDocumentFilename, err)
	}
	return &specV3Analyser{
		d:                  apiSpec,
		openAPIDocumentURL: openAPIDocumentFilename,
	}, nil
}

func (specAnalyser *specV3Analyser) GetTerraformCompliantResources() ([]SpecResource, error) {
	var resources []SpecResource
	start := time.Now()
	for resourcePath, pathItem := range specAnalyser.d.Paths {
		resourceRootPath, resourceRoot, resourcePayloadSchemaDef, err := specAnalyser.isEndPointFullyTerraformResourceCompliant(resourcePath)
		if err != nil {
			log.Printf("[DEBUG] resource path '%s' not terraform compliant: %s", resourcePath, err)
			continue
		}

		// TODO: add multiregion support

		// TODO: add support for other components besides Components.Schemas
		r, err := newSpecV3Resource(resourceRootPath, resourcePayloadSchemaDef, resourceRoot, pathItem, specAnalyser.d.Components.Schemas, specAnalyser.d.Paths)
		if err != nil {
			log.Printf("[WARN] ignoring resource '%s' due to an error while creating a creating the SpecV3Resource: %s", resourceRootPath, err)
			continue
		}

		// TODO: add subresource support

		log.Printf("[INFO] found terraform compliant resource [name='%s', rootPath='%s', instancePath='%s']", r.GetResourceName(), resourceRootPath, resourcePath)
		resources = append(resources, r)
	}
	log.Printf("[INFO] found %d terraform compliant resources (time: %s)", len(resources), time.Since(start))
	return resources, nil
}

func (specAnalyser specV3Analyser) GetTerraformCompliantDataSources() []SpecResource {
	return []SpecResource{}
}

func (specAnalyser specV3Analyser) GetSecurity() SpecSecurity {
	// TODO: replace this stub
	return &specSecurityStub{
		securityDefinitions: &SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition("apikey_auth", "Authorization"),
		},
		globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
	}
}

func (specAnalyser specV3Analyser) GetAllHeaderParameters() SpecHeaderParameters {
	// TODO: add support for header params
	return []SpecHeaderParam{
		{
			Name:          "X-Request-ID",
			TerraformName: "x_request_id",
			IsRequired:    true,
		},
	}
}

func (specAnalyser specV3Analyser) GetAPIBackendConfiguration() (SpecBackendConfiguration, error) {
	return newOpenAPIBackendConfigurationV3(specAnalyser.d, specAnalyser.openAPIDocumentURL)
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
func (specAnalyser *specV3Analyser) isEndPointFullyTerraformResourceCompliant(resourcePath string) (string, *openapi3.PathItem, *openapi3.Schema, error) {
	log.Printf("[DEBUG] validating end point terraform compatibility %s", resourcePath)
	// TODO: implement this
	//err := specAnalyser.validateInstancePath(resourcePath)
	//if err != nil {
	//	return "", nil, nil, err
	//}
	resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, err := specAnalyser.validateRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}
	// TODO: implement this
	//err = specAnalyser.validateResourceSchemaDefinition(resourceRootPostSchemaDef)
	//if err != nil {
	//	return "", nil, nil, err
	//}
	return resourceRootPath, resourceRootPathItem, resourceRootPostSchemaDef, nil
}

func (specAnalyser *specV3Analyser) validateRootPath(resourcePath string) (string, *openapi3.PathItem, *openapi3.Schema, error) {
	resourceRootPath, err := specAnalyser.findMatchingResourceRootPath(resourcePath)
	if err != nil {
		return "", nil, nil, err
	}

	postExist := specAnalyser.postDefined(resourceRootPath)
	if !postExist {
		return "", nil, nil, fmt.Errorf("resource root path '%s' missing required POST operation", resourceRootPath)
	}

	resourceRootPathItem, _ := specAnalyser.d.Paths[resourceRootPath]
	resourceRootPostOperation := resourceRootPathItem.Post

	resourceRootPostRequestSchemaDef, err := specAnalyser.getBodyParameterBodySchema(resourceRootPostOperation)
	if err != nil {
		// TODO: Use case where resource does not expect any input as part of the POST root operation, and only produces computed properties
		return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation validation error: %s", resourceRootPath, err)
	}

	resourceRootPostResponseSchemaDef, err := specAnalyser.getSuccessfulResponseDefinition(resourceRootPostOperation)
	if err != nil {
		log.Printf("[DEBUG] failed to get the resource '%s' root path POST successful response configuration: %s", resourceRootPath, err)
		return "", nil, nil, fmt.Errorf("resource root path '%s' POST operation is missing a successful response definition: %s", resourceRootPath, err)
	}

	if specAnalyser.schemaIsEqual(resourceRootPostRequestSchemaDef, resourceRootPostResponseSchemaDef) {
		log.Printf("[DEBUG] resource '%s' root path POST's req and resp schema definitions are the same", resourceRootPath)
		return resourceRootPath, resourceRootPathItem, resourceRootPostRequestSchemaDef, nil
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
	return resourceRootPath, resourceRootPathItem, mergedPostReqAndRespPayloadSchemas, nil
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
func (specAnalyser *specV3Analyser) mergeRequestAndResponseSchemas(requestSchema *openapi3.Schema, responseSchema *openapi3.Schema) (*openapi3.Schema, error) {
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
	mergedSchema := &openapi3.Schema{
		Properties: map[string]*openapi3.SchemaRef{},
	}

	// Copy response schema props into the merge schema. This avoids potential issues with pointers where when overriding
	// the response schema with the requests it would override the original response schema property too
	for responsePropName, responseProp := range responseSchemaProps {
		// TODO: support property $ref
		if !responseProp.Value.ReadOnly {
			log.Printf("[WARN] resource's response schema property '%s' must be readOnly as response properties are considered computed (returned by the API). Therefore, the provider will automatically convert it to readOnly in the final resource schema", responsePropName)
			responseProp.Value.ReadOnly = true
		}
		// TODO: ensure we support property $ref here
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
		// TODO: support property $ref
		if requestSchemaProp.Value.ReadOnly {
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
		// TODO: support $ref
		for extensionName, extensionValue := range responseSchemaProps[requestSchemaPropName].Value.Extensions {
			mergedProp := mergedSchema.Properties[requestSchemaPropName]
			// TODO: support $ref
			if mergedProp.Value.Extensions == nil {
				mergedProp.Value.Extensions = map[string]interface{}{}
				mergedSchema.Properties[requestSchemaPropName] = mergedProp
			}
			// TODO: support property $ref
			mergedProp.Value.Extensions[extensionName] = extensionValue
		}
	}
	return mergedSchema, nil
}

func (specAnalyser *specV3Analyser) getBodyParameterBodySchema(resourceRootPostOperation *openapi3.Operation) (*openapi3.Schema, error) {
	bodyParameter := specAnalyser.bodyParameterExists(resourceRootPostOperation)
	if bodyParameter == nil {
		return nil, fmt.Errorf("resource root operation missing body parameter")
	}

	if bodyParameter.Schema == nil {
		return nil, fmt.Errorf("resource root operation missing the schema for the POST operation body parameter")
	}

	// TODO: test that this schema ref worked the way I think (removed .String())
	if bodyParameter.Schema.Ref != "" {
		return nil, fmt.Errorf("the operation ref was not expanded properly, check that the ref is valid (no cycles, bogus, etc)")
	}

	// TODO: test that Schema.Ref was expanded to Schema.Value already
	if len(bodyParameter.Schema.Value.Properties) > 0 {
		return bodyParameter.Schema.Value, nil
	}
	return nil, fmt.Errorf("POST operation contains an schema with no properties")
}

// findMatchingResourceRootPath returns the corresponding POST root and path for a given end point
// Example: Given 'resourcePath' being "/users/{username}" the result could be "/users" or "/users/" depending on
// how the POST operation (resourceRootPath) of the given resource is defined in swagger.
// If there is no match the returned string will be empty
func (specAnalyser *specV3Analyser) findMatchingResourceRootPath(resourceInstancePath string) (string, error) {
	r, _ := regexp.Compile(resourceInstanceRegex)
	result := r.FindStringSubmatch(resourceInstancePath)
	log.Printf("[DEBUG] resource '%s' root path match: %s", resourceInstancePath, result)
	if len(result) != 2 {
		return "", fmt.Errorf("resource instance path '%s' missing valid resource root path, more than two results returned from match '%s'", resourceInstancePath, result)
	}

	resourceRootPath := result[1] // e,g: /v1/cdns/{id} /v1/cdns/

	if _, exists := specAnalyser.d.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path with trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	// Handles the case where the swagger file root path does not have a trailing slash in the path
	resourceRootPath = strings.TrimRight(resourceRootPath, "/")
	if _, exists := specAnalyser.d.Paths[resourceRootPath]; exists {
		log.Printf("[DEBUG] found resource root path without trailing '/' - %+s", resourceRootPath)
		return resourceRootPath, nil
	}

	return "", fmt.Errorf("resource instance path '%s' missing resource root path", resourceInstancePath)
}

func (specAnalyser *specV3Analyser) schemaIsEqual(requestSchema *openapi3.Schema, responseSchema *openapi3.Schema) bool {
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
func (specAnalyser *specV3Analyser) getSuccessfulResponseDefinition(operation *openapi3.Operation) (*openapi3.Schema, error) {
	if operation == nil || operation.Responses == nil {
		return nil, fmt.Errorf("operation is missing responses")
	}
	for responseStatusCode, response := range operation.Responses {
		if responseStatusCode == strconv.Itoa(http.StatusOK) || responseStatusCode == strconv.Itoa(http.StatusCreated) || responseStatusCode == strconv.Itoa(http.StatusAccepted) {
			// TODO: support response $ref
			if response.Value.Content == nil {
				return nil, fmt.Errorf("operation response '%d' is missing the schema definition", responseStatusCode)
			}
			// TODO: support response $ref, other content-types, schema $ref, etc
			return response.Value.Content.Get("application/json").Schema.Value, nil
		}
	}
	return nil, fmt.Errorf("operation is missing successful response")
}

// postIsPresent checks if the given resource has a POST implementation returning true if the path is found
// in paths and the path exposes a POST operation
func (specAnalyser *specV3Analyser) postDefined(resourceRootPath string) bool {
	b, exists := specAnalyser.d.Paths[resourceRootPath]
	if !exists || b.Post == nil {
		return false
	}
	return true
}

func (specAnalyser *specV3Analyser) bodyParameterExists(resourceRootPostOperation *openapi3.Operation) *openapi3.Parameter {
	if resourceRootPostOperation == nil {
		return nil
	}
	for _, parameter := range resourceRootPostOperation.Parameters {
		// TODO: support parameter $ref
		if parameter.Value.In == "body" {
			return parameter.Value
		}
	}
	return nil
}
