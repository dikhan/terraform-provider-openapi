package zendesk

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
