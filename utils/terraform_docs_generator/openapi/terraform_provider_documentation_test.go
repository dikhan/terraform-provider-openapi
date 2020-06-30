package openapi

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestProviderResources_RenderZendesk(t *testing.T) {
	r := ProviderResources{
		Resources: []Resource{
			{
				Name:        "cdn",
				Description: "The 'cdn' allows you to manage 'cdn' resources using Terraform.",
				Properties: []Property{
					// Arguments
					createProperty("string_prop", "string", "string property description", true, false),
					createProperty("integer_prop", "integer", "integer property description", true, false),
					createProperty("float_prop", "number", "float property description", true, false),
					createProperty("bool_prop", "boolean", "boolean property description", true, false),
					createArrayProperty("list_string_prop", "", "string", "list_string_prop property description", true, false),
					createArrayProperty("list_integer_prop", "list", "integer", "list_integer_prop property description", true, false),
					createArrayProperty("list_boolean_prop", "list", "boolean", "list_boolean_prop property description", true, false),
					createArrayProperty("list_float_prop", "list", "number", "list_float_prop property description", true, false),
					Property{Name: "object_prop", Type: "object", Description: "this is an object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
					Property{Name: "list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
					Property{Name: "optional_computed_prop", Type: "string", Description: "this is an optional computed property", IsOptionalComputed: true},
					Property{Name: "optional_prop", Type: "string", Description: "this is an optional computed property", Required: false},
					// Attributes
					createProperty("computed_string_prop", "string", "string property description", false, true),
					createProperty("computed_integer_prop", "integer", "integer property description", false, true),
					createProperty("computed_float_prop", "number", "float property description", false, true),
					createProperty("computed_bool_prop", "boolean", "boolean property description", false, true),
					Property{Name: "computed_sensitive_prop", Type: "string", Description: "this is sensitive property", Computed: true, IsSensitive: true},
					createArrayProperty("computed_list_string_prop", "", "string", "list_string_prop property description", false, true),
					createArrayProperty("computed_list_integer_prop", "list", "integer", "list_integer_prop property description", false, true),
					createArrayProperty("computed_list_boolean_prop", "list", "boolean", "list_boolean_prop property description", false, true),
					createArrayProperty("computed_list_float_prop", "list", "number", "list_float_prop property description", false, true),
					Property{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
					Property{Name: "computed_list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
				},
				ParentProperties: []string{"parent_id"},
				ArgumentsReference: ArgumentsReference{
					Notes: []string{"Sample note"},
				},
			},
		},
	}
	var buf bytes.Buffer
	err := r.RenderZendesk(&buf)
	fmt.Println(strings.Trim(buf.String(), "\n"))
	assert.Nil(t, err)
}

func TestDataSources_RenderZendesk(t *testing.T) {
	d := DataSources{
		DataSourceInstances: []DataSource{
			{
				Name:         "cdn_instance",
				OtherExample: "",
				Properties: []Property{
					createProperty("computed_string_prop", "string", "string property description", false, true),
					createProperty("computed_integer_prop", "integer", "integer property description", false, true),
					createProperty("computed_float_prop", "number", "float property description", false, true),
					createProperty("computed_bool_prop", "boolean", "boolean property description", false, true),
					Property{Name: "computed_sensitive_prop", Type: "string", Description: "this is sensitive property", Computed: true, IsSensitive: true},
					createArrayProperty("computed_list_string_prop", "", "string", "list_string_prop property description", false, true),
					createArrayProperty("computed_list_integer_prop", "list", "integer", "list_integer_prop property description", false, true),
					createArrayProperty("computed_list_boolean_prop", "list", "boolean", "list_boolean_prop property description", false, true),
					createArrayProperty("computed_list_float_prop", "list", "number", "list_float_prop property description", false, true),
					Property{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
					Property{Name: "computed_list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
				},
			},
		},
		DataSources: []DataSource{
			{
				Name:         "cdn",
				OtherExample: "",
				Properties: []Property{
					createProperty("computed_string_prop", "string", "string property description", false, true),
					createProperty("computed_integer_prop", "integer", "integer property description", false, true),
					createProperty("computed_float_prop", "number", "float property description", false, true),
					createProperty("computed_bool_prop", "boolean", "boolean property description", false, true),
					Property{Name: "computed_sensitive_prop", Type: "string", Description: "this is sensitive property", Computed: true, IsSensitive: true},
					createArrayProperty("computed_list_string_prop", "", "string", "list_string_prop property description", false, true),
					createArrayProperty("computed_list_integer_prop", "list", "integer", "list_integer_prop property description", false, true),
					createArrayProperty("computed_list_boolean_prop", "list", "boolean", "list_boolean_prop property description", false, true),
					createArrayProperty("computed_list_float_prop", "list", "number", "list_float_prop property description", false, true),
					Property{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
					Property{Name: "computed_list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
				},
			},
		},
	}
	var buf bytes.Buffer
	err := d.RenderZendesk(&buf)
	fmt.Println(strings.Trim(buf.String(), "\n"))
	assert.Nil(t, err)
}

func TestResource_BuildImportIDsExample(t *testing.T) {
	testCases := []struct {
		name              string
		parentProperties  []string
		expectedImportIDs string
	}{
		{
			name:              "resource configured with resource parent properties",
			parentProperties:  []string{"parent_id"},
			expectedImportIDs: "parent_id/fw_id",
		},
		{
			name:              "resource configured with NO resource parent properties",
			parentProperties:  nil,
			expectedImportIDs: "id",
		},
	}
	for _, tc := range testCases {
		resource := Resource{
			Name:             "fw",
			ParentProperties: tc.parentProperties,
		}
		result := resource.BuildImportIDsExample()
		assert.Equal(t, tc.expectedImportIDs, result)
	}
}

func TestProperty_ContainsComputedSubProperties(t *testing.T) {
	testCases := []struct {
		name           string
		property       Property
		expectedResult bool
	}{
		{
			name: "property does not have schema",
			property: Property{
				Name:   "some primitive property",
				Schema: nil,
			},
			expectedResult: false,
		},
		{
			name: "property does have a schema",
			property: Property{
				Name: "some property with schema (eg: object or array of objects) containing computed props",
				Schema: []Property{
					{
						Name:     "subProperty",
						Computed: true,
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "property does have a schema",
			property: Property{
				Name: "some property with schema (eg: object or array of objects) with no computed props",
				Schema: []Property{
					{
						Name:     "subProperty",
						Computed: false,
					},
				},
			},
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		result := tc.property.ContainsComputedSubProperties()
		assert.Equal(t, tc.expectedResult, result)
	}
}

func TestTerraformProviderDocumentation_RenderZendeskHTML(t *testing.T) {
	terraformProviderDocumentation := TerraformProviderDocumentation{
		ProviderName: "openapi",
		ProviderInstallation: ProviderInstallation{
			Example:      "➜ ~ This is an example",
			Other:        "Some more info about the installation",
			OtherCommand: "➜ ~ init_command do_something",
		},
		ProviderConfiguration: ProviderConfiguration{
			Regions: []string{"rst1"},
			ConfigProperties: []Property{
				{
					Name:     "token",
					Required: true,
					Type:     "string",
				},
			},
			ExampleUsage: nil,
			ArgumentsReference: ArgumentsReference{
				Notes: []string{"Note: some special notes..."},
			},
		},
		ShowSpecialTermsDefinitions: true,
		ProviderResources: ProviderResources{
			Resources: []Resource{
				{
					Name:        "cdn",
					Description: "The 'cdn' allows you to manage 'cdn' resources using Terraform.",
					Properties: []Property{
						createProperty("string_prop", "string", "string property description", true, false),
					},
					ParentProperties: []string{"parent_id"},
					ArgumentsReference: ArgumentsReference{
						Notes: []string{"Another note..."},
					},
				},
			},
		},
		DataSources: DataSources{
			DataSourceInstances: []DataSource{
				{
					Name:         "cdn_instance",
					OtherExample: "",
					Properties: []Property{
						createProperty("computed_string_prop", "string", "string property description", false, true),
					},
				},
			},
			DataSources: []DataSource{
				{
					Name:         "cdn",
					OtherExample: "",
					Properties: []Property{
						createProperty("computed_string_prop", "string", "string property description", false, true),
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	err := terraformProviderDocumentation.RenderZendeskHTML(&buf)
	fmt.Println(strings.Trim(buf.String(), "\n"))
	assert.Nil(t, err)
}

func createProperty(name, properType, description string, required, computed bool) Property {
	return Property{
		Name:        name,
		Required:    required,
		Computed:    computed,
		Type:        properType,
		Description: description,
	}
}

func createArrayProperty(name, properType, propItemsType, description string, required, computed bool) Property {
	return Property{
		Name:           name,
		Required:       required,
		Computed:       computed,
		Type:           properType,
		ArrayItemsType: propItemsType,
		Description:    description,
	}
}
