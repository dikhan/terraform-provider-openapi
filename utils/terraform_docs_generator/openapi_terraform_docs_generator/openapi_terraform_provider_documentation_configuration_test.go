package openapi_terraform_docs_generator

import (
	"bytes"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi_terraform_docs_generator/templates/zendesk"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestProviderConfiguration_RenderZendesk(t *testing.T) {
	pc := ProviderConfiguration{
		ProviderName: "openapi",
		Regions:      []string{"rst1"},
		ConfigProperties: []Property{
			{
				Name:     "token",
				Required: true,
				Type:     "string",
			},
		},
		ExampleUsage: nil,
		ArgumentsReference: ArgumentsReference{
			Notes: []string{"Note: some special notes..."},
		},
	}
	var buf bytes.Buffer
	expectedHTML := `<h2 id="provider_configuration">Provider Configuration</h2>
<h4 id="provider_configuration_example_usage" dir="ltr">Example Usage</h4>
    <pre>
<span>provider </span><span>"openapi" </span>{
<span>  token  </span>= <span>"..."</span>
<span>}</span>
</pre>

    <p>Using the default region (rst1):</p>
    <pre>
<span>provider </span><span>"openapi" </span>{
<span>  # Resources using this default provider will be created in the 'rst1' region<br>  ...<br></span>}
    </pre>
    

    <h4 id="provider_configuration_arguments_reference" dir="ltr">Arguments Reference</h4>
    <p dir="ltr">The following arguments are supported:</p>
    <ul dir="ltr">
        <li><span>token [string] - (Required) .</span></li></li>
      <li>
          region [string] - (Optional) The region location to be used&nbsp;([rst1]). If region isn't specified, the default is "rst1".
      </li>
    
    </ul>`
	err := pc.Render(&buf, zendesk.ProviderConfigurationTmpl)
	assert.Equal(t, expectedHTML, strings.Trim(buf.String(), "\n"))
	assert.Nil(t, err)
}
