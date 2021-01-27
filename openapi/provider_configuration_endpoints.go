package openapi

import (
	"bytes"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi/openapiutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"hash/crc32"
)

type providerConfigurationEndPoints struct {
	resourceNames []string
}

// endpointsSchema returns a schema for the provider's endpoint property
func (p *providerConfigurationEndPoints) endpointsSchema() *schema.Schema {
	if p.resourceNames != nil && len(p.resourceNames) > 0 {
		endpoints := map[string]*schema.Schema{}
		for _, name := range p.resourceNames {
			endpoints[name] = &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: p.endpointsValidateFunc(),
				Description:  "Use this to override the resource endpoint URL (the default one or the one constructed from the `region`).\n",
			}
		}
		return &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: endpoints,
			},
			Set: p.endpointsToHash(p.resourceNames),
		}
	}
	return nil
}

func (p *providerConfigurationEndPoints) endpointsValidateFunc() schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warns []string, errs []error) {
		userValue := value.(string)
		if openapiutils.IsValidHost(userValue) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("property '%s' value '%s' is not valid, please make sure the value is a valid FQDN or well formed IP (the host may contain non standard ports too followed by a colon - e,g: www.api.com:8080). The protocol used when performing the API call will be populated based on the swagger specification", key, userValue)}
	}
}

// endpointsToHash calculates the unique ID used to store the endpoints element in a hash.
func (p *providerConfigurationEndPoints) endpointsToHash(resources []string) schema.SchemaSetFunc {
	return func(v interface{}) int {
		var buf bytes.Buffer
		m := v.(map[string]interface{})
		for _, name := range resources {
			buf.WriteString(fmt.Sprintf("%s-", m[name].(string)))
		}
		// Terraform SDK 2.0 upgrade: https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#removal-of-helper-hashcode-package
		return int(crc32.ChecksumIEEE(buf.Bytes()))
	}
}

// configureEndpoints creates a map containing one endpoint per resource exposed by the provider and maps the values
// with the ones provided by the user (if present)
func (p *providerConfigurationEndPoints) configureEndpoints(data *schema.ResourceData) map[string]string {
	if p.resourceNames != nil && len(p.resourceNames) > 0 {
		providerConfigEndPoints := map[string]string{}
		if data.Get(providerPropertyEndPoints) != nil {
			endpointsSet := data.Get(providerPropertyEndPoints).(*schema.Set)
			for _, endpointsSetI := range endpointsSet.List() {
				endpoints := endpointsSetI.(map[string]interface{})
				for _, resource := range p.resourceNames {
					providerConfigEndPoints[resource] = endpoints[resource].(string)
				}
			}
			return providerConfigEndPoints
		}
	}
	return nil
}
