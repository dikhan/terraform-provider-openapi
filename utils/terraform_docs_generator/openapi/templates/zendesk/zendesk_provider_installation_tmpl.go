package zendesk

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
</pre>`
