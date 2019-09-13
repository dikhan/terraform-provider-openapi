package openapi

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
)

type dataSourceFactory struct {
	openAPIResource SpecResource
}

func newDataSourceFactory(openAPIResource SpecResource) dataSourceFactory {
	return dataSourceFactory{
		openAPIResource: openAPIResource,
	}
}

func (d dataSourceFactory) createTerraformDataSource() (*schema.Resource, error) {
	s := d.createTerraformDataSourceSchema()
	return &schema.Resource{
		Schema: s,
		Read:   d.read,
	}, nil
}

func (d dataSourceFactory) createTerraformDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"filter": d.dataSourceFiltersSchema(),
	}
}

func (d dataSourceFactory) dataSourceFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"values": {
					Type:     schema.TypeList,
					Required: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func (d dataSourceFactory) read(data *schema.ResourceData, i interface{}) error {
	openAPIClient := i.(ClientOpenAPI)

	if err := d.validateInput(data); err != nil {
		return err
	}

	parentIDs, resourcePath, err := getParentIDsAndResourcePath(d.openAPIResource, data)
	if err != nil {
		return err
	}

	responsePayload := map[string]interface{}{}
	resp, err := openAPIClient.List(d.openAPIResource, &responsePayload, parentIDs...)
	if err != nil {
		return err
	}

	if err := checkHTTPStatusCode(d.openAPIResource, resp, []int{http.StatusOK}); err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("[resource='%s'] GET %s failed: %s", d.openAPIResource.getResourceName(), resourcePath, err)
	}

	// TODO: make use of responsePayload to filter out results

	// TODO: If there are multiple matches after applying the filters return an error
	// TODO: If there are no matches after applying the filters return an error

	// TODO: update the state data object with the filtered result data
	// d.updateStateWithPayloadData(remoteData, data)

	return nil
}

func (d dataSourceFactory) validateInput(data *schema.ResourceData) error {
	filters := data.Get("filter")
	for _, filter := range filters.(*schema.Set).List() {
		f := filter.(map[string]interface{})
		filterPropertyName := f["name"].(string)
		s, err := d.openAPIResource.getResourceSchema()
		if err != nil {
			return err
		}

		// This validates that the property from the filter exists one of the properties from the data source schema
		_, err = s.getProperty(filterPropertyName)
		if err != nil {
			return err
		}

		// TODO: validate that the filter values contain just one element for specSchemaDefinitionProperty of type primitive. error out otherwise

	}
	return nil
}
