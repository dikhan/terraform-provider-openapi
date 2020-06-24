package templates

// ZendeskTmpl contains the template used to render the TerraformProviderDocumentation struct as HTML formatted for Zendesk
var ZendeskTmpl = `{{define "resource_example"}}
{{- if .Required}}
    {{if eq .Type "string" -}}
        <span>{{.Name}}  </span>= <span>"{{.Name}}"</span>
    {{- else if eq .Type "integer" -}}
        <span>{{.Name}}  </span>= <span>1234</span>
    {{- else if eq .Type "boolean" -}}
        <span>{{.Name}}  </span>= <span>true</span>
    {{- else if eq .Type "float" -}}
        <span>{{.Name}}  </span>= <span>12.95</span>
    {{- else if and (eq .Type "array") (eq .ArrayItemsType "string") -}}
        <span>{{.Name}}  </span>= <span>["string1", "string2"]</span>
    {{- else if and (eq .Type "array") (eq .ArrayItemsType "integer") -}}
        <span>{{.Name}}  </span>= <span>[1234, 4567]</span>
    {{- else if and (eq .Type "array") (eq .ArrayItemsType "boolean") -}}
        <span>{{.Name}}  </span>= <span>[true, false]</span>
    {{- else if and (eq .Type "array") (eq .ArrayItemsType "float") -}}
        <span>{{.Name}}  </span>= <span>[12.36, 99.45]</span>
    {{- else -}}
        {{- if and (eq .Type "object") -}}
        <span>{{.Name}}  </span><span>{</span>
            {{- range .Schema}}
                {{template "resource_example" .}}
            {{- end}}
            <span>}</span>
        {{- end -}}
    {{- end -}}
{{- end -}}
{{end}}

{{- define "resource_argument_reference" -}}
    {{- $required := "Optional" -}}
    {{- if .Required -}}
        {{- $required = "Required" -}}
    {{end}}
	{{- if or .Required (and (not .Required) (not .Computed)) .IsOptionalComputed -}}
    <li>{{if eq .Type "object"}}<span class="wysiwyg-color-red">*</span>{{end}} {{.Name}} [{{.Type}} {{- if eq .Type "array" }} of {{.ArrayItemsType}}s{{- end -}}] - ({{$required}}) {{.Description}}
        {{- if or (eq .Type "object") (eq .ArrayItemsType "object")}}. The following properties compose the object schema
        :<ul dir="ltr">
            {{- range .Schema}}
                {{- template "resource_argument_reference" .}}
            {{- end}}
        </ul>
        {{ end -}}
    </li>
	{{end}}
{{- end -}}

{{- define "resource_attribute_reference" -}}
    {{- if or .Computed .ContainsComputedSubProperties -}}
		{{- if and .Schema (not .ContainsComputedSubProperties) -}}{{/* objects or arrays of objects that DO NOT have computed props are ignored since they will be documented in the arguments section */}}
		{{else}}
        <li>{{if eq .Type "object"}}<span class="wysiwyg-color-red">*</span>{{end}} {{.Name}} [{{.Type}} {{- if eq .Type "array" }} of {{.ArrayItemsType}}s{{- end -}}] - {{.Description}}
            {{- if or (eq .Type "object") (eq .ArrayItemsType "object")}} The following properties compose the object schema:
            <ul dir="ltr">
                {{- range .Schema}}
					{{- template "resource_attribute_reference" .}}
                {{- end}}
            </ul>
            {{ end -}}
        </li>
		{{end}}
    {{end}}
{{- end -}}


<p dir="ltr">
  This guide lists the configuration for '{{.ProviderName}}' Terraform provider
  resources that can be managed using
  <a href="https://www.hashicorp.com/blog/announcing-terraform-0-12/" target="_self">Terraform v0.12</a>.&nbsp;
</p>
<ul>
  <li>
    <a href="#provider_installation" target="_self">Provider Installation</a>
  </li>
  {{if .ProviderConfiguration.Regions }}
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
            {{- end}}
        </ul>
    </li>
    <li>
        <a href="#provider_datasources" target="_self">Data Sources (using resource id)</a>
        <ul>
            {{range .DataSources.DataSourceInstances -}}
                <li><a href="#{{.Name}}" target="_self">{{$.ProviderName}}_{{.Name}}</a></li>
            {{- end}}
        </ul>
    </li>
    <li>
        <a href="#provider_datasources_filters" target="_self">Data Sources (using filters)</a>
        <ul>
            {{range .DataSources.DataSources -}}
                <li><a href="#{{.Name}}_datasource" target="_self">{{$.ProviderName}}_{{.Name}}</a></li>
            {{- end}}
        </ul>
    </li>
</ul>
<h2 id="provider_installation">Provider Installation</h2>
<p>
  In order to provision '{{.ProviderName}}' Terraform resources, you need to first install the '{{.ProviderName}}'
  Terraform plugin by running&nbsp;the following command (you must be running Terraform &gt;= 0.12):
</p>
<pre>{{.ProviderInstallation.Example}}</pre>
<p>
  <span>{{.ProviderInstallation.Other}}</span>
</p>
<pre dir="ltr">{{.ProviderInstallation.OtherCommand}}$ terraform init &amp;&amp; terraform plan</pre>

<h2 id="provider_configuration">Provider Configuration</h2>
<h4 id="provider_configuration_example_usage" dir="ltr">Example Usage</h4>
{{if .ProviderConfiguration.ConfigProperties}}
    <pre>
<span>provider </span><span>"{{.ProviderName}}" </span>{
{{- range .ProviderConfiguration.ConfigProperties}}
<span>  {{.Name}}  </span>= <span>"..."</span>
{{- end}}
<span>}</span>
</pre>
{{- end}}
{{if .ProviderConfiguration.Regions }}
    <p>Using the default region ({{index .ProviderConfiguration.Regions 0}}):</p>
    <pre>
<span>provider </span><span>"{{.ProviderName}}" </span>{
<span>  # Resources using this default provider will be created in the '{{index .ProviderConfiguration.Regions 0}}' region<br>  ...<br></span>}
    </pre>
    {{ if gt (len .ProviderConfiguration.Regions) 1 -}}
        <p>Using a specific region ({{index .ProviderConfiguration.Regions 1}}):</p>
    <pre>
<span>provider </span><span>"{{.ProviderName}}" </span>{
<span>  alias  </span>= <span>"{{index .ProviderConfiguration.Regions 1}}"</span>
<span>  region </span>= <span>"{{index .ProviderConfiguration.Regions 1}}"<br>  ...<br></span>}
<br>resource<span>"{{.ProviderName}}_resource" "my_resource" {</span>
<span>  provider = "{{.ProviderName}}.{{index .ProviderConfiguration.Regions 1}}"<br>  ...<br>}</span>
    </pre>
    {{- end }}

    <h4 id="provider_configuration_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
        {{- range .ProviderConfiguration.ConfigProperties -}}
        {{- $required := "Optional" -}}
        {{- if .Required -}}
            {{- $required = "Required" -}}
        {{end}}
        <li><span>{{.Name}} [{{.Type}}] - ({{$required}}) {{.Description}}.</span></li></li>
        {{- end -}}
    {{if .ProviderConfiguration.Regions }}
      <li>
          region [string] - (Optional) The region location to be used&nbsp;({{.ProviderConfiguration.Regions}}). If region isn't specified, the default is "{{index .ProviderConfiguration.Regions 0}}".
      </li>
    {{end}}
    </ul>
{{end}}
<h2 id="provider_resources">Provider Resources</h2>

{{range .ProviderResources.Resources -}}
    {{ $resource := . }}
    <h3 id="{{.Name}}" dir="ltr">{{$.ProviderName}}_{{.Name}}</h3>
    {{- if ne .Description "" -}}
        <p>{{.Description}}</p>
    {{- end}}
    <h4 id="resource_{{.Name}}_example_usage" dir="ltr">Example usage</h4>
{{- if .ExampleUsage}}
	{{- range .ExampleUsage -}}
		<pre>
{{.Example}}
		</pre>
	{{- end}}
{{else}}
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
{{ $object_property_name := "" }}
{{- range $resource.Properties -}}
    {{- if (eq .Type "object")}}
        {{ $object_property_name = .Name }}
    {{end}}
{{- end}}
{{- if ne $object_property_name ""}}
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: {{$.ProviderName}}_{{.Name}}.my_{{.Name}}.<b>{{$object_property_name}}[0]</b>.object_property)</p>
{{end}}
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
{{- end}}
{{ $object_property_name := "" }}
{{- range $resource.Properties -}}
    {{- if (eq .Type "object")}}
        {{ $object_property_name = .Name }}
    {{end}}
{{- end}}
{{- if ne $object_property_name ""}}
    <p><span class="wysiwyg-color-red">* </span>Note: Object type properties are internally represented (in the state file) as a list of one elem due to <a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_blank">Terraform SDK's limitation for supporting complex object types</a>. Please index on the first elem of the array to reference the object values (eg: {{$.ProviderName}}_{{.Name}}.my_{{.Name}}.<b>{{$object_property_name}}[0]</b>.object_property)</p>
{{end}}

<h4 id="resource_{{.Name}}_import" dir="ltr">Import</h4>
<p dir="ltr">
    {{.Name}} resources can be imported using the&nbsp;<code>id</code>, e.g:
</p>
<pre dir="ltr">$ terraform import {{.Name}}.my_{{.Name}} id</pre>
<p dir="ltr">
    <strong>Note</strong>: In order for the import to work, the '{{$.ProviderName}}' terraform
    provider must be&nbsp;<a href="#provider_installation" target="_self">properly installed</a>. Read more about Terraform import usage&nbsp;<a href="https://www.terraform.io/docs/import/usage.html" target="_blank" rel="noopener noreferrer">here</a>.
</p>

{{end}} {{/* END range .ProviderResources.Resources */}}

<h2 id="provider_datasources">Data Sources (using resource id)</h2>

{{range .DataSources.DataSourceInstances -}}
    {{ $datasource := . }}
    <h3 id="{{.Name}}" dir="ltr">{{$.ProviderName}}_{{.Name}}</h3>
    <p>Retrieve an existing resource using it's ID</p>
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
{{end}} {{/* END range .DataSources.DataSourceInstances */}}

<h2 id="provider_datasources_filters">Data Sources (using filters)</h2>
{{range .DataSources.DataSources -}}
    {{ $datasource := . }}
    <h3 id="{{.Name}}_datasource" dir="ltr">{{$.ProviderName}}_{{.Name}} (filters)</h3>
    <p>The {{.Name}} data source allows you to retrieve an already existing {{.Name}} resource using filters. Refer to the arguments section to learn more about how to configure the filters.</p>
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
                    {{if or (eq .Type "string") (eq .Type "integer") (eq .Type "float") (eq .Type "boolean")}}
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
{{end}} {{/* END range .DataSources.DataSources */}}`
