package openapi_terraform_docs_generator

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestProviderResources_RenderZendesk(t *testing.T) {
	r := ProviderResources{
		ProviderName: "openapi",
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
	expectedHTML := `<h2 id="provider_resources">Provider Resources</h2>


    <h3 id="cdn" dir="ltr">openapi_cdn</h3><p>The &#39;cdn&#39; allows you to manage &#39;cdn&#39; resources using Terraform.</p>
    <h4 id="resource_cdn_example_usage" dir="ltr">Example usage</h4>
<pre>
<span>resource </span><span>"openapi_cdn" "my_cdn"</span>{
    <span>string_prop  </span>= <span>"string_prop"</span>
    <span>integer_prop  </span>= <span>1234</span>
    <span>float_prop  </span>= <span>12.95</span>
    <span>bool_prop  </span>= <span>true</span>
    
    <span>list_integer_prop  </span>= <span>[1234, 4567]</span>
    <span>list_boolean_prop  </span>= <span>[true, false]</span>
    <span>list_float_prop  </span>= <span>[12.36, 99.45]</span>
    <span>object_prop  </span><span>{</span>
                
    <span>objectPropertyRequired  </span>= <span>"objectPropertyRequired"</span>
                
            <span>}</span>
    <span>list_object_prop  </span><span>{</span>
                
    <span>objectPropertyRequired  </span>= <span>"objectPropertyRequired"</span>
                
            <span>}</span>
<span>}</span>
</pre>
<h4 id="resource_cdn_arguments_reference" dir="ltr">Arguments Reference</h4>
<p dir="ltr">The following arguments are supported:</p>
<ul dir="ltr"><li> string_prop [string]  - (Required) string property description</li>
	<li> integer_prop [integer]  - (Required) integer property description</li>
	<li> float_prop [number]  - (Required) float property description</li>
	<li> bool_prop [boolean]  - (Required) boolean property description</li>
	<li> list_string_prop []  - (Required) list_string_prop property description</li>
	<li> list_integer_prop [list of integers]  - (Required) list_integer_prop property description</li>
	<li> list_boolean_prop [list of booleans]  - (Required) list_boolean_prop property description</li>
	<li> list_float_prop [list of numbers]  - (Required) list_float_prop property description</li>
	<li><span class="wysiwyg-color-red">*</span> object_prop [object]  - (Required) this is an object property. The following properties compose the object schema
        :<ul dir="ltr"><li> objectPropertyRequired [string]  - (Required) </li>
	
        </ul>
        </li>
	<li> list_object_prop [list of objects]  - (Required) this is an object property. The following properties compose the object schema
        :<ul dir="ltr"><li> objectPropertyRequired [string]  - (Required) </li>
	
        </ul>
        </li>
	<li> optional_computed_prop [string]  - (Optional) this is an optional computed property</li>
	<li> optional_prop [string]  - (Optional) this is an optional computed property</li>
	
    </ul>
        
    
        
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn.my_cdn.<b>computed_object_prop[0]</b>.object_property)</p>
    <p><span class="wysiwyg-color-red">*Note: Sample note</span></p>


<h4 id="resource_cdn_attributes_reference" dir="ltr">Attributes Reference</h4>
<p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
<ul dir="ltr"><li><span class="wysiwyg-color-red">*</span> object_prop [object]  - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string]  - </li>
		
            </ul>
            </li>
		<li> list_object_prop [list of objects]  - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string]  - </li>
		
            </ul>
            </li>
		<li> computed_string_prop [string]  - string property description</li>
		<li> computed_integer_prop [integer]  - integer property description</li>
		<li> computed_float_prop [number]  - float property description</li>
		<li> computed_bool_prop [boolean]  - boolean property description</li>
		<li> computed_sensitive_prop [string] (<a href="#special_terms_definitions_sensitive_property" target="_self">sensitive</a>) - this is sensitive property</li>
		<li> computed_list_string_prop []  - list_string_prop property description</li>
		<li> computed_list_integer_prop [list of integers]  - list_integer_prop property description</li>
		<li> computed_list_boolean_prop [list of booleans]  - list_boolean_prop property description</li>
		<li> computed_list_float_prop [list of numbers]  - list_float_prop property description</li>
		<li><span class="wysiwyg-color-red">*</span> computed_object_prop [object]  - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string]  - </li>
		
            </ul>
            </li>
		<li> computed_list_object_prop [list of objects]  - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string]  - </li>
		
            </ul>
            </li>
		
    </ul>
        
    
        
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn.my_cdn.<b>computed_object_prop[0]</b>.object_property)</p><h4 id="resource_cdn_import" dir="ltr">Import</h4>
<p dir="ltr">
    cdn resources can be imported using the&nbsp;<code>id</code> . This is a sub-resource so the parent resource IDs (<code>[parent_id]</code>) are required to be able to retrieve an instance of this resource, e.g:
</p>
<pre dir="ltr">$ terraform import cdn.my_cdn parent_id/cdn_id</pre>
<p dir="ltr">
    <strong>Note</strong>: In order for the import to work, the 'openapi' terraform
    provider must be&nbsp;<a href="#provider_installation" target="_self">properly installed</a>. Read more about Terraform import usage&nbsp;<a href="https://www.terraform.io/docs/import/usage.html" target="_blank" rel="noopener noreferrer">here</a>.
</p>

 `
	err := r.Render(&buf, ProviderResourcesTmpl)
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
	assert.Nil(t, err)
}

func TestContainsResourcesWithSecretProperties(t *testing.T) {
	testCases := []struct {
		name                    string
		providerResources       ProviderResources
		expectedContainsSecrets bool
	}{
		{
			name: "no resources with sensitive props",
			providerResources: ProviderResources{
				Resources: []Resource{
					{Properties: []Property{{IsSensitive: false}}},
				},
			},
			expectedContainsSecrets: false,
		},
		{
			name: "resource has sensitive prop",
			providerResources: ProviderResources{
				Resources: []Resource{
					{Properties: []Property{{IsSensitive: true}}},
				},
			},
			expectedContainsSecrets: true,
		},
	}

	for _, tc := range testCases {
		actualContainsSecrets := tc.providerResources.ContainsResourcesWithSecretProperties()
		assert.Equal(t, tc.expectedContainsSecrets, actualContainsSecrets, tc.name)
	}
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
