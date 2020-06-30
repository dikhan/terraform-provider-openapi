package zendesk

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
        <p>Using a specific region ({{index .ProviderConfiguration.Regions 1}}):</p>
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
