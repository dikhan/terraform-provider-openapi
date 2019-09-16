package openapi

import (
	"fmt"
	"net/http"
	"strconv"

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
	s, err := d.createTerraformDataSourceSchema()
	if err != nil {
		return nil, err
	}
	return &schema.Resource{
		Schema: s,
		Read:   d.read,
	}, nil
}

func (d dataSourceFactory) createTerraformDataSourceSchema() (map[string]*schema.Schema, error) {
	specSchema, err := d.openAPIResource.getResourceSchema()
	if err != nil {
		return nil, err
	}
	dataSourceSchema, err := specSchema.createDataSourceSchema()
	dataSourceSchema[dataSourceFilterPropertyName] = d.dataSourceFiltersSchema()
	return dataSourceSchema, nil
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

	var filteredResults []map[string]interface{}
	for _, payloadItem := range responsePayload {
		match, err := d.filterMatch(filters, payloadItem)
		if err != nil {
			return err
		}
		if match {
			filteredResults = append(filteredResults, payloadItem)
		}
	}

	if len(filteredResults) == 0 {
		return fmt.Errorf("your query returned no results. Please change your search criteria and try again")
	}

	if len(filteredResults) > 1 {
		return fmt.Errorf("your query returned contains more than one result. Please change your search criteria to make it more specific")
	}

	return updateStateWithPayloadData(d.openAPIResource, filteredResults[0], data)
}

// TODO: add tests
func (d dataSourceFactory) filterMatch(filters filters, payloadItem map[string]interface{}) (bool, error) {
	specSchemaDefinition, err := d.openAPIResource.getResourceSchema()
	if err != nil {
		return false, err
	}
	for _, filter := range filters {
		if val, exists := payloadItem[filter.name]; exists {
			schemaProperty, _ := specSchemaDefinition.getProperty(filter.name)
			var value string
			switch schemaProperty.Type {
			case typeInt:
				value = strconv.Itoa(val.(int))
			case typeFloat:
				value = fmt.Sprintf("%g", val.(float64))
			case typeBool:
				value = strconv.FormatBool(val.(bool))
			default:
				value = val.(string)
			}
			if value == filter.value {
				continue
			}
		}
		return false, nil
	}
	return true, nil
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
