package main

import (
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi"
	"log"
	"os"
)

func main() {

	terraformProviderDocumentation := openapi.TerraformProviderDocumentation{
		ProviderName: "openapi",
		ProviderInstallation: openapi.ProviderInstallation{
			Example: "$ export PROVIDER_NAME=openapi && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME<br>" +
				"[INFO] Downloading https://github.com/dikhan/terraform-provider-openapi/releases/download/v0.29.4/terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz in temporally folder /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh...<br>" +
				"[INFO] Extracting terraform-provider-openapi from terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz...<br>" +
				"[INFO] Cleaning up tmp dir created for installation purposes: /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh<br>" +
				"[INFO] Terraform provider 'terraform-provider-openapi_v0.29.4' successfully installed at: '~/.terraform.d/plugins'!",
			Other:        "You can then start running the Terraform provider:",
			OtherCommand: "$ export OTF_VAR_openapi_PLUGIN_CONFIGURATION_FILE=\"https://api.service.com/openapi.yaml\"<br>",
		},
		ProviderConfiguration: openapi.ProviderConfiguration{
			Regions: []string{"sea", "rst"},
		},
		ProviderResources: openapi.ProviderResources{
			Resources: []openapi.Resource{
				openapi.Resource{
					Name:        "openapi_resource1",
					Description: "Allows management of resource1",
					//ExampleUsage: []openapi.ExampleUsage{openapi.ExampleUsage{"example usage"}},
					ArgumentsReference: openapi.ArgumentsReference{Notes: []string{`This is internally represented as an array of one elem due to 
<a href="https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737" target="_self">Terraform SDK's limitation for supporting complex object types</a>. 
Please index on the first elem of the array to&nbsp;reference&nbsp;the object values&nbsp; (eg: openapi_resource1<span>.my_resource1.<strong>prop_object[0]</strong>.<strong>obj_prop_object[0]</strong>.prop_id</span>)`}},
					Properties: []openapi.Property{
						openapi.Property{
							Name:        "prop_string",
							Type:        "string",
							Required:    true,
							Description: "prop1 description",
						},
						openapi.Property{
							Name:        "prop_int",
							Type:        "integer",
							Required:    true,
							Description: "prop_int description",
						},
						openapi.Property{
							Name:        "prop_bool",
							Type:        "boolean",
							Required:    true,
							Description: "prop_bool description",
						},
						openapi.Property{
							Name:        "prop_float",
							Type:        "float",
							Required:    true,
							Description: "prop_float description",
						},
						openapi.Property{
							Name:           "prop_array_string",
							Type:           "array",
							ArrayItemsType: "string",
							Required:       true,
							Description:    "prop_float description",
						},
						openapi.Property{
							Name:           "prop_array_object",
							Type:           "array",
							ArrayItemsType: "object",
							Required:       true,
							Description:    "prop_array_object description",
							Schema: []openapi.Property{
								openapi.Property{
									Name:        "obj_prop_string",
									Type:        "string",
									Required:    false,
									Description: "obj_prop_string description",
								},
								openapi.Property{
									Name:           "obj_prop_array_string",
									Type:           "array",
									ArrayItemsType: "string",
									Required:       false,
									Description:    "obj_prop_array_string description",
								},
							},
						},
						openapi.Property{
							Name:        "prop_object",
							Type:        "object",
							Required:    true,
							Description: "prop_object description",
							Schema: []openapi.Property{
								openapi.Property{
									Name:        "obj_prop_string",
									Type:        "string",
									Required:    true,
									Description: "obj_prop_string description",
								},
								openapi.Property{
									Name:           "obj_prop_array_string",
									Type:           "array",
									ArrayItemsType: "string",
									Required:       false,
									Description:    "obj_prop_array_string description",
								},
								openapi.Property{
									Name:        "obj_prop_object",
									Type:        "object",
									Required:    true,
									Description: "obj_prop_object description",
									Schema: []openapi.Property{
										openapi.Property{
											Name:        "obj_obj_prop_string",
											Type:        "string",
											Required:    true,
											Description: "obj_obj_prop_string description",
										},
									},
								},
							},
						},
						openapi.Property{
							Name:        "prop_string_computed",
							Type:        "string",
							Required:    false,
							Computed:    true,
							Description: "prop1 description",
						},
					},
					//AttributesReference: openapi.AttributesReference{},
					//Import:              openapi.Import{},
				},
			},
		},
		DataSources: openapi.DataSources{
			DataSources: []openapi.DataSource{
				openapi.DataSource{
					Name:        "openapi_resource1",
					Description: "The resource1 instance data source allows you to retrieve an already existing resource1 using filters",
					Properties: []openapi.Property{
						openapi.Property{
							Name:        "filter",
							Type:        "object",
							Required:    true,
							Description: "The filter object property enables users to specify the criteria to filter the results",
							Schema: []openapi.Property{
								openapi.Property{
									Name:        "name",
									Type:        "string",
									Required:    true,
									Description: "The value should match one of the properties to filter by. Only value allowed for now is 'name'",
								},
								openapi.Property{
									Name:           "values",
									Type:           "array",
									ArrayItemsType: "string",
									Required:       true,
									Description:    "Values to filter out the results. For primitive properties the array should contain only one item containing the value to filter by.",
								},
							},
						},
						openapi.Property{
							Name:        "computed_string",
							Type:        "string",
							Required:    false,
							Computed:    true,
							Description: "computed string description",
						},
						openapi.Property{
							Name:        "computed_int",
							Type:        "integer",
							Required:    false,
							Computed:    true,
							Description: "computed int description",
						},
					},
				},
			},
			DataSourceInstances: []openapi.DataSource{
				openapi.DataSource{
					Name:        "openapi_resource1_instance",
					Description: "The resource1 instance data source allows you to retrieve an already existing resource1",
					Properties: []openapi.Property{
						openapi.Property{
							Name:        "id",
							Type:        "string",
							Required:    true,
							Description: "prop1 description",
						},
						openapi.Property{
							Name:        "parent_id",
							Type:        "string",
							Required:    true,
							Description: "prop1 description",
						},
						openapi.Property{
							Name:        "optional_string",
							Type:        "string",
							Required:    false,
							Description: "optional prop description",
						},
						openapi.Property{
							Name:        "computed_string",
							Type:        "string",
							Required:    false,
							Computed:    true,
							Description: "computed prop description",
						},
					},
				},
			},
		},
	}

	f, err := os.Create("./utils/terraform_docs_generator/openapi/templates/zendesk_output.html")
	if err != nil {
		log.Fatal(err)
	}

	err = terraformProviderDocumentation.RenderZendeskHTML(f)
	if err != nil {
		log.Fatal(err)
	}

}
