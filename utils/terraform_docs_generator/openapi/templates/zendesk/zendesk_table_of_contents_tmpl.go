package zendesk

// TableOfContentsTmpl contains the template used to render the table of contents as HTML formatted for Zendesk
var TableOfContentsTmpl = `<p dir="ltr">
  This guide lists the configuration for '{{.ProviderName}}' Terraform provider
  resources that can be managed using
  <a href="https://www.hashicorp.com/blog/announcing-terraform-0-12/" target="_self">Terraform v0.12</a>.&nbsp;
</p>
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
