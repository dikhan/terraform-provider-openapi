package openapi_terraform_docs_generator

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestDataSources_RenderZendesk(t *testing.T) {
	d := DataSources{
		ProviderName: "openapi",
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
	expectedHTML := `<h2 id="provider_datasources">Data Sources (using resource id)</h2>


    <h3 id="cdn_instance" dir="ltr">openapi_cdn_instance</h3>
    <p>Retrieve an existing resource using it's ID</p>
    <h4 id="datasource_cdn_instance_example_usage" dir="ltr">Example usage</h4>
<pre><span>data </span><span>"openapi_cdn_instance" "my_cdn_instance"</span>{
    id = "existing_resource_id"
<span>}</span></pre>
    <h4 id="datasource_cdn_instance_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
        <li>id - (Required) ID of the existing resource to retrieve</li>
    </ul>
    <h4 id="datasource_cdn_instance_attributes_reference" dir="ltr">Attributes Reference</h4>
    <p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
    <ul dir="ltr"><li> computed_string_prop [string]  - string property description</li>
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
        
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn_instance.my_cdn_instance.<b>computed_object_prop[0]</b>.object_property)</p> 

<h2 id="provider_datasources_filters">Data Sources (using filters)</h2>

    <h3 id="cdn_datasource" dir="ltr">openapi_cdn (filters)</h3>
    <p>The cdn data source allows you to retrieve an already existing cdn resource using filters. Refer to the arguments section to learn more about how to configure the filters.</p>
    <h4 id="datasource_cdn_example_usage" dir="ltr">Example usage</h4>
    <pre>
<span>data </span><span>"openapi_cdn" "my_cdn"</span>{
    <span>filter  </span><span>{</span>
        <span>name  </span>= <span>"property name to filter by, see docs below for more info about available filter name options"</span>
        <span>values  </span>= <span>["filter value"]</span>
    <span>}</span>
<span>}</span></pre>

    <h4 id="datasource_cdn_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
            <li>filter - (Required) Object containing two properties.</li>
            <ul>
                <li>name [string]: the name should match one of the properties to filter by. The following property names are supported:
                
                    
                        <span>computed_string_prop, </span>
                    
                
                    
                        <span>computed_integer_prop, </span>
                    
                
                    
                        <span>computed_float_prop, </span>
                    
                
                    
                        <span>computed_bool_prop, </span>
                    
                
                    
                        <span>computed_sensitive_prop, </span>
                    
                
                    
                
                    
                
                    
                
                    
                
                    
                
                    
                
                </li>
                <li>values [array of string]: Values to filter by (only one value is supported at the moment).</li>
            </ul>
        </ul>
    <p dir="ltr"><b>Note: </b>If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single result only.</p>
    <h4 id="datasource_cdn_attributes_reference" dir="ltr">Attributes Reference</h4>
    <p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
    <ul dir="ltr"><li> computed_string_prop [string]  - string property description</li>
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
        
    
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn.my_cdn.<b>computed_object_prop[0]</b>.object_property)</p> `
	err := d.Render(&buf, DataSourcesTmpl)
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
	assert.Nil(t, err)
}
