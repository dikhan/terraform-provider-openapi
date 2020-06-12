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

{{if .ProviderConfiguration.Regions }}
    <h2 id="provider_configuration">Provider Configuration</h2>
    <h4 id="provider_configuration_example_usage" dir="ltr">Example Usage</h4>

    <p>Using the default region ({{index .ProviderConfiguration.Regions 0}}):</p>
    <pre><span>provider </span><span>"{{.ProviderName}}" </span>{<br><span>  # Resources using this default provider will be created in the '{{index .ProviderConfiguration.Regions 0}}' region<br></span>}<br><br>resource<span>"{{.ProviderName}}_resource" "my_resource" {<br>  ...<br>}<br></span></pre>

    {{ if gt (len .ProviderConfiguration.Regions) 1 }}
            <p>Using a specific region ({{index .ProviderConfiguration.Regions 1}}):</p>
            <pre><span>provider </span><span>"{{.ProviderName}}" </span>{<br>  <span>alias  </span>= <span>"{{index .ProviderConfiguration.Regions 1}}"<br></span><span>  </span><span>region </span>= <span>"{{index .ProviderConfiguration.Regions 1}}"<br></span>}<br><br>resource<span>"{{.ProviderName}}_resource" "my_resource" {<br>  provider = "{{.ProviderName}}.{{index .ProviderConfiguration.Regions 1}}"<br>  ...<br>}</span></pre>
    {{ end }}

    <h4 id="provider_configuration_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
      <li>
        region - (Optional) The region location to be used&nbsp;({{.ProviderConfiguration.Regions}}). If region isn't specified, the default is "{{index .ProviderConfiguration.Regions 0}}".
      </li>
    </ul>
{{end}}
