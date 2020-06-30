package zendesk

// DataSourcesTmpl contains the template used to render the TerraformProviderDocumentation.DataSources struct as HTML formatted for Zendesk
var DataSourcesTmpl = `{{- define "resource_attribute_reference" -}}
    {{- if or .Computed .ContainsComputedSubProperties -}}
		{{- if and .Schema (not .ContainsComputedSubProperties) -}}{{- /* objects or arrays of objects that DO NOT have computed props are ignored since they will be documented in the arguments section */ -}}
		{{- else -}}
        <li>{{if eq .Type "object"}}<span class="wysiwyg-color-red">*</span>{{end}} {{.Name}} [{{.Type}} {{- if eq .Type "list" }} of {{.ArrayItemsType}}s{{- end -}}] {{ if .IsSensitive }}(<a href="#special_terms_definitions_sensitive_property" target="_self">sensitive</a>){{end}} - {{.Description}}
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
{{- end -}}

<h2 id="provider_datasources">Data Sources (using resource id)</h2>

{{range .DataSourceInstances -}}
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
{{range .DataSources -}}
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
{{end}} {{/* END range .DataSources */}}`
