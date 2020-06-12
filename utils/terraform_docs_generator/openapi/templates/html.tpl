<p dir="ltr">
  This guide lists the configuration for '{{.ProviderName}}' Terraform provider
  resources that can be managed using
  <a href="https://www.hashicorp.com/blog/announcing-terraform-0-12/" target="_self">Terraform v0.12</a>.&nbsp;
</p>
<ul>
  <li>
    <a href="#provider_installation" target="_self">Provider Installation</a>
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