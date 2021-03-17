package openapiterraformdocsgenerator

import "fmt"

// TableOfContentsTmpl contains the template used to render the table of contents as HTML formatted for Zendesk
var TableOfContentsTmpl = `<p dir="ltr">
  This guide lists the configuration for '{{.ProviderName}}' Terraform provider
  resources that can be managed using
  <a href="https://www.hashicorp.com/blog/announcing-terraform-0-12/" target="_self">Terraform v0.12</a>.&nbsp;
</p>
{{- if .ProviderNotes}}
{{range .ProviderNotes -}}
	<p><span class="wysiwyg-color-red">*Note: {{.}}</span></p>
{{end}}
{{- end -}}
<ul>
  <li>
    <a href="#provider_installation" target="_self">Provider Installation</a>
  </li>
  {{if or .ProviderConfiguration.Regions .ProviderConfiguration.ConfigProperties }}
    <li>
      <a href="#provider_configuration" target="_self">Provider Configuration</a>
      <ul>
        <li>
          <a href="#provider_configuration_example_usage" target="_self">Example Usage</a>
        </li>
        <li>
          <a href="#provider_configuration_arguments_reference" target="_self">Arguments Reference</a>
        </li>
      </ul>
    </li>
  {{end}}
    <li>
        <a href="#provider_resources" target="_self">Provider Resources</a>
        <ul>
            {{range .ProviderResources.Resources -}}
                <li><a href="#{{.Name}}" target="_self">{{$.ProviderName}}_{{.Name}}</a></li>
            {{end -}}
        </ul>
    </li>
    <li>
        <a href="#provider_datasources" target="_self">Data Sources (using resource id)</a>
        <ul>
            {{range .DataSources.DataSourceInstances -}}
                <li><a href="#{{.Name}}" target="_self">{{$.ProviderName}}_{{.Name}}</a></li>
            {{end -}}
        </ul>
    </li>
    <li>
        <a href="#provider_datasources_filters" target="_self">Data Sources (using filters)</a>
        <ul>
            {{range .DataSources.DataSources -}}
                <li><a href="#{{.Name}}_datasource" target="_self">{{$.ProviderName}}_{{.Name}}</a></li>
            {{end -}}
        </ul>
    </li>

{{ if .ShowSpecialTermsDefinitions }}
  <li>
    <a href="#special_terms_definitions" target="_self">Special Terms Definitions</a>
    <ul>
{{ if .ProviderResources.ContainsResourcesWithSecretProperties }}
      <li>
        <a href="#special_terms_definitions_sensitive_property" target="_self">Sensitive Property</a>
      </li>
{{end}}
    </ul>
  </li>
{{end}}
</ul>`

// ProviderInstallationTmpl contains the template used to render the TerraformProviderDocumentation.ProviderInstallation struct as HTML formatted for Zendesk
var ProviderInstallationTmpl = `<h2 id="provider_installation">Provider Installation</h2>
<p>
  In order to provision '{{.ProviderName}}' Terraform resources, you need to first install the '{{.ProviderName}}'
  Terraform plugin by running&nbsp;the following command (you must be running Terraform &gt;= 0.12):
</p>
<pre>{{.Example}}</pre>
<p>
  <span>{{.Other}}</span>
</p>
<pre dir="ltr">
{{- if .OtherCommand -}}
	{{- .OtherCommand -}}
{{- end}}
➜ ~ terraform init &amp;&amp; terraform plan
</pre>
<p>
<b>Note:</b> As of Terraform &gt;= 0.13 each Terraform module must declare which providers it requires, so that Terraform can install and use them. If you are using Terraform &gt;= 0.13, copy into your .tf file the 
following snippet already populated with the provider configuration: 
<pre dir="ltr">
terraform {
  required_providers {
    {{.ProviderName}} = {
      source  = "{{.Hostname}}/{{.Namespace}}/{{.ProviderName}}"
{{- if .PluginVersionConstraint }}
      version = "{{.PluginVersionConstraint}}"
{{- else}}
      version = ">= 2.0.1"
{{- end}} 
    }
  }
}
</pre>
</p>`

// ProviderConfigurationTmpl contains the template used to render the TerraformProviderDocumentation.ProviderConfiguration struct as HTML formatted for Zendesk
var ProviderConfigurationTmpl = `{{if or .Regions .ConfigProperties}}
<h2 id="provider_configuration">Provider Configuration</h2>
<h4 id="provider_configuration_example_usage" dir="ltr">Example Usage</h4>
    <pre>
<span>provider </span><span>"{{.ProviderName}}" </span>{
{{- range .ConfigProperties}}
<span>  {{.Name}}  </span>= <span>"..."</span>
{{- end}}
<span>}</span>
</pre>
{{- end}}
{{if .Regions }}
    <p>Using the default region ({{index .Regions 0}}):</p>
    <pre>
<span>provider </span><span>"{{.ProviderName}}" </span>{
<span>  # Resources using this default provider will be created in the '{{index .Regions 0}}' region<br>  ...<br></span>}
    </pre>
    {{ if gt (len .Regions) 1 -}}
        <p>Using a specific region ({{index .Regions 1}}):</p>
    <pre>
<span>provider </span><span>"{{.ProviderName}}" </span>{
<span>  alias  </span>= <span>"{{index .Regions 1}}"</span>
<span>  region </span>= <span>"{{index .Regions 1}}"<br>  ...<br></span>}
<br>resource<span>"{{.ProviderName}}_resource" "my_resource" {</span>
<span>  provider = "{{.ProviderName}}.{{index .Regions 1}}"<br>  ...<br>}</span>
    </pre>
    {{- end }}

    <h4 id="provider_configuration_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
        {{- range .ConfigProperties -}}
        {{- $required := "Optional" -}}
        {{- if .Required -}}
            {{- $required = "Required" -}}
        {{end}}
        <li><span>{{.Name}} [{{.Type}}] - ({{$required}}) {{.Description}}.</span></li></li>
        {{- end -}}
    {{if .Regions }}
      <li>
          region [string] - (Optional) The region location to be used&nbsp;({{.Regions}}). If region isn't specified, the default is "{{index .Regions 0}}".
      </li>
    {{end}}
    </ul>
{{end}}`

// ProviderResourcesTmpl contains the template used to render the TerraformProviderDocumentation.ProviderResources struct as HTML formatted for Zendesk
var ProviderResourcesTmpl = fmt.Sprintf(`{{define "resource_example"}}
{{- if .Required}}
    {{if eq .Type "string" -}}
        <span>{{.Name}}  </span>= <span>"{{.Name}}"</span>
    {{- else if eq .Type "integer" -}}
        <span>{{.Name}}  </span>= <span>1234</span>
    {{- else if eq .Type "boolean" -}}
        <span>{{.Name}}  </span>= <span>true</span>
    {{- else if eq .Type "number" -}}
        <span>{{.Name}}  </span>= <span>12.95</span>
    {{- else if and (eq .Type "list") (eq .ArrayItemsType "string") -}}
        <span>{{.Name}}  </span>= <span>["{{.Name}}1", "{{.Name}}2"]</span>
    {{- else if and (eq .Type "list") (eq .ArrayItemsType "integer") -}}
        <span>{{.Name}}  </span>= <span>[1234, 4567]</span>
    {{- else if and (eq .Type "list") (eq .ArrayItemsType "boolean") -}}
        <span>{{.Name}}  </span>= <span>[true, false]</span>
    {{- else if and (eq .Type "list") (eq .ArrayItemsType "number") -}}
        <span>{{.Name}}  </span>= <span>[12.36, 99.45]</span>
    {{- else -}}
        {{- if or (eq .Type "object") (and (eq .Type "list") (eq .ArrayItemsType "object")) -}}
        <span>{{.Name}}  </span><span>{</span>
            {{- range .Schema}}
                {{template "resource_example" .}}
            {{- end}}
            <span>}</span>
        {{- end -}}
    {{- end -}}
{{- end -}}
{{end}}

%s

%s

<h2 id="provider_resources">Provider Resources</h2>
	{{if not .Resources}}
<p>No resources are supported at the moment.</p>
	{{- end -}}
{{range .Resources -}}
	{{ $resource := . }}
<h3 id="{{.Name}}" dir="ltr">{{$.ProviderName}}_{{.Name}}</h3>
{{if ne .Description "" -}}
<p>{{.Description}}</p>
{{- end}}
{{- if .KnownIssues}}
<p>If you experience any issues using this resource, please check the <a href="#resource_{{.Name}}_known_issues" target="_self">Known Issues</a> section to see if there is a fix/workaround.</p>
{{end -}}
<h4 id="resource_{{.Name}}_example_usage" dir="ltr">Example usage</h4>
	{{- if .ExampleUsage}}
		{{- range .ExampleUsage}}
			{{- if .Title}}
<p>{{.Title}}</p>
			{{- end}}
<pre>
{{- .Example}}
</pre>
		{{- end}}
	{{- else}}
<pre>
<span>resource </span><span>"{{$.ProviderName}}_{{$resource.Name}}" "my_{{$resource.Name}}"</span>{
{{- range $resource.Properties -}}
    {{template "resource_example" .}}
{{- end}}
<span>}</span>
</pre>
{{- end}}
<h4 id="resource_{{.Name}}_arguments_reference" dir="ltr">Arguments Reference</h4>
<p dir="ltr">The following arguments are supported:</p>
{{if $resource.Properties -}}
    <ul dir="ltr">
        {{- range $resource.Properties -}}
            {{- template "resource_argument_reference" . -}}
        {{- end}}
    </ul>
{{- end}}
{{- $object_property_name := "" -}}
{{- range $resource.Properties -}}
    {{- if (eq .Type "object")}}
        {{ $object_property_name = .Name }}
    {{end}}
{{- end -}}
{{- if ne $object_property_name "" -}}
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: {{$.ProviderName}}_{{.Name}}.my_{{.Name}}.<b>{{$object_property_name}}[0]</b>.object_property)</p>
{{- end -}}
{{range .ArgumentsReference.Notes}}
    <p><span class="wysiwyg-color-red">*Note: {{.}}</span></p>
{{end}}

<h4 id="resource_{{.Name}}_attributes_reference" dir="ltr">Attributes Reference</h4>
<p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
{{if $resource.Properties -}}
    <ul dir="ltr">
        {{- range $resource.Properties -}}
            {{- template "resource_attribute_reference" . -}}
        {{- end}}
    </ul>
{{- end -}}
{{- $object_property_name := "" -}}
{{- range $resource.Properties -}}
    {{- if (eq .Type "object")}}
        {{ $object_property_name = .Name }}
    {{end}}
{{- end -}}
{{- if ne $object_property_name "" -}}
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: {{$.ProviderName}}_{{.Name}}.my_{{.Name}}.<b>{{$object_property_name}}[0]</b>.object_property)</p>
{{- end -}}

<h4 id="resource_{{.Name}}_import" dir="ltr">Import</h4>
<p dir="ltr">
    {{.Name}} resources can be imported using the&nbsp;<code>id</code> {{if ne $resource.BuildImportIDsExample "id"}}. This is a sub-resource so the parent resource IDs (<code>{{$resource.ParentProperties}}</code>) are required to be able to retrieve an instance of this resource{{end}}, e.g:
</p>
<pre dir="ltr">$ terraform import {{$.ProviderName}}_{{.Name}}.my_{{.Name}} {{$resource.BuildImportIDsExample}}</pre>
<p dir="ltr">
    <strong>Note</strong>: In order for the import to work, the '{{$.ProviderName}}' terraform
    provider must be&nbsp;<a href="#provider_installation" target="_self">properly installed</a>. Read more about Terraform import usage&nbsp;<a href="https://www.terraform.io/docs/import/usage.html" target="_blank" rel="noopener noreferrer">here</a>.
</p>
	{{if .KnownIssues}}
<h4 id="resource_{{.Name}}_known_issues" dir="ltr">Known Issues</h4>
		{{range .KnownIssues}}
<p><i>{{.Title}}</i></p>
<p>{{.Description}}</p>
			{{- range .Examples}}
				{{- if .Title}}
<p>{{.Title}}</p>
				{{- end}}
<pre>{{.Example}}</pre>
			{{- end}}
		{{end}}
	{{- end}}
{{end}} {{/* END range .Resources */}}`, ArgumentReferenceTmpl, AttributeReferenceTmpl)

// ArgumentReferenceTmpl contains the definition used in resources to render the arguments
var ArgumentReferenceTmpl = `{{- define "resource_argument_reference" -}}
    {{- $required := "Optional" -}}
    {{- if .Required -}}
        {{- $required = "Required" -}}
    {{end}}
	{{- if or .Required (and (not .Required) (not .Computed)) .IsOptionalComputed -}}
    <li>{{if eq .Type "object"}}<span class="wysiwyg-color-red">*</span>{{end}} {{.Name}} [{{.Type}} {{- if eq .Type "list" }} of {{.ArrayItemsType}}s{{- end -}}] {{- if .IsSensitive -}}(<a href="#special_terms_definitions_sensitive_property" target="_self">sensitive</a>){{- end}} - ({{$required}}) {{if .IsParent}}The {{.Name}} that this resource belongs to{{else}}{{.Description}}{{end}}
        {{- if or (eq .Type "object") (eq .ArrayItemsType "object")}}. The following properties compose the object schema
        :<ul dir="ltr">
            {{- range .Schema}}
                {{- template "resource_argument_reference" .}}
            {{- end}}
        </ul>
        {{ end -}}
    </li>
	{{end}}
{{- end -}}`

// AttributeReferenceTmpl contains the definition used in resources to render the attributes
var AttributeReferenceTmpl = `{{- define "resource_attribute_reference" -}}
    {{- if or .Computed .ContainsComputedSubProperties -}}
		{{- if and .Schema (not .ContainsComputedSubProperties) -}}{{- /* objects or arrays of objects that DO NOT have computed props are ignored since they will be documented in the arguments section */ -}}
		{{- else -}}
        <li>{{if eq .Type "object"}}<span class="wysiwyg-color-red">*</span>{{end}} {{.Name}} [{{.Type}} {{- if eq .Type "list" }} of {{.ArrayItemsType}}s{{- end -}}] {{ if .IsSensitive }}(<a href="#special_terms_definitions_sensitive_property" target="_self">sensitive</a>) {{end -}}{{- if .Description }}- {{.Description}} {{- end -}}
            {{- if or (eq .Type "object") (eq .ArrayItemsType "object")}} The following properties compose the object schema:
            <ul dir="ltr">
                {{- range .Schema}}
					{{- template "resource_attribute_reference" .}}
                {{- end}}
            </ul>
            {{ end -}}
        </li>
		{{end}}
    {{- end}}
{{- end -}}`

// DataSourcesTmpl contains the template used to render the TerraformProviderDocumentation.DataSources struct as HTML formatted for Zendesk
var DataSourcesTmpl = fmt.Sprintf(`%s

<h2 id="provider_datasources">Data Sources (using resource id)</h2>
{{if not .DataSourceInstances}}
No data sources using resource id are supported at the moment.
{{- end -}}
{{range .DataSourceInstances -}}
    {{ $datasource := . }}
    <h3 id="{{.Name}}" dir="ltr">{{$.ProviderName}}_{{.Name}}</h3>
	{{if ne .Description "" -}}
	<p>{{.Description}}</p>
	{{else}}
	<p>Retrieve an existing resource using it's ID</p>
	{{- end}}
    <h4 id="datasource_{{.Name}}_example_usage" dir="ltr">Example usage</h4>
<pre><span>data </span><span>"{{$.ProviderName}}_{{$datasource.Name}}" "my_{{$datasource.Name}}"</span>{
    id = "existing_resource_id"
<span>}</span></pre>
    <h4 id="datasource_{{.Name}}_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
        <li>id - (Required) ID of the existing resource to retrieve</li>
    </ul>
    <h4 id="datasource_{{.Name}}_attributes_reference" dir="ltr">Attributes Reference</h4>
    <p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
    {{if $datasource.Properties -}}
        <ul dir="ltr">
            {{- range $datasource.Properties -}}
                {{- template "resource_attribute_reference" . -}}
            {{- end}}
        </ul>
    {{- end}}
{{- $object_property_name := "" -}}
{{- range $datasource.Properties -}}
    {{- if (eq .Type "object")}}
        {{ $object_property_name = .Name }}
    {{end}}
{{- end -}}
{{- if ne $object_property_name "" -}}
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: {{$.ProviderName}}_{{.Name}}.my_{{.Name}}.<b>{{$object_property_name}}[0]</b>.object_property)</p>
{{- end -}}
{{end}} {{/* END range .DataSourceInstances */}}

<h2 id="provider_datasources_filters">Data Sources (using filters)</h2>
{{if not .DataSources}}
No data sources using filters are supported at the moment.
{{- end -}}
{{range .DataSources -}}
    {{ $datasource := . }}
	<h3 id="{{.Name}}_datasource" dir="ltr">{{$.ProviderName}}_{{.Name}} (filters)</h3>
	{{if ne .Description "" -}}
	<p>{{.Description}}</p>
	{{else}}
	<p>The {{.Name}} data source allows you to retrieve an already existing {{.Name}} resource using filters. Refer to the arguments section to learn more about how to configure the filters.</p>
	{{- end}}
    <h4 id="datasource_{{.Name}}_example_usage" dir="ltr">Example usage</h4>
    <pre>
<span>data </span><span>"{{$.ProviderName}}_{{$datasource.Name}}" "my_{{$datasource.Name}}"</span>{
    <span>filter  </span><span>{</span>
        <span>name  </span>= <span>"property name to filter by, see docs below for more info about available filter name options"</span>
        <span>values  </span>= <span>["filter value"]</span>
    <span>}</span>
<span>}</span></pre>

    <h4 id="datasource_{{.Name}}_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    {{if $datasource.Properties -}}
        <ul dir="ltr">
            <li>filter - (Required) Object containing two properties.</li>
            <ul>
                <li>name [string]: the name should match one of the properties to filter by. The following property names are supported:
                {{range $datasource.Properties}}
                    {{if or (eq .Type "string") (eq .Type "integer") (eq .Type "number") (eq .Type "boolean")}}
                        <span>{{.Name}}, </span>
                    {{end}}
                {{end}}
                </li>
                <li>values [array of string]: Values to filter by (only one value is supported at the moment).</li>
            </ul>
        </ul>
    {{- end}}
    <p dir="ltr"><b>Note: </b>If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single result only.</p>
    <h4 id="datasource_{{.Name}}_attributes_reference" dir="ltr">Attributes Reference</h4>
    <p dir="ltr">In addition to all arguments above, the following attributes are exported:</p>
    {{if $datasource.Properties -}}
        <ul dir="ltr">
            {{- range $datasource.Properties -}}
                {{if .Computed -}}
                    {{- template "resource_attribute_reference" . -}}
                {{- end}}
            {{- end}}
        </ul>
    {{- end}}
{{- $object_property_name := "" -}}
{{- range $datasource.Properties -}}
    {{- if (eq .Type "object")}}
        {{ $object_property_name = .Name }}
    {{end}}
{{- end -}}
{{- if ne $object_property_name ""}}
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: {{$.ProviderName}}_{{.Name}}.my_{{.Name}}.<b>{{$object_property_name}}[0]</b>.object_property)</p>
{{- end -}}
{{end}} {{/* END range .DataSources */}}`, AttributeReferenceTmpl)

// SpecialTermsTmpl contains the template used to render the special terms definitions section as HTML formatted for Zendesk
var SpecialTermsTmpl = `{{ if .ShowSpecialTermsDefinitions }}
<h2 id="special_terms_definitions">Special Terms Definitions</h2>
<p>
  This section describes specific terms used throughout this document to clarify their meaning in the context of Terraform.
</p>
{{ if .ProviderResources.ContainsResourcesWithSecretProperties }}
<h3 id="special_terms_definitions_sensitive_property">
  <span style="font-weight:400">Sensitive Property</span>
</h3>
<p>
  <span style="font-weight:400">The ‘{{.ProviderName}}’ Terraform plugin treats secret properties as ‘sensitive’. As per </span><a href="https://github.com/hashicorp/terraform-plugin-sdk/blob/9f0df37a8fdb2627ae32db6ceaf7f036d89b6768/helper/schema/schema.go#L245" target="_self">Terraform documentation</a><span style="font-weight:400">, this means the following:</span>
</p>
<pre><span style="font-weight:400">// Sensitive ensures that the attribute's value does not get displayed <br>// in logs or regular output. It should be used for passwords or other <br>// secret fields. Future versions of Terraform may encrypt these values.</span></pre>
<p>
  <span style="font-weight:400">Please note that even though the secret values don’t get displayed in the logs or regular output, the state file will still store the secrets. </span><span style="font-weight:400">As per <a href="https://www.terraform.io/docs/state/sensitive-data.html">Terraform’s official recommendations</a> on how to treat Sensitive Data in State, if your state file may contain sensitive information always treat the State itself as sensitive data.</span>
</p>
<p>&nbsp;</p>
{{end}}
{{end}}`
