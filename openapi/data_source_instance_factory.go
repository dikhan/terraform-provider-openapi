package openapi

import (
	"fmt"

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
		//Read:   d.read,
	}, nil
}

func (d dataSourceInstanceFactory) createTerraformDataSourceInstanceSchema() (map[string]*schema.Schema, error) {
	specSchema, err := d.openAPIResource.getResourceSchema()
	if err != nil {
		return nil, err
	}
	dataSourceSchema, err := specSchema.createDataSourceSchema()
	dataSourceSchema[dataSourceInstanceIDProperty] = d.dataSourceInstanceSchema()
	return dataSourceSchema, nil
}

func (d dataSourceInstanceFactory) dataSourceInstanceSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
}
