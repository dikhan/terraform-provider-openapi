package openapi

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestV3_mergeRequestAndResponseSchemas(t *testing.T) {
	testCases := []struct {
		name                 string
		requestSchema        *openapi3.Schema
		responseSchema       *openapi3.Schema
		expectedMergedSchema *openapi3.Schema
		expectedError        string
	}{
		{
			name:                 "request schema is nil",
			requestSchema:        nil,
			responseSchema:       &openapi3.Schema{},
			expectedMergedSchema: nil,
			expectedError:        "resource missing request schema",
		},
		{
			name:                 "response schema is nil",
			requestSchema:        &openapi3.Schema{},
			responseSchema:       nil,
			expectedMergedSchema: nil,
			expectedError:        "resource missing response schema",
		},
		{
			name: "request schema contains more properties than response schema, this is not valid as response should always contain the request properties plus any other computed that is computed",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"optional_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedMergedSchema: nil,
			expectedError:        "resource response schema contains less properties than the request schema, response schema must contain the request schema properties to be able to merge both schemas",
		},
		{
			name: "response schema is missing request schema properties",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedMergedSchema: nil,
			expectedError:        "resource's request schema property 'required_prop' not contained in the response schema",
		},
		{
			name: "request properties contain readOnly properties and the response schema contains the request input properties (required/optional) as well as any other computed property",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
					"some_computed_property": { // readOnly props from the request schema are not considered in the final merged schema
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"some_computed_property": { // since the response schema also contains the some_computed_property it will be included in the final merged schema
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"some_computed_property": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "response contains properties that are not readOnly and the provide will automatically configure them as readOnly in the final merged schema",
			requestSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"some_property": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"some_property": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"some_computed_property": {
						Value: &openapi3.Schema{
							Type: "string",
							// Not readOnly although it should
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"some_property": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
					"some_computed_property": { // The merged schema converted automatically the response property as readOnly
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "request and response schemas are merged successfully, request's required properties are kept as is as well as the optional properties and any other response's computed property is merged into the final schema",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
					"optional_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"id", "optional_prop", "required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"optional_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"optional_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "request and response schemas are merged successfully, extensions in the response schema are kept as is",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"identifier_property": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfID: true,
								},
							},
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"identifier_property": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfID: true,
								},
							},
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "request and response schemas are merged successfully, extensions in the request schema are kept as is",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"id", "required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "request and response schemas are merged successfully, extensions in the request schema is nil and the response does have an extension",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "request and response schemas are merged successfully, extensions in the response schema is nil and the request does have an extension",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "request and response schemas are merged successfully, response schema extensions take preference when both the request and response have the same extension in a property and with different values",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_request_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"id", "required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_response_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									extTfFieldName: "required_response_preferred_name_prop",
								},
							},
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "request and response schemas are merged successfully, final merged schema only keeps in the required list the required properties in the request schema",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"id"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			expectedMergedSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		specV3Analyser := specV3Analyser{}
		mergedSchema, err := specV3Analyser.mergeRequestAndResponseSchemas(tc.requestSchema, tc.responseSchema)
		if tc.expectedError != "" {
			assert.Equal(t, tc.expectedError, err.Error(), tc.name)
		} else {
			assert.Equal(t, tc.expectedMergedSchema, mergedSchema, tc.name)
		}
	}
}

func TestV3_schemaIsEqual(t *testing.T) {
	testSchema := &openapi3.Schema{}
	testCases := []struct {
		name           string
		requestSchema  *openapi3.Schema
		responseSchema *openapi3.Schema
		expectedOutput bool
	}{
		{
			name:           "request schema and response schema are equal (empty schemas)",
			requestSchema:  &openapi3.Schema{},
			responseSchema: &openapi3.Schema{},
			expectedOutput: true,
		},
		{
			name:           "request schema and response schema are equal (same pointer)",
			requestSchema:  testSchema,
			responseSchema: testSchema,
			expectedOutput: true,
		},
		{
			name: "request schema and response schema are equal",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedOutput: true,
		},
		{
			name: "request schema and response schema are equal (though the properties are not in the same order)",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"required_prop": { // changing order here on purpose to see if it makes any difference
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedOutput: true,
		},
		{
			name: "request schema and response schema are NOT equal (request schema contains required props while response schema does not)",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedOutput: false,
		},
		{
			name: "request schema and response schema are NOT equal (they are completely different)",
			requestSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"some_property": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Properties: map[string]*openapi3.SchemaRef{
					"some_other_property": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedOutput: false,
		},
		{
			name: "request schema and response schema are NOT equal (request schema contains properties with extensions and response schema does not)",
			requestSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									"x-terraform-field-name": "required_prop_preferred_name",
								},
							},
						},
					},
				},
			},
			responseSchema: &openapi3.Schema{
				Required: []string{"required_prop"},
				Properties: map[string]*openapi3.SchemaRef{
					"id": {
						Value: &openapi3.Schema{
							Type:     "string",
							ReadOnly: true,
						},
					},
					"required_prop": {
						Value: &openapi3.Schema{
							Type: "string",
						},
					},
				},
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		specV3Analyser := specV3Analyser{}
		isEqual := specV3Analyser.schemaIsEqual(tc.requestSchema, tc.responseSchema)
		assert.Equal(t, tc.expectedOutput, isEqual, tc.name)
	}
}
