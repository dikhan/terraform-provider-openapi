package openapi_terraform_docs_generator

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"strings"
	"testing"
)

func TestArgumentReferenceTmpl(t *testing.T) {
	testCases := []struct {
		name           string
		property       Property
		expectedOutput string
	}{
		{
			name:           "required string property",
			property:       createProperty("string_prop", "string", "string property description", true, false),
			expectedOutput: "<li> string_prop [string]  - (Required) string property description</li>\n\t",
		},
		{
			name:           "required integer property",
			property:       createProperty("integer_prop", "integer", "integer property description", true, false),
			expectedOutput: "<li> integer_prop [integer]  - (Required) integer property description</li>\n\t",
		},
		{
			name:           "required float property",
			property:       createProperty("float_prop", "number", "float property description", true, false),
			expectedOutput: "<li> float_prop [number]  - (Required) float property description</li>\n\t",
		},
		{
			name:           "required boolean property",
			property:       createProperty("bool_prop", "boolean", "boolean property description", true, false),
			expectedOutput: "<li> bool_prop [boolean]  - (Required) boolean property description</li>\n\t",
		},
		{
			name:           "required list string property",
			property:       createArrayProperty("list_string_prop", "", "string", "list_string_prop property description", true, false),
			expectedOutput: "<li> list_string_prop []  - (Required) list_string_prop property description</li>\n\t",
		},
		{
			name:           "required list integer property",
			property:       createArrayProperty("list_integer_prop", "list", "integer", "list_integer_prop property description", true, false),
			expectedOutput: "<li> list_integer_prop [list of integers]  - (Required) list_integer_prop property description</li>\n\t",
		},
		{
			name:           "required list boolean property",
			property:       createArrayProperty("list_boolean_prop", "list", "boolean", "list_boolean_prop property description", true, false),
			expectedOutput: "<li> list_boolean_prop [list of booleans]  - (Required) list_boolean_prop property description</li>\n\t",
		},
		{
			name:           "required list float property",
			property:       createArrayProperty("list_float_prop", "list", "number", "list_float_prop property description", true, false),
			expectedOutput: "<li> list_float_prop [list of numbers]  - (Required) list_float_prop property description</li>\n\t",
		},
		{
			name:           "required object property",
			property:       Property{Name: "object_prop", Type: "object", Description: "this is an object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li><span class=\"wysiwyg-color-red\">*</span> object_prop [object]  - (Required) this is an object property. The following properties compose the object schema\n        :<ul dir=\"ltr\"><li> objectPropertyRequired [string]  - (Required) </li>\n\t\n        </ul>\n        </li>\n\t",
		},
		{
			name:           "required object array property",
			property:       Property{Name: "list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an list object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li> list_object_prop [list of objects]  - (Required) this is an list object property. The following properties compose the object schema\n        :<ul dir=\"ltr\"><li> objectPropertyRequired [string]  - (Required) </li>\n\t\n        </ul>\n        </li>\n\t",
		},
		{
			name:           "optional computed property",
			property:       Property{Name: "optional_computed_prop", Type: "string", Description: "this is an optional computed property", IsOptionalComputed: true},
			expectedOutput: "<li> optional_computed_prop [string]  - (Optional) this is an optional computed property</li>\n\t",
		},
		{
			name:           "optional property",
			property:       Property{Name: "optional_prop", Type: "string", Description: "this is an optional property", Required: false},
			expectedOutput: "<li> optional_prop [string]  - (Optional) this is an optional property</li>\n\t",
		},
	}

	for _, tc := range testCases {
		var output bytes.Buffer
		tmpl := fmt.Sprintf(`%s
{{- template "resource_argument_reference" .}}`, ArgumentReferenceTmpl)

		renderTest(t, &output, "ArgumentReference", tmpl, tc.property, tc.name)
		assert.Equal(t, tc.expectedOutput, output.String(), tc.name)
	}
}

func TestAttributeReferenceTmpl(t *testing.T) {
	testCases := []struct {
		name           string
		property       Property
		expectedOutput string
	}{
		{
			name:           "computed string property",
			property:       createProperty("computed_string_prop", "string", "string property description", false, true),
			expectedOutput: "<li> computed_string_prop [string]  - string property description</li>\n\t\t",
		},
		{
			name:           "computed integer property",
			property:       createProperty("computed_integer_prop", "integer", "integer property description", false, true),
			expectedOutput: "<li> computed_integer_prop [integer]  - integer property description</li>\n\t\t",
		},
		{
			name:           "computed float property",
			property:       createProperty("computed_float_prop", "number", "float property description", false, true),
			expectedOutput: "<li> computed_float_prop [number]  - float property description</li>\n\t\t",
		},
		{
			name:           "computed boolean property",
			property:       createProperty("computed_bool_prop", "boolean", "boolean property description", false, true),
			expectedOutput: "<li> computed_bool_prop [boolean]  - boolean property description</li>\n\t\t",
		},
		{
			name:           "computed sensitive property",
			property:       Property{Name: "computed_sensitive_prop", Type: "string", Description: "this is sensitive property", Computed: true, IsSensitive: true},
			expectedOutput: "<li> computed_sensitive_prop [string] (<a href=\"#special_terms_definitions_sensitive_property\" target=\"_self\">sensitive</a>) - this is sensitive property</li>\n\t\t",
		},
		{
			name:           "computed list string property",
			property:       createArrayProperty("computed_list_string_prop", "", "string", "list_string_prop property description", false, true),
			expectedOutput: "<li> computed_list_string_prop []  - list_string_prop property description</li>\n\t\t",
		},
		{
			name:           "computed list integer property",
			property:       createArrayProperty("computed_list_integer_prop", "list", "integer", "list_integer_prop property description", false, true),
			expectedOutput: "<li> computed_list_integer_prop [list of integers]  - list_integer_prop property description</li>\n\t\t",
		},
		{
			name:           "computed list boolean property",
			property:       createArrayProperty("computed_list_boolean_prop", "list", "boolean", "list_boolean_prop property description", false, true),
			expectedOutput: "<li> computed_list_boolean_prop [list of booleans]  - list_boolean_prop property description</li>\n\t\t",
		},
		{
			name:           "computed list float property",
			property:       createArrayProperty("computed_list_float_prop", "list", "number", "list_float_prop property description", false, true),
			expectedOutput: "<li> computed_list_float_prop [list of numbers]  - list_float_prop property description</li>\n\t\t",
		},
		{
			name:           "computed object property",
			property:       Property{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li><span class=\"wysiwyg-color-red\">*</span> computed_object_prop [object]  - this is an object property The following properties compose the object schema:\n            <ul dir=\"ltr\"><li> objectPropertyComputed [string]  - </li>\n\t\t\n            </ul>\n            </li>\n\t\t",
		},
		{
			name:           "computed object array property",
			property:       Property{Name: "computed_list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li> computed_list_object_prop [list of objects]  - this is an object property The following properties compose the object schema:\n            <ul dir=\"ltr\"><li> objectPropertyComputed [string]  - </li>\n\t\t\n            </ul>\n            </li>\n\t\t",
		},
		{
			name:           "required property",
			property:       Property{Name: "optional_computed_prop", Type: "string", Description: "this is a required property", Required: true},
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		var output bytes.Buffer
		tmpl := fmt.Sprintf(`%s
{{- template "resource_attribute_reference" .}}`, AttributeReferenceTmpl)

		renderTest(t, &output, "AttributeReference", tmpl, tc.property, tc.name)
		assert.Equal(t, tc.expectedOutput, output.String(), tc.name)
	}
}

func TestProviderInstallationTmpl(t *testing.T) {
	pi := ProviderInstallation{
		ProviderName: "openapi",
		Example:      "➜ ~ This is an example",
		Other:        "Some more info about the installation",
		OtherCommand: "➜ ~ init_command do_something",
	}
	var buf bytes.Buffer
	expectedHTML := `<h2 id="provider_installation">Provider Installation</h2>
<p>
  In order to provision 'openapi' Terraform resources, you need to first install the 'openapi'
  Terraform plugin by running&nbsp;the following command (you must be running Terraform &gt;= 0.12):
</p>
<pre>➜ ~ This is an example</pre>
<p>
  <span>Some more info about the installation</span>
</p>
<pre dir="ltr">➜ ~ init_command do_something
➜ ~ terraform init &amp;&amp; terraform plan
</pre>`
	renderTest(t, &buf, "ProviderInstallation", ProviderInstallationTmpl, pi, "ProviderInstallationTmpl")
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
}

func TestProviderConfigurationTmpl(t *testing.T) {
	pc := ProviderConfiguration{
		ProviderName: "openapi",
		Regions:      []string{"rst1"},
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
	}
	var buf bytes.Buffer
	expectedHTML := `<h2 id="provider_configuration">Provider Configuration</h2>
<h4 id="provider_configuration_example_usage" dir="ltr">Example Usage</h4>
    <pre>
<span>provider </span><span>"openapi" </span>{
<span>  token  </span>= <span>"..."</span>
<span>}</span>
</pre>

    <p>Using the default region (rst1):</p>
    <pre>
<span>provider </span><span>"openapi" </span>{
<span>  # Resources using this default provider will be created in the 'rst1' region<br>  ...<br></span>}
    </pre>
    

    <h4 id="provider_configuration_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
        <li><span>token [string] - (Required) .</span></li></li>
      <li>
          region [string] - (Optional) The region location to be used&nbsp;([rst1]). If region isn't specified, the default is "rst1".
      </li>
    
    </ul>`
	renderTest(t, &buf, "TestProviderConfiguration", ProviderConfigurationTmpl, pc, "TestProviderConfigurationTmpl")
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
}

func TestProviderResourcesTmpl(t *testing.T) {
	r := ProviderResources{
		ProviderName: "openapi",
		Resources: []Resource{
			{
				Name:        "cdn",
				Description: "The 'cdn' allows you to manage 'cdn' resources using Terraform",
				Properties: []Property{
					// Arguments
					Property{Name: "object_prop", Type: "object", Description: "this is an object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
					// Attributes
					Property{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
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


    <h3 id="cdn" dir="ltr">openapi_cdn</h3><p>The &#39;cdn&#39; allows you to manage &#39;cdn&#39; resources using Terraform</p>
    <h4 id="resource_cdn_example_usage" dir="ltr">Example usage</h4>
<pre>
<span>resource </span><span>"openapi_cdn" "my_cdn"</span>{
    <span>object_prop  </span><span>{</span>
                
    <span>objectPropertyRequired  </span>= <span>"objectPropertyRequired"</span>
                
            <span>}</span>
<span>}</span>
</pre>
<h4 id="resource_cdn_arguments_reference" dir="ltr">Arguments Reference</h4>
<p dir="ltr">The following arguments are supported:</p>
<ul dir="ltr"><li><span class="wysiwyg-color-red">*</span> object_prop [object]  - (Required) this is an object property. The following properties compose the object schema
        :<ul dir="ltr"><li> objectPropertyRequired [string]  - (Required) </li>
	
        </ul>
        </li>
	
    </ul>
        
    
        
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn.my_cdn.<b>computed_object_prop[0]</b>.object_property)</p>
    <p><span class="wysiwyg-color-red">*Note: Sample note</span></p>


<h4 id="resource_cdn_attributes_reference" dir="ltr">Attributes Reference</h4>
<p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
<ul dir="ltr"><li><span class="wysiwyg-color-red">*</span> object_prop [object]  - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string]  - </li>
		
            </ul>
            </li>
		<li><span class="wysiwyg-color-red">*</span> computed_object_prop [object]  - this is an object property The following properties compose the object schema:
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
	renderTest(t, &buf, "ProviderResources", ProviderResourcesTmpl, r, "TestProviderResourcesTmpl")
	fmt.Println(buf.String())
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
}

func renderTest(t *testing.T, w io.Writer, templateName string, templateContent string, data interface{}, testName string) {
	tmpl, err := template.New(templateName).Parse(templateContent)
	assert.Nil(t, err, testName)
	err = tmpl.Execute(w, data)
	assert.Nil(t, err, testName)
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
