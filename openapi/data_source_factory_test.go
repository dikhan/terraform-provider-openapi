package openapi

import (
	"github.com/hashicorp/terraform/helper/schema"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestValidateInput(t *testing.T) {

	// TODO: expand this test and add more coverage, it's missing validations checks
	// 	created this set of tests to enable better understanding about the set up and further extension, the code will
	//  have to be refactored at some point removing boiler plate etc

	Convey("Given a data source factory and a resourceLocalData populated with a correct filter", t, func() {
		// This is representing the corresponding schema for a valid swagger model definition
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: specSchemaDefinitionProperties{
						&specSchemaDefinitionProperty{
							Name:     "id",
							ReadOnly: true,
						},
						&specSchemaDefinitionProperty{
							Name: "label",
						},
					},
				},
			},
		}
		resourceSchema := dataSourceFactory.createTerraformDataSourceSchema()
		// This is the input we would expect from the user
		resourceDataMap := map[string]interface{}{
			"filter": []map[string]interface{}{
				{
					"name":   "label",
					"values": []string{"label_to_fetch"},
				},
			},
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		Convey("When create is called with resource data", func() {
			err := dataSourceFactory.validateInput(resourceLocalData)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a data source factory and a resourceLocalData populated with an incorrect filter containing a property that does not match any of the schema definition", t, func() {
		// This is representing the corresponding schema for a valid swagger model definition
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: specSchemaDefinitionProperties{
						&specSchemaDefinitionProperty{
							Name:     "id",
							ReadOnly: true,
						},
						&specSchemaDefinitionProperty{
							Name: "label",
						},
					},
				},
			},
		}
		resourceSchema := dataSourceFactory.createTerraformDataSourceSchema()
		// This is the input we would expect from the user
		resourceDataMap := map[string]interface{}{
			"filter": []map[string]interface{}{
				{
					"name":   "non_matching_property_name",
					"values": []string{"label_to_fetch"},
				},
			},
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		Convey("When create is called with resource data", func() {
			err := dataSourceFactory.validateInput(resourceLocalData)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a data source factory and a resourceLocalData populated with an incorrect filter containing multiple values for a primitive property", t, func() {
		// This is representing the corresponding schema for a valid swagger model definition
		dataSourceFactory := dataSourceFactory{
			openAPIResource: &specStubResource{
				schemaDefinition: &specSchemaDefinition{
					Properties: specSchemaDefinitionProperties{
						&specSchemaDefinitionProperty{
							Name:     "id",
							ReadOnly: true,
						},
						&specSchemaDefinitionProperty{
							Name: "label",
						},
					},
				},
			},
		}
		resourceSchema := dataSourceFactory.createTerraformDataSourceSchema()
		// This is the input we would expect from the user
		resourceDataMap := map[string]interface{}{
			"filter": []map[string]interface{}{
				{
					"name":   "label",
					"values": []string{"value1", "value2"},
				},
			},
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		Convey("When create is called with resource data", func() {
			err := dataSourceFactory.validateInput(resourceLocalData)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
