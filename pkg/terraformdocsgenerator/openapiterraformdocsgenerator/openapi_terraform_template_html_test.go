package openapiterraformdocsgenerator

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
			expectedOutput: "<li> string_prop [string] - (Required) string property description</li>\n\t",
		},
		{
			name:           "required integer property",
			property:       createProperty("integer_prop", "integer", "integer property description", true, false),
			expectedOutput: "<li> integer_prop [integer] - (Required) integer property description</li>\n\t",
		},
		{
			name:           "required float property",
			property:       createProperty("float_prop", "number", "float property description", true, false),
			expectedOutput: "<li> float_prop [number] - (Required) float property description</li>\n\t",
		},
		{
			name:           "required boolean property",
			property:       createProperty("bool_prop", "boolean", "boolean property description", true, false),
			expectedOutput: "<li> bool_prop [boolean] - (Required) boolean property description</li>\n\t",
		},
		{
			name:           "required list string property",
			property:       createArrayProperty("list_string_prop", "", "string", "list_string_prop property description", true, false),
			expectedOutput: "<li> list_string_prop [] - (Required) list_string_prop property description</li>\n\t",
		},
		{
			name:           "required list integer property",
			property:       createArrayProperty("list_integer_prop", "list", "integer", "list_integer_prop property description", true, false),
			expectedOutput: "<li> list_integer_prop [list of integers] - (Required) list_integer_prop property description</li>\n\t",
		},
		{
			name:           "required list boolean property",
			property:       createArrayProperty("list_boolean_prop", "list", "boolean", "list_boolean_prop property description", true, false),
			expectedOutput: "<li> list_boolean_prop [list of booleans] - (Required) list_boolean_prop property description</li>\n\t",
		},
		{
			name:           "required list float property",
			property:       createArrayProperty("list_float_prop", "list", "number", "list_float_prop property description", true, false),
			expectedOutput: "<li> list_float_prop [list of numbers] - (Required) list_float_prop property description</li>\n\t",
		},
		{
			name:           "required object property",
			property:       Property{Name: "object_prop", Type: "object", Description: "this is an object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li><span class=\"wysiwyg-color-red\">*</span> object_prop [object] - (Required) this is an object property. The following properties compose the object schema\n        :<ul dir=\"ltr\"><li> objectPropertyRequired [string] - (Required) </li>\n\t\n        </ul>\n        </li>\n\t",
		},
		{
			name:           "required object array property",
			property:       Property{Name: "list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an list object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li> list_object_prop [list of objects] - (Required) this is an list object property. The following properties compose the object schema\n        :<ul dir=\"ltr\"><li> objectPropertyRequired [string] - (Required) </li>\n\t\n        </ul>\n        </li>\n\t",
		},
		{
			name:           "optional computed property",
			property:       Property{Name: "optional_computed_prop", Type: "string", Description: "this is an optional computed property", IsOptionalComputed: true},
			expectedOutput: "<li> optional_computed_prop [string] - (Optional) this is an optional computed property</li>\n\t",
		},
		{
			name:           "optional property",
			property:       Property{Name: "optional_prop", Type: "string", Description: "this is an optional property", Required: false},
			expectedOutput: "<li> optional_prop [string] - (Optional) this is an optional property</li>\n\t",
		},
		{
			name:           "optional property without description",
			property:       Property{Name: "optional_prop", Type: "string", Description: "", Required: false},
			expectedOutput: "<li> optional_prop [string] - (Optional) </li>\n\t",
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
			expectedOutput: "<li> computed_string_prop [string] - string property description</li>\n\t\t",
		},
		{
			name:           "computed integer property",
			property:       createProperty("computed_integer_prop", "integer", "integer property description", false, true),
			expectedOutput: "<li> computed_integer_prop [integer] - integer property description</li>\n\t\t",
		},
		{
			name:           "computed float property",
			property:       createProperty("computed_float_prop", "number", "float property description", false, true),
			expectedOutput: "<li> computed_float_prop [number] - float property description</li>\n\t\t",
		},
		{
			name:           "computed boolean property",
			property:       createProperty("computed_bool_prop", "boolean", "boolean property description", false, true),
			expectedOutput: "<li> computed_bool_prop [boolean] - boolean property description</li>\n\t\t",
		},
		{
			name:           "computed sensitive property",
			property:       Property{Name: "computed_sensitive_prop", Type: "string", Description: "this is sensitive property", Computed: true, IsSensitive: true},
			expectedOutput: "<li> computed_sensitive_prop [string] (<a href=\"#special_terms_definitions_sensitive_property\" target=\"_self\">sensitive</a>) - this is sensitive property</li>\n\t\t",
		},
		{
			name:           "computed list string property",
			property:       createArrayProperty("computed_list_string_prop", "list", "string", "list_string_prop property description", false, true),
			expectedOutput: "<li> computed_list_string_prop [list of strings] - list_string_prop property description</li>\n\t\t",
		},
		{
			name:           "computed list integer property",
			property:       createArrayProperty("computed_list_integer_prop", "list", "integer", "list_integer_prop property description", false, true),
			expectedOutput: "<li> computed_list_integer_prop [list of integers] - list_integer_prop property description</li>\n\t\t",
		},
		{
			name:           "computed list boolean property",
			property:       createArrayProperty("computed_list_boolean_prop", "list", "boolean", "list_boolean_prop property description", false, true),
			expectedOutput: "<li> computed_list_boolean_prop [list of booleans] - list_boolean_prop property description</li>\n\t\t",
		},
		{
			name:           "computed list float property",
			property:       createArrayProperty("computed_list_float_prop", "list", "number", "list_float_prop property description", false, true),
			expectedOutput: "<li> computed_list_float_prop [list of numbers] - list_float_prop property description</li>\n\t\t",
		},
		{
			name:           "computed object property",
			property:       Property{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li><span class=\"wysiwyg-color-red\">*</span> computed_object_prop [object] - this is an object property The following properties compose the object schema:\n            <ul dir=\"ltr\"><li> objectPropertyComputed [string] </li>\n\t\t\n            </ul>\n            </li>\n\t\t",
		},
		{
			name:           "computed object array property",
			property:       Property{Name: "computed_list_object_prop", Type: "list", ArrayItemsType: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
			expectedOutput: "<li> computed_list_object_prop [list of objects] - this is an object property The following properties compose the object schema:\n            <ul dir=\"ltr\"><li> objectPropertyComputed [string] </li>\n\t\t\n            </ul>\n            </li>\n\t\t",
		},
		{
			name:           "required property",
			property:       Property{Name: "optional_computed_prop", Type: "string", Description: "this is a required property", Required: true},
			expectedOutput: "",
		},
		{
			name:           "computed string property without description",
			property:       createProperty("computed_string_prop", "string", "", false, true),
			expectedOutput: "<li> computed_string_prop [string] </li>\n\t\t",
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
	example1 := `
resource "openapi_cdn" "my_cdn" {
  label    = "some label"
}`
	example2 := `
resource "openapi_cdn" "my_cdn" {
  label    = "some label"
}`
	r := ProviderResources{
		ProviderName: "openapi",
		Resources: []Resource{
			{
				Name:        "cdn",
				Description: "The 'cdn' allows you to manage 'cdn' resources using Terraform",
				ExampleUsage: []ExampleUsage{
					{
						Title:   "example title 1:",
						Example: example1,
					},
					{
						//Title:   "", (example with no title)
						Example: example2,
					},
				},
				Properties: []Property{
					// Arguments
					{Name: "object_prop", Type: "object", Description: "this is an object property", Required: true, Schema: []Property{{Name: "objectPropertyRequired", Type: "string", Required: true}, {Name: "objectPropertyComputed", Type: "string", Computed: true}}},
					// Attributes
					{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
				},
				ParentProperties: []string{"parent_id"},
				ArgumentsReference: ArgumentsReference{
					Notes: []string{"Sample note"},
				},
				KnownIssues: []KnownIssue{
					{
						Title:       "Known Issue 1",
						Description: "Description of known issue 1.",
						Examples: []ExampleUsage{
							{Title: "Example of fix for known issue 1:", Example: "known issue 1 example code"},
							{Title: "Another example of fix for known issue 1:", Example: "known 1 issue example code"},
						},
					},
					{
						Title:       "Known Issue 2",
						Description: "Description of known issue 2.",
						Examples: []ExampleUsage{
							{Title: "Example of fix for known issue 2:", Example: "known issue 2 example code"},
							{Title: "Another example of fix for known issue 2:", Example: "known issue 2 example code"},
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	expectedHTML := `<h2 id="provider_resources">Provider Resources</h2>
	
<h3 id="cdn" dir="ltr">openapi_cdn</h3>
<p>The &#39;cdn&#39; allows you to manage &#39;cdn&#39; resources using Terraform</p>
<p>If you experience any issues using this resource, please check the <a href="#resource_cdn_known_issues" target="_self">Known Issues</a> section to see if there is a fix/workaround.</p>
<h4 id="resource_cdn_example_usage" dir="ltr">Example usage</h4>
<p>example title 1:</p>
<pre>
resource &#34;openapi_cdn&#34; &#34;my_cdn&#34; {
  label    = &#34;some label&#34;
}
</pre>
<pre>
resource &#34;openapi_cdn&#34; &#34;my_cdn&#34; {
  label    = &#34;some label&#34;
}
</pre>
<h4 id="resource_cdn_arguments_reference" dir="ltr">Arguments Reference</h4>
<p dir="ltr">The following arguments are supported:</p>
<ul dir="ltr"><li><span class="wysiwyg-color-red">*</span> object_prop [object] - (Required) this is an object property. The following properties compose the object schema
        :<ul dir="ltr"><li> objectPropertyRequired [string] - (Required) </li>
	
        </ul>
        </li>
	
    </ul>
        
    
        
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn.my_cdn.<b>computed_object_prop[0]</b>.object_property)</p>
    <p><span class="wysiwyg-color-red">*Note: Sample note</span></p>


<h4 id="resource_cdn_attributes_reference" dir="ltr">Attributes Reference</h4>
<p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
<ul dir="ltr"><li><span class="wysiwyg-color-red">*</span> object_prop [object] - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string] </li>
		
            </ul>
            </li>
		<li><span class="wysiwyg-color-red">*</span> computed_object_prop [object] - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string] </li>
		
            </ul>
            </li>
		
    </ul>
        
    
        
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn.my_cdn.<b>computed_object_prop[0]</b>.object_property)</p><h4 id="resource_cdn_import" dir="ltr">Import</h4>
<p dir="ltr">
    cdn resources can be imported using the&nbsp;<code>id</code> . This is a sub-resource so the parent resource IDs (<code>[parent_id]</code>) are required to be able to retrieve an instance of this resource, e.g:
</p>
<pre dir="ltr">$ terraform import openapi_cdn.my_cdn parent_id/cdn_id</pre>
<p dir="ltr">
    <strong>Note</strong>: In order for the import to work, the 'openapi' terraform
    provider must be&nbsp;<a href="#provider_installation" target="_self">properly installed</a>. Read more about Terraform import usage&nbsp;<a href="https://www.terraform.io/docs/import/usage.html" target="_blank" rel="noopener noreferrer">here</a>.
</p>
	
<h4 id="resource_cdn_known_issues" dir="ltr">Known Issues</h4>
		
<p><i>Known Issue 1</i></p>
<p>Description of known issue 1.</p>
<p>Example of fix for known issue 1:</p>
<pre>known issue 1 example code</pre>
<p>Another example of fix for known issue 1:</p>
<pre>known 1 issue example code</pre>
		
<p><i>Known Issue 2</i></p>
<p>Description of known issue 2.</p>
<p>Example of fix for known issue 2:</p>
<pre>known issue 2 example code</pre>
<p>Another example of fix for known issue 2:</p>
<pre>known issue 2 example code</pre>
		
 `
	renderTest(t, &buf, "ProviderResources", ProviderResourcesTmpl, r, "TestProviderResourcesTmpl")
	fmt.Println(buf.String())
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
}

func TestProviderResourcesTmpl_NoResources(t *testing.T) {
	r := ProviderResources{
		ProviderName: "openapi",
		Resources:    []Resource{},
	}
	var buf bytes.Buffer
	expectedHTML := `<h2 id="provider_resources">Provider Resources</h2>
	
<p>No resources are supported at the moment.</p> `
	renderTest(t, &buf, "ProviderResources", ProviderResourcesTmpl, r, "TestProviderResourcesTmpl")
	//fmt.Println(buf.String())
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
}

func TestDataSourcesTmpl(t *testing.T) {
	dataSource := DataSources{
		ProviderName: "openapi",
		DataSourceInstances: []DataSource{
			{
				Name:         "cdn_instance",
				Description:  "Custom description for the cdn_instance data source.",
				OtherExample: "",
				Properties: []Property{
					{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
				},
			},
		},
		DataSources: []DataSource{
			{
				Name:         "cdn",
				Description:  "", // Use the default description in the template
				OtherExample: "",
				Properties: []Property{
					{Name: "computed_object_prop", Type: "object", Description: "this is an object property", Computed: true, Schema: []Property{{Name: "objectPropertyComputed", Type: "string", Computed: true}}},
				},
			},
		},
	}
	expectedHTML := `<h2 id="provider_datasources">Data Sources (using resource id)</h2>

    <h3 id="cdn_instance" dir="ltr">openapi_cdn_instance</h3>
	<p>Custom description for the cdn_instance data source.</p>
	
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
    <ul dir="ltr"><li><span class="wysiwyg-color-red">*</span> computed_object_prop [object] - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string] </li>
		
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
                
                    
                
                </li>
                <li>values [array of string]: Values to filter by (only one value is supported at the moment).</li>
            </ul>
        </ul>
    <p dir="ltr"><b>Note: </b>If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single result only.</p>
    <h4 id="datasource_cdn_attributes_reference" dir="ltr">Attributes Reference</h4>
    <p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
    <ul dir="ltr"><li><span class="wysiwyg-color-red">*</span> computed_object_prop [object] - this is an object property The following properties compose the object schema:
            <ul dir="ltr"><li> objectPropertyComputed [string] </li>
		
            </ul>
            </li>
		
        </ul>
        
    
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: openapi_cdn.my_cdn.<b>computed_object_prop[0]</b>.object_property)</p> `
	var output bytes.Buffer
	renderTest(t, &output, DataSourcesTmpl, DataSourcesTmpl, dataSource, "TestDataSourcesTmpl")
	//fmt.Println(strings.Trim(output.String(), "\n"))
	assert.Equal(t, expectedHTML, strings.Trim(output.String(), "\n"))
}

func TestDataSourcesTmpl_NoDataSources(t *testing.T) {
	dataSource := DataSources{
		ProviderName:        "openapi",
		DataSourceInstances: []DataSource{},
	}
	expectedHTML := `<h2 id="provider_datasources">Data Sources (using resource id)</h2>

No data sources using resource id are supported at the moment. 

<h2 id="provider_datasources_filters">Data Sources (using filters)</h2>

No data sources using filters are supported at the moment. `
	var output bytes.Buffer
	renderTest(t, &output, DataSourcesTmpl, DataSourcesTmpl, dataSource, "TestDataSourcesTmpl")
	//fmt.Println(strings.Trim(output.String(), "\n"))
	assert.Equal(t, expectedHTML, strings.Trim(output.String(), "\n"))
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
