package zendesk

// ProviderResourcesTmpl contains the template used to render the TerraformProviderDocumentation.ProviderResources struct as HTML formatted for Zendesk
var ProviderResourcesTmpl = `{{define "resource_example"}}
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

{{- define "resource_argument_reference" -}}
    {{- $required := "Optional" -}}
    {{- if .Required -}}
        {{- $required = "Required" -}}
    {{end}}
	{{- if or .Required (and (not .Required) (not .Computed)) .IsOptionalComputed -}}
    <li>{{if eq .Type "object"}}<span class="wysiwyg-color-red">*</span>{{end}} {{.Name}} [{{.Type}} {{- if eq .Type "list" }} of {{.ArrayItemsType}}s{{- end -}}] {{ if .IsSensitive }}(<a href="#special_terms_definitions_sensitive_property" target="_self">sensitive</a>){{end}} - ({{$required}}) {{.Description}}
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

<h2 id="provider_resources">Provider Resources</h2>

{{range .Resources -}}
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
<pre dir="ltr">$ terraform import {{.Name}}.my_{{.Name}} {{$resource.BuildImportIDsExample}}</pre>
<p dir="ltr">
    <strong>Note</strong>: In order for the import to work, the '{{$.ProviderName}}' terraform
    provider must be&nbsp;<a href="#provider_installation" target="_self">properly installed</a>. Read more about Terraform import usage&nbsp;<a href="https://www.terraform.io/docs/import/usage.html" target="_blank" rel="noopener noreferrer">here</a>.
</p>

{{end}} {{/* END range .Resources */}}`
