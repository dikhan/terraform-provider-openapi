package openapi

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
)

const dataSourceFilterPropertyName = "filter"
const dataSourceFilterSchemaNamePropertyName = "name"
const dataSourceFilterSchemaValuesPropertyName = "values"

type dataSourceFactory struct {
	openAPIResource SpecResource
}

type filters []filter
type filter struct {
	name  string
	value string
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
		dataSourceFilterPropertyName: d.dataSourceFiltersSchema(),
	}
}

func (d dataSourceFactory) dataSourceFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				dataSourceFilterSchemaNamePropertyName: {
					Type:     schema.TypeString,
					Required: true,
				},
				dataSourceFilterSchemaValuesPropertyName: {
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

	filters, err := d.validateInput(data)
	if err != nil {
		return err
	}

	parentIDs, resourcePath, err := getParentIDsAndResourcePath(d.openAPIResource, data)
	if err != nil {
		return err
	}

	responsePayload := []map[string]interface{}{}
	resp, err := openAPIClient.List(d.openAPIResource, &responsePayload, parentIDs...)
	if err != nil {
		return err
	}

	if err := checkHTTPStatusCode(d.openAPIResource, resp, []int{http.StatusOK}); err != nil {
		return fmt.Errorf("[data source='%s'] GET %s failed: %s", d.openAPIResource.getResourceName(), resourcePath, err)
	}

	fmt.Println(filters)
	//for _, payloadItem := range responsePayload {
	//
	//}

	// TODO: make use of responsePayload to filter out results

	// TODO: If there are multiple matches after applying the filters return an error
	// TODO: If there are no matches after applying the filters return an error

	// TODO: update the state data object with the filtered result data
	// d.updateStateWithPayloadData(remoteData, data)

	return nil
}

func (d dataSourceFactory) validateInput(data *schema.ResourceData) (filters, error) {
	filters := filters{}
	inputFilters := data.Get(dataSourceFilterPropertyName)
	for _, inputFilter := range inputFilters.(*schema.Set).List() {
		f := inputFilter.(map[string]interface{})
		filterPropertyName := f[dataSourceFilterSchemaNamePropertyName].(string)
		s, err := d.openAPIResource.getResourceSchema()
		if err != nil {
			return nil, err
		}

		specSchemaDefinitionProperty, err := s.getProperty(filterPropertyName)
		if err != nil {
			return nil, fmt.Errorf("filter name does not match any of the schema properties: %s", err)
		}

		if !specSchemaDefinitionProperty.isPrimitiveProperty() {
			return nil, fmt.Errorf("property not supported as as filter: %s", specSchemaDefinitionProperty.getTerraformCompliantPropertyName())
		}

		filterValue := f[dataSourceFilterSchemaValuesPropertyName].([]interface{})
		if len(filterValue) > 1 {
			return nil, fmt.Errorf("filters for primitive properties can not have more than one value in the values field")
		}
		filters = append(filters, filter{filterPropertyName, filterValue[0].(string)})
	}
	return filters, nil
}
