package openapi

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
)

const dataSourceInstanceIDProperty = "id"

type dataSourceInstanceFactory struct {
	openAPIResource SpecResource
}

func newDataSourceInstanceFactory(openAPIResource SpecResource) dataSourceInstanceFactory {
	return dataSourceInstanceFactory{
		openAPIResource: openAPIResource,
	}
}

func (d dataSourceInstanceFactory) getDataSourceInstanceName() string {
	return fmt.Sprintf("%s_instance", d.openAPIResource.getResourceName())
}

func (d dataSourceInstanceFactory) createTerraformInstanceDataSource() (*schema.Resource, error) {
	s, err := d.createTerraformDataSourceInstanceSchema()
	if err != nil {
		return nil, err
	}
	return &schema.Resource{
		Schema: s,
		Read:   d.read,
	}, nil
}

func (d dataSourceInstanceFactory) createTerraformDataSourceInstanceSchema() (map[string]*schema.Schema, error) {
	specSchema, err := d.openAPIResource.getResourceSchema()
	if err != nil {
		return nil, err
	}
	dataSourceSchema, err := specSchema.createDataSourceSchema()
	if err != nil {
		return nil, err
	}
	dataSourceSchema[dataSourceInstanceIDProperty] = d.dataSourceInstanceSchema()
	return dataSourceSchema, nil
}

func (d dataSourceInstanceFactory) dataSourceInstanceSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
}

func (d dataSourceInstanceFactory) read(data *schema.ResourceData, i interface{}) error {
	openAPIClient := i.(ClientOpenAPI)
	parentIDs, resourcePath, err := getParentIDsAndResourcePath(d.openAPIResource, data)
	if err != nil {
		return err
	}
	id := data.Get(dataSourceInstanceIDProperty)
	if id == nil || id == "" {
		return fmt.Errorf("data source 'id' property value must be populated")
	}
	responsePayload := map[string]interface{}{}
	resp, err := openAPIClient.Get(d.openAPIResource, id.(string), &responsePayload, parentIDs...)
	if err != nil {
		return err
	}
	if err := checkHTTPStatusCode(d.openAPIResource, resp, []int{http.StatusOK}); err != nil {
		return fmt.Errorf("[data source instance='%s'] GET %s failed: %s", d.openAPIResource.getResourceName(), resourcePath, err)
	}
	err = setStateID(d.openAPIResource, data, responsePayload)
	if err != nil {
		return err
	}
	return updateStateWithPayloadData(d.openAPIResource, responsePayload, data)
}
