package openapi

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	return fmt.Sprintf("%s_instance", d.openAPIResource.GetResourceName())
}

func (d dataSourceInstanceFactory) createTerraformInstanceDataSource() (*schema.Resource, error) {
	s, err := d.createTerraformDataSourceInstanceSchema()
	if err != nil {
		return nil, err
	}
	return &schema.Resource{
		Schema:      s,
		ReadContext: crudWithContext(d.read, schema.TimeoutRead, d.openAPIResource.GetResourceName()),
	}, nil
}

func (d dataSourceInstanceFactory) createTerraformDataSourceInstanceSchema() (map[string]*schema.Schema, error) {
	specSchema, err := d.openAPIResource.GetResourceSchema()
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

	if d.openAPIResource == nil {
		return fmt.Errorf("missing openAPI resource configuration")
	}
	resourceName := d.getDataSourceInstanceName()

	submitTelemetryMetricDataSource(openAPIClient, TelemetryResourceOperationRead, resourceName)

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
		return fmt.Errorf("[data source instance='%s'] GET %s failed: %s", resourceName, resourcePath, err)
	}
	err = setStateID(d.openAPIResource, data, responsePayload)
	if err != nil {
		return err
	}
	return dataSourceUpdateStateWithPayloadData(d.openAPIResource, responsePayload, data)
}
