package openapi

import (
	"bytes"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/templates/zendesk"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestProviderInstallation_RenderZendesk(t *testing.T) {
	pi := ProviderInstallation{
		ProviderName: "openapi",
		Example:      "➜ ~ This is an example",
		Other:        "Some more info about the installation",
		OtherCommand: "➜ ~ init_command do_something",
	}
	var buf bytes.Buffer
	expectedHTML := `<h2 id="provider_installation">Provider Installation</h2>
<p>
  In order to provision 'openapi' Terraform resources, you need to first install the 'openapi'
  Terraform plugin by running&nbsp;the following command (you must be running Terraform &gt;= 0.12):
</p>
<pre>➜ ~ This is an example</pre>
<p>
  <span>Some more info about the installation</span>
</p>
<pre dir="ltr">➜ ~ init_command do_something
➜ ~ terraform init &amp;&amp; terraform plan
</pre>`
	err := pi.Render(&buf, zendesk.ProviderInstallationTmpl)
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
	assert.Nil(t, err)
}
