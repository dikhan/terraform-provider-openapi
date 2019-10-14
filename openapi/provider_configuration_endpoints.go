package openapi

import (
	"bytes"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type providerConfigurationEndPoints struct {
	specAnalyser SpecAnalyser
}

func newProviderConfigurationEndPoints(specAnalyser SpecAnalyser) (*providerConfigurationEndPoints, error) {
	if specAnalyser == nil {
		return nil, fmt.Errorf("specAnalyser must be provided to create a providerConfigurationEndPoints struct")
	}
	return &providerConfigurationEndPoints{specAnalyser}, nil
}

// getResourceNames returns the resources exposed by the provider. The list of resources names returned will then be
// used to create the provider's endpoint schema property as well as to configure the endpoints values with the data
// provided bu the user
func (p *providerConfigurationEndPoints) getResourceNames() ([]string, error) {
	openAPIResources, err := p.specAnalyser.GetTerraformCompliantResources()
	if err != nil {
		return nil, err
	}
	resourceNames := []string{}
	for _, openAPIResource := range openAPIResources {
		resourceNames = append(resourceNames, openAPIResource.getResourceName())

	}
	return resourceNames, nil
}

// endpointsSchema returns a schema for the provider's endpoint property
func (p *providerConfigurationEndPoints) endpointsSchema() (*schema.Schema, error) {
	resourceNames, err := p.getResourceNames()
	if err != nil {
		return nil, err
	}
	if len(resourceNames) > 0 {
		endpoints := map[string]*schema.Schema{}
		for _, name := range resourceNames {
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
			Set: p.endpointsToHash(resourceNames),
		}, nil
	}
	return nil, nil
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
		return hashcode.String(buf.String())
	}
}

// configureEndpoints creates a map containing one endpoint per resource exposed by the provider and maps the values
// with the ones provided by the user (if present)
func (p *providerConfigurationEndPoints) configureEndpoints(data *schema.ResourceData) (map[string]string, error) {
	providerConfigEndPoints := map[string]string{}
	if data.Get(providerPropertyEndPoints) != nil {
		endpointsSet := data.Get(providerPropertyEndPoints).(*schema.Set)
		for _, endpointsSetI := range endpointsSet.List() {
			endpoints := endpointsSetI.(map[string]interface{})
			resourceNames, err := p.getResourceNames()
			if err != nil {
				return nil, err
			}
			for _, resource := range resourceNames {
				providerConfigEndPoints[resource] = endpoints[resource].(string)
			}
		}
	}
	return providerConfigEndPoints, nil
}
