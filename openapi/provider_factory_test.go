package openapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewProviderFactory(t *testing.T) {
	Convey("Given a provider name, an analyser and the service config", t, func() {
		providerName := "provider"
		analyser := &specAnalyserStub{}
		serviceConfig := &ServiceConfigV1{}
		Convey("When newProviderFactory is called ", func() {
			p, err := newProviderFactory(providerName, analyser, serviceConfig)
			Convey("Then the provider returned should NOT be nil and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(p, ShouldNotBeNil)
			})
		})
	})
	Convey("Given a provider name that is empty, an analyser and the service config", t, func() {
		providerName := ""
		analyser := &specAnalyserStub{}
		serviceConfig := &ServiceConfigV1{}
		Convey("When newProviderFactory is called ", func() {
			_, err := newProviderFactory(providerName, analyser, serviceConfig)
			Convey("Then the provider returned should NOT be nil and the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "provider name not specified")
			})
		})
	})
	Convey("Given a provider name that is not terraform compliant, an analyser and the service config", t, func() {
		providerName := "someNonTerraformCompliantName"
		analyser := &specAnalyserStub{}
		serviceConfig := &ServiceConfigV1{}
		Convey("When newProviderFactory is called ", func() {
			_, err := newProviderFactory(providerName, analyser, serviceConfig)
			Convey("Then the error returned should be nil", func() {
				So(err.Error(), ShouldEqual, "provider name 'someNonTerraformCompliantName' not terraform name compliant, please consider renaming provider to 'some_non_terraform_compliant_name'")
			})
		})
	})
	Convey("Given a provider name, a nil analyser and the service config", t, func() {
		providerName := "provider"
		Convey("When newProviderFactory is called ", func() {
			_, err := newProviderFactory(providerName, nil, nil)
			Convey("Then the error returned should be nil", func() {
				So(err.Error(), ShouldEqual, "provider missing an OpenAPI Spec Analyser")
			})
		})
	})

	Convey("Given a provider name, an analyser and a nil service config", t, func() {
		providerName := "provider"
		analyser := &specAnalyserStub{}
		Convey("When newProviderFactory is called ", func() {
			_, err := newProviderFactory(providerName, analyser, nil)
			Convey("Then the error returned should be nil", func() {
				So(err.Error(), ShouldEqual, "provider missing the service configuration")
			})
		})
	})
}

func TestGetResourceNames(t *testing.T) {
	Convey("Given a provider factory", t, func() {
		p := providerFactory{
			name: "provider",
		}
		Convey("When getResourceNames is called with a map of resources", func() {
			resources := map[string]*schema.Resource{
				"provider_resource_name_v1": {},
			}
			resourceNames := p.getResourceNames(resources)
			Convey("Then the list should contain the expected resources", func() {
				So(resourceNames, ShouldContain, "resource_name_v1")
			})
		})
	})
}

func TestCreateProvider(t *testing.T) {
	Convey("Given a provider factory", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{newSpecStubResource("resource_v1", "/v1/resource", false, &SpecSchemaDefinition{})},
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name: headerProperty.Name,
					},
				},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
				backendConfiguration: &specStubBackendConfiguration{},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createProvider is called ", func() {
			p, err := p.createProvider()
			Convey("Then the provider returned should be configured as expected and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(p, ShouldNotBeNil)
				So(p.ResourcesMap, ShouldContainKey, "provider_resource_v1")
				So(p.DataSourcesMap, ShouldContainKey, "provider_resource_v1_instance")
				So(p.Schema[apiKeyAuthProperty.Name], ShouldNotBeNil)
				So(p.Schema[headerProperty.Name], ShouldNotBeNil)
				So(p.Schema["region"], ShouldBeNil)
				So(p.Schema[providerPropertyEndPoints], ShouldNotBeNil)
				So(p.Schema[providerPropertyEndPoints].Elem.(*schema.Resource).Schema, ShouldContainKey, "resource_v1")
			})
		})
	})

	Convey("Given a provider factory with multi-region backend configuration", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name: headerProperty.Name,
					},
				},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
				backendConfiguration: &specStubBackendConfiguration{
					host:    "some-service.${region}.api.com",
					regions: []string{"rst", "dub"},
				},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createProvider is called ", func() {
			p, err := p.createProvider()
			Convey("Then the provider returned should be configured as expected and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(p, ShouldNotBeNil)
				So(p.Schema["region"], ShouldNotBeNil)
				value, err := p.Schema["region"].DefaultFunc()
				So(err, ShouldBeNil)
				So(value.(string), ShouldEqual, "rst")
				warns, errors := p.Schema["region"].ValidateFunc("rst", "")
				So(warns, ShouldBeNil)
				So(errors, ShouldBeNil)
				_, errors = p.Schema["region"].ValidateFunc("nonExisting", "region")
				So(errors, ShouldNotBeNil)
				So(errors[0].Error(), ShouldEqual, "property region value nonExisting is not valid, please make sure the value is one of [rst dub]")
			})
		})
	})

	Convey("Given a provider factory where the specAnalyser has an error", t, func() {
		expectedError := "specAnalyser has an error"
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				error: errors.New(expectedError),
			},
		}
		Convey("When createProvider is called ", func() {
			p, err := p.createProvider()
			Convey("Then the provider returned should be nil and the error returned should be as expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
				So(p, ShouldBeNil)
			})
		})
	})

	Convey("Given a provider factory where the specAnalyser has an error on the backendConfiguration", t, func() {
		expectedError := "backendConfiguration error"
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				backendConfiguration: &specStubBackendConfiguration{
					err: errors.New(expectedError),
				},
			},
		}
		Convey("When createProvider is called ", func() {
			p, err := p.createProvider()
			Convey("Then the provider returned should be nil and the error returned should be as expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
				So(p, ShouldBeNil)
			})
		})
	})

	Convey("Given a provider factory where createTerraformProviderResourceMapAndDataSourceInstanceMap fails", t, func() {
		expectedError := "resource name can not be empty"
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{
					&specStubResource{},
				},
				headers: SpecHeaderParameters{
					SpecHeaderParam{Name: headerProperty.Name},
				},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
				backendConfiguration: &specStubBackendConfiguration{},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createProvider is called ", func() {
			p, err := p.createProvider()
			Convey("Then the provider returned should be nil and the error returned should be as expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
				So(p, ShouldBeNil)
			})
		})
	})

	Convey("Given a provider factory where createTerraformProviderDataSourceMap fails", t, func() {
		expectedError := "resource name can not be empty"
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				dataSources: []SpecResource{
					&specStubResource{
						name:         "",
						path:         "/v1/resource",
						shouldIgnore: false,
						schemaDefinition: &SpecSchemaDefinition{
							Properties: SpecSchemaDefinitionProperties{},
						},
						resourceGetOperation: &specResourceOperation{},
						timeouts:             &specTimeouts{},
					},
				},
				headers: SpecHeaderParameters{
					SpecHeaderParam{Name: headerProperty.Name},
				},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
				backendConfiguration: &specStubBackendConfiguration{},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createProvider is called ", func() {
			p, err := p.createProvider()
			Convey("Then the provider returned should be nil and the error returned should be as expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
				So(p, ShouldBeNil)
			})
		})
	})
}

func TestCreateValidateFunc(t *testing.T) {
	Convey("Given a provider factory", t, func() {
		p := providerFactory{}
		allowedValues := []string{"allowed"}
		Convey("When createValidateFunc is called ", func() {
			validate := p.createValidateFunc(allowedValues)
			Convey("Then the validate function should not be nil and work as expected", func() {
				So(validate, ShouldNotBeNil)
				//  when the function is called with a valid value it should return nil errors
				warns, errors := validate("allowed", "")
				So(warns, ShouldBeNil)
				So(errors, ShouldBeNil)
				//  when the function is called with a valid value it should return nil errors
				warns, errors = validate("notAllowed", "region")
				So(warns, ShouldBeNil)
				So(errors, ShouldNotBeNil)
				So(errors[0].Error(), ShouldEqual, "property region value notAllowed is not valid, please make sure the value is one of [allowed]")
			})
		})
	})
}

func TestCreateTerraformProviderSchema(t *testing.T) {
	Convey("Given a provider factory containing couple properties with commands (that exit with no error)", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")

		serviceConfig := &ServiceConfigStub{
			SchemaConfiguration: []*ServiceSchemaPropertyConfigurationStub{
				{
					SchemaPropertyName:   "apikey_auth",
					DefaultValue:         "someDefaultAuthToken",
					ExecuteCommandCalled: false,
				},
				{
					SchemaPropertyName:   "header_name",
					DefaultValue:         "someDefaultHeaderValue",
					ExecuteCommandCalled: false,
				},
			},
		}
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name: headerProperty.Name,
					},
				},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
			},
			serviceConfiguration: serviceConfig,
		}
		Convey("When createTerraformProviderSchema is called with a backend configuration that is not multi-region", func() {
			backendConfig := &specStubBackendConfiguration{}
			providerSchema, err := p.createTerraformProviderSchema(backendConfig, nil)
			Convey("Then the provider schema for the resource should contain the expected attributes and be configured as expected", func() {
				So(err, ShouldBeNil)
				So(providerSchema, ShouldContainKey, apiKeyAuthProperty.Name)
				So(providerSchema, ShouldContainKey, headerProperty.Name)
				So(providerSchema[apiKeyAuthProperty.Name].DefaultFunc, ShouldNotBeNil)
				So(serviceConfig.SchemaConfiguration[0].ExecuteCommandCalled, ShouldBeTrue)
				So(serviceConfig.SchemaConfiguration[1].ExecuteCommandCalled, ShouldBeTrue)
				// the provider schema 'apikey_auth' property default value should be the expected default
				defaultValue, err := providerSchema[apiKeyAuthProperty.Name].DefaultFunc()
				So(err, ShouldBeNil)
				So(defaultValue, ShouldEqual, "someDefaultAuthToken")
				// the provider schema 'header_name' property default value should be the expected default
				defaultValue, err = providerSchema[headerProperty.Name].DefaultFunc()
				So(err, ShouldBeNil)
				So(defaultValue, ShouldEqual, "someDefaultHeaderValue")
			})
		})
	})

	Convey("Given a provider factory containing a property with command (that exit with error) set up", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		expectedError := "some error executing the command"
		serviceConfig := &ServiceConfigStub{
			SchemaConfiguration: []*ServiceSchemaPropertyConfigurationStub{
				{
					SchemaPropertyName:   "apikey_auth",
					ExecuteCommandCalled: false,
					Err:                  fmt.Errorf(expectedError),
				},
			},
		}
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
			},
			serviceConfiguration: serviceConfig,
		}
		Convey("When createTerraformProviderSchema is called with a backend configuration that is not multi-region", func() {
			backendConfig := &specStubBackendConfiguration{}
			providerSchema, err := p.createTerraformProviderSchema(backendConfig, nil)
			Convey("Then the provider schema for the resource should contain the attribute with default empty value and error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(providerSchema, ShouldContainKey, apiKeyAuthProperty.Name)
				So(providerSchema[apiKeyAuthProperty.Name].Default, ShouldBeEmpty)
				So(serviceConfig.SchemaConfiguration[0].ExecuteCommandCalled, ShouldBeTrue)
			})
		})
	})

	Convey("Given a provider factory that is configured with security definitions that are not all part of the global schemes", t, func() {
		var globalSecurityDefinitionName = "api_key_auth"
		var otherSecurityDefinitionName = "other_security_definition_name"
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(globalSecurityDefinitionName, authorizationHeader),
						newAPIKeyHeaderSecurityDefinition(otherSecurityDefinitionName, "Authorization2"),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							globalSecurityDefinitionName: []string{""},
						},
					}),
				},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createTerraformProviderSchema is called with a backend configuration that is not multi-region", func() {
			backendConfig := &specStubBackendConfiguration{}
			providerSchema, err := p.createTerraformProviderSchema(backendConfig, nil)
			Convey("Then the provider schema for the resource should contain the expected attributes with names automatically converted to be compliant", func() {
				So(err, ShouldBeNil)
				So(providerSchema, ShouldContainKey, globalSecurityDefinitionName)
				So(providerSchema, ShouldContainKey, otherSecurityDefinitionName)
				// the api_key_auth should be required as it's a global scheme
				So(providerSchema[globalSecurityDefinitionName].Required, ShouldBeTrue)
				// the other_security_definition_name should be optional as it's not referred in the global schemes
				So(providerSchema[otherSecurityDefinitionName].Optional, ShouldBeTrue)
				So(providerSchema[globalSecurityDefinitionName].DefaultFunc, ShouldNotBeNil)
				So(providerSchema[otherSecurityDefinitionName].DefaultFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a provider factory", t, func() {
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createTerraformProviderSchema is called with a backend configuration that IS multi-region", func() {
			multiRegionHost := "api.${region}.server.com"
			expectedDefaultRegion := "rst1"
			backendConfig := newStubBackendMultiRegionConfiguration(multiRegionHost, []string{expectedDefaultRegion})
			providerSchema, err := p.createTerraformProviderSchema(backendConfig, nil)
			Convey("And the provider schema for the resource should contain the region property and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(providerSchema, ShouldContainKey, providerPropertyRegion)
				So(providerSchema[providerPropertyRegion].DefaultFunc, ShouldNotBeNil)
				//  the provider schema region property should match the first element of the regions array
				value, err := providerSchema[providerPropertyRegion].DefaultFunc()
				So(err, ShouldBeNil)
				So(value, ShouldEqual, expectedDefaultRegion)
			})
		})
	})

	Convey("Given a provider factory with an spec analyser containing one resource (testing endpoints)", t, func() {
		resourceName := "resource_name_v1"
		resource := newSpecStubResource(resourceName, "", false, nil)
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{resource},
				headers:   SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createTerraformProviderSchema is called a list of resource names", func() {
			backendConfig := &specStubBackendConfiguration{}
			providerSchema, err := p.createTerraformProviderSchema(backendConfig, &providerConfigurationEndPoints{resourceNames: []string{"resourceName"}})
			Convey("And the providerConfigurationEndPoints should be configured with endpoints and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(providerSchema, ShouldContainKey, providerPropertyEndPoints)
				So(providerSchema[providerPropertyEndPoints].Elem.(*schema.Resource).Schema, ShouldContainKey, "resourceName")
			})
		})
	})
	Convey("Given a provider factory (testing endpoints)", t, func() {
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{},
				headers:   SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
			serviceConfiguration: &ServiceConfigStub{},
		}
		Convey("When createTerraformProviderSchema is called with a list of empty resource names", func() {
			backendConfig := &specStubBackendConfiguration{}
			providerSchema, err := p.createTerraformProviderSchema(backendConfig, &providerConfigurationEndPoints{resourceNames: []string{}})
			Convey(fmt.Sprintf("And the provider schema should NOT contain the %s property and the error returned should be nil", providerPropertyEndPoints), func() {
				So(err, ShouldBeNil)
				So(providerSchema, ShouldNotContainKey, providerPropertyEndPoints)
			})
		})
	})
}

func TestConfigureProviderPropertyFromPluginConfig(t *testing.T) {

	Convey("Given a provider factory containing a command that works and also gets the default value from the external source successfully", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		serviceConfig := &ServiceConfigStub{
			SchemaConfiguration: []*ServiceSchemaPropertyConfigurationStub{
				{
					SchemaPropertyName:   "apikey_auth",
					ExecuteCommandCalled: false,
					DefaultValue:         "someDefaultValue",
				},
			},
		}
		p := providerFactory{
			name:                 "provider",
			specAnalyser:         &specAnalyserStub{},
			serviceConfiguration: serviceConfig,
		}
		Convey("When createTerraformProviderSchema is called with an empty provider schema", func() {
			providerSchema := map[string]*schema.Schema{}
			p.configureProviderPropertyFromPluginConfig(providerSchema, "apikey_auth", true)
			Convey("Then the provider schema for the resource should contain the attribute with the expected default value and the provider schema properties commands should have been executed", func() {
				So(providerSchema, ShouldContainKey, apiKeyAuthProperty.Name)
				defaultValue, err := providerSchema[apiKeyAuthProperty.Name].DefaultFunc()
				So(err, ShouldBeNil)
				So(defaultValue, ShouldEqual, "someDefaultValue")
				So(serviceConfig.SchemaConfiguration[0].ExecuteCommandCalled, ShouldBeTrue)
			})
		})
	})

	Convey("Given a provider factory containing a command that fails to execute and also fails to get the default value from the external source", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		serviceConfig := &ServiceConfigStub{
			SchemaConfiguration: []*ServiceSchemaPropertyConfigurationStub{
				{
					SchemaPropertyName:   "apikey_auth",
					ExecuteCommandCalled: false,
					Err:                  errors.New("some error executing the command"),
					GetDefaultValueFunc:  func() (string, error) { return "", fmt.Errorf("some error gettign the default value") },
				},
			},
		}
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
			},
			serviceConfiguration: serviceConfig,
		}
		Convey("When createTerraformProviderSchema is called with an empty provider schema", func() {
			providerSchema := map[string]*schema.Schema{}
			p.configureProviderPropertyFromPluginConfig(providerSchema, "apikey_auth", true)
			Convey("Then the provider schema for the resource should contain the attribute with default empty value and the provider schema properties commands should have been executed", func() {
				// The APIs are expected to complain about the value being empty instead of the plugin failing at this stage
				So(providerSchema, ShouldContainKey, apiKeyAuthProperty.Name)
				defaultValue, err := providerSchema[apiKeyAuthProperty.Name].DefaultFunc()
				So(err, ShouldBeNil)
				So(defaultValue, ShouldBeEmpty)
				So(serviceConfig.SchemaConfiguration[0].ExecuteCommandCalled, ShouldBeTrue)
			})
		})
	})
}

func TestConfigureProvider(t *testing.T) {
	Convey("Given a provider factory configured with an analyser and graphite telemetry", t, func() {
		metricChannel := make(chan string)
		pc, telemetryHost, telemetryPort := udpServer(metricChannel)
		port, _ := strconv.Atoi(telemetryPort)
		defer pc.Close()
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name: headerProperty.Name,
					},
				},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
			},
			serviceConfiguration: &ServiceConfigStub{
				Telemetry: &TelemetryProviderGraphite{
					Port:   port,
					Host:   telemetryHost,
					Prefix: "openapi",
				},
			},
		}
		testProviderSchema := newTestSchema(apiKeyAuthProperty, headerProperty)
		Convey("When configureProvider is called with a backend that is not multi-region and the returned configureFunc is invoked upon ", func() {
			backendConfig := &specStubBackendConfiguration{}
			configureFunc := p.configureProvider(backendConfig, &providerConfigurationEndPoints{})
			client, err := configureFunc(testProviderSchema.getResourceData(t))
			providerClient := client.(*ProviderClient)
			Convey("And the client should implement ClientOpenAPI interface and the telemetry server should have been received the expected counter metrics increase", func() {
				var _ ClientOpenAPI = providerClient
				So(err, ShouldBeNil)
				assertExpectedMetric(t, metricChannel, "openapi.terraform.openapi_plugin_version.total_runs:1|c|#openapi_plugin_version:dev")
			})
		})
	})

	Convey("Given a provider factory configured with an analyser and http_endpoint telemetry", t, func() {
		httpMetricsSubmitted := false
		metricsReceived := []byte{}
		headersReceived := http.Header{}
		api := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			metricsReceived, _ = ioutil.ReadAll(req.Body)
			headersReceived = req.Header
			httpMetricsSubmitted = true
		}))
		// Close the server when test finishes
		defer api.Close()

		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name: headerProperty.Name,
					},
				},
				security: &specSecurityStub{
					securityDefinitions: &SpecSecurityDefinitions{
						newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
					},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
			},
			serviceConfiguration: &ServiceConfigStub{
				Telemetry: &TelemetryProviderHTTPEndpoint{
					URL:                      fmt.Sprintf("%s/v1/metrics", api.URL),
					Prefix:                   "openapi",
					ProviderSchemaProperties: []string{"header_name"},
				},
			},
		}
		testProviderSchema := newTestSchema(apiKeyAuthProperty, headerProperty)
		Convey("When configureProvider is called with a backend that is not multi-region and the returned configureFunc is invoked upon ", func() {
			backendConfig := &specStubBackendConfiguration{}
			configureFunc := p.configureProvider(backendConfig, &providerConfigurationEndPoints{})
			client, err := configureFunc(testProviderSchema.getResourceData(t))
			providerClient := client.(*ProviderClient)
			Convey("And the client should implement ClientOpenAPI interface and the http_endpoint telemetry server should have been received the expected counter metrics increase", func() {
				So(err, ShouldBeNil)
				var _ ClientOpenAPI = providerClient
				So(httpMetricsSubmitted, ShouldBeTrue)
				So(headersReceived.Get("header_name"), ShouldEqual, "someHeaderValue")

				tm := telemetryMetric{}
				err = json.Unmarshal(metricsReceived, &tm)
				So(err, ShouldBeNil)
				So(tm.MetricType, ShouldEqual, metricTypeCounter)
				So(tm.MetricName, ShouldEqual, "openapi.terraform.openapi_plugin_version.total_runs")
				So(tm.Tags, ShouldResemble, []string{"openapi_plugin_version:dev"})
			})
		})
	})
}

func TestCreateProviderConfig(t *testing.T) {
	Convey("Given a provider factory configured with a global header and security scheme", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		expectedSecurityDefinitions := SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition(apiKeyAuthProperty.Name, authorizationHeader),
		}
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name: headerProperty.Name,
					},
				},
				security: &specSecurityStub{
					securityDefinitions: &expectedSecurityDefinitions,
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthProperty.Name: []string{""},
						},
					}),
				},
			},
		}
		Convey(fmt.Sprintf("When createProviderConfig is called with a resource data containing the values for api key auth property (%s) and a header property (%s)", apiKeyAuthProperty.Default, headerProperty.Default), func() {
			testProviderSchema := newTestSchema(apiKeyAuthProperty, headerProperty)
			providerConfiguration, err := p.createProviderConfig(testProviderSchema.getResourceData(t), &providerConfigurationEndPoints{})
			Convey("Then the provider configuration returned should contain the header with its value (coming from the resource schema) and the provider configuration returned should contain the apiKey security with the correct apiKey header name and value (coming from the resource schema)", func() {
				So(err, ShouldBeNil)
				So(providerConfiguration.Headers[headerProperty.Name], ShouldEqual, headerProperty.Default)
				So(providerConfiguration.SecuritySchemaDefinitions[apiKeyAuthProperty.Name].getContext().(apiKey).name, ShouldEqual, expectedSecurityDefinitions[0].getAPIKey().Name)
				So(providerConfiguration.SecuritySchemaDefinitions[apiKeyAuthProperty.Name].getContext().(apiKey).value, ShouldEqual, apiKeyAuthProperty.Default)
			})
		})
	})

	Convey("Given a provider factory configured with a global header and security scheme that use non terraform compliant names", t, func() {
		apiKeyAuthProperty := newStringSchemaDefinitionPropertyWithDefaults("apikey_auth", "", true, false, "someAuthValue")
		var headerProperty = newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		var apiKeyAuthPreferredNonCompliantNameProperty = newStringSchemaDefinitionPropertyWithDefaults("apiKeyAuth", "", true, false, "someAuthValue")
		var headerNonCompliantNameProperty = newStringSchemaDefinitionPropertyWithDefaults("headerName", "", true, false, "someHeaderValue")

		expectedSecurityDefinitions := SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition(apiKeyAuthPreferredNonCompliantNameProperty.Name, authorizationHeader),
		}
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name: headerNonCompliantNameProperty.Name,
					},
				},
				security: &specSecurityStub{
					securityDefinitions: &expectedSecurityDefinitions,
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{
						{
							apiKeyAuthPreferredNonCompliantNameProperty.Name: []string{""},
						},
					}),
				},
			},
		}
		Convey(fmt.Sprintf("When createProviderConfig is called with a resource data containing the values for api key auth property (%s) and a header property (%s)", apiKeyAuthProperty.Default, headerProperty.Default), func() {
			testProviderSchema := newTestSchema(apiKeyAuthPreferredNonCompliantNameProperty, headerNonCompliantNameProperty)
			providerConfiguration, err := p.createProviderConfig(testProviderSchema.getResourceData(t), &providerConfigurationEndPoints{})
			Convey("Then the provider configuration returned should contain the header with its value (coming from the resource schema) and the provider configuration returned should contain the apiKey security with the correct apiKey header name and value (coming from the resource schema)", func() {
				So(err, ShouldBeNil)
				// provider config keys are always snake_case
				So(providerConfiguration.Headers["header_name"], ShouldEqual, headerProperty.Default)
				// The key values stored in the provider configuration are always terraform compliant names, hence querying 'apiKeyAuth' with its corresponding snake_case name
				So(providerConfiguration.SecuritySchemaDefinitions["api_key_auth"].getContext().(apiKey).name, ShouldEqual, expectedSecurityDefinitions[0].getAPIKey().Name)
				So(providerConfiguration.SecuritySchemaDefinitions["api_key_auth"].getContext().(apiKey).value, ShouldEqual, apiKeyAuthPreferredNonCompliantNameProperty.Default)
			})
		})
	})

	Convey("Given a provider factory configured with a global header that has a preferred terraform name configured", t, func() {
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "", true, false, "someHeaderValue")
		headerPreferredNameProperty := newStringSchemaDefinitionPropertyWithDefaults("header_name", "preferred_header_name", true, false, "someHeaderValue")
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{
					SpecHeaderParam{
						Name:          headerPreferredNameProperty.Name,
						TerraformName: headerPreferredNameProperty.PreferredName,
					},
				},
				security: &specSecurityStub{},
			},
		}
		Convey(fmt.Sprintf("When createProviderConfig is called with a resource data containing the values for a header property (%s)", headerProperty.Default), func() {
			testProviderSchema := newTestSchema(headerPreferredNameProperty)
			providerConfiguration, err := p.createProviderConfig(testProviderSchema.getResourceData(t), &providerConfigurationEndPoints{})
			Convey("Then the provider configuration returned should contain the header with its value (coming from the resource schema), the key used to look up the value is the actual header name and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(providerConfiguration.Headers[headerPreferredNameProperty.PreferredName], ShouldEqual, headerProperty.Default)
			})
		})
	})

	Convey("Given a provider factory where the internal specAnalyser.GetSecurity().GetAPIKeySecurityDefinitions() call somehow return an error", t, func() {
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				security: &specSecurityStub{
					error: fmt.Errorf("some error"),
				},
			},
		}
		Convey("When createProviderConfig is called", func() {
			_, err := p.createProviderConfig(nil, &providerConfigurationEndPoints{})
			Convey("Then the error returned should be nil", func() {
				So(err.Error(), ShouldEqual, "some error")
			})
		})
	})

}

func TestGetProviderResourceName(t *testing.T) {
	Convey("Given a provider factory", t, func() {
		p := providerFactory{
			name: "provider",
		}
		Convey("When getProviderResourceName is called with a resource name", func() {
			expectedResourceName := "resource"
			providerResourceName, err := p.getProviderResourceName(expectedResourceName)
			Convey("Then the ", func() {

			})
			Convey("Then the value returned should be the expected and err returned should be nil", func() {
				So(err, ShouldBeNil)
				So(providerResourceName, ShouldEqual, fmt.Sprintf("%s_%s", p.name, expectedResourceName))
			})
		})
		Convey("When getProviderResourceName is called with a resource name that has version", func() {
			expectedResourceName := "resource_v1"
			providerResourceName, err := p.getProviderResourceName(expectedResourceName)
			Convey("Then the value returned should be the expected and err returned should be nil", func() {
				So(err, ShouldBeNil)
				So(providerResourceName, ShouldEqual, fmt.Sprintf("%s_%s", p.name, expectedResourceName))
			})
		})
		Convey("When getProviderResourceName is called with an empty resource name", func() {
			expectedResourceName := ""
			_, err := p.getProviderResourceName(expectedResourceName)
			Convey("Then the err returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "resource name can not be empty")
			})
		})
	})
}

func TestCreateTerraformProviderResourceMapAndDataSourceInstanceMap(t *testing.T) {
	testCases := []struct {
		name                   string
		specV2stub             SpecAnalyser
		expectedResourceName   string
		expectedDataSourceName string
		expectedError          error
	}{
		{
			name: "happy path",
			specV2stub: &specAnalyserStub{
				resources: []SpecResource{newSpecStubResource("resource", "/v1/resource", false, &SpecSchemaDefinition{})},
			},
			expectedResourceName:   "provider_resource",
			expectedDataSourceName: "provider_resource_instance",
		},
		{
			name: "getTerraformCompliantResources fails ",
			specV2stub: &specAnalyserStub{
				error: fmt.Errorf("error getting terraform compliant resources"),
			},
			expectedError: errors.New("error getting terraform compliant resources"),
		},
		{
			name: "getProviderResourceName fails ",
			specV2stub: &specAnalyserStub{
				resources: []SpecResource{newSpecStubResource("", "/v1/resource", false, &SpecSchemaDefinition{})},
			},
			expectedError: errors.New("resource name can not be empty"),
		},
		{
			name: "createTerraformDataSource fails",
			specV2stub: &specAnalyserStub{
				resources: []SpecResource{&specStubResource{
					name: "hello",
					funcGetResourceSchema: func() (*SpecSchemaDefinition, error) {
						return nil, errors.New("createTerraformDataSource failed")
					},
				}},
			},
			expectedError: errors.New("createTerraformDataSource failed"),
		},
	}

	Convey("Given a providerFactory", t, func() {
		for _, tc := range testCases {
			p := providerFactory{
				name:         "provider",
				specAnalyser: tc.specV2stub,
			}
			Convey(fmt.Sprintf("When createTerraformProviderResourceMapAndDataSourceInstanceMap method is called: %s", tc.name), func() {
				resourceMap, dataSourceMap, err := p.createTerraformProviderResourceMapAndDataSourceInstanceMap()
				Convey("Then the result returned should be the expected one", func() {
					So(err, ShouldResemble, tc.expectedError)
					if tc.expectedError == nil {
						So(resourceMap, ShouldContainKey, tc.expectedResourceName)
						So(dataSourceMap, ShouldContainKey, tc.expectedDataSourceName)
					}
				})
			})
		}
	})
}

func TestCreateTerraformProviderDataSourceInstanceMap_ignore_resource(t *testing.T) {
	p := providerFactory{
		name: "provider",
		specAnalyser: &specAnalyserStub{
			resources: []SpecResource{
				newSpecStubResource("resource", "/v1/resource", true, &SpecSchemaDefinition{}),
			},
		},
	}
	resourceMap, dataSourceMap, err := p.createTerraformProviderResourceMapAndDataSourceInstanceMap()
	assert.Nil(t, err)
	assert.Empty(t, resourceMap)
	assert.Empty(t, dataSourceMap)
}

func TestCreateTerraformProviderDataSourceInstanceMap_duplicate_resource(t *testing.T) {
	Convey("Given a providerFactory", t, func() {
		p := providerFactory{
			name: "provider",
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{
					newSpecStubResource("resource", "/v1/resource", false, &SpecSchemaDefinition{}),
					newSpecStubResource("resource", "/v1/resource", false, &SpecSchemaDefinition{})},
			},
		}
		Convey("When the createTerraformProviderResourceMapAndDataSourceInstanceMap method is called", func() {
			resourceMap, dataSourceMap, err := p.createTerraformProviderResourceMapAndDataSourceInstanceMap()
			Convey("Then the returned resource and data source maps should be empty and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceMap, ShouldBeEmpty)
				So(dataSourceMap, ShouldBeEmpty)
			})
		})
	})
}

func TestCreateTerraformProviderDataSourceMap(t *testing.T) {

	testCases := []struct {
		name                 string
		specV2stub           SpecAnalyser
		expectedResourceName string
		expectedError        error
	}{
		{
			name: "happy path",
			specV2stub: &specAnalyserStub{
				dataSources: []SpecResource{newSpecStubResource("resource", "/v1/resource", false, &SpecSchemaDefinition{})},
			},
			expectedResourceName: "provider_resource",
		},
		{
			name: "getProviderResourceName fails ",
			specV2stub: &specAnalyserStub{
				dataSources: []SpecResource{newSpecStubResource("", "/v1/resource", false, &SpecSchemaDefinition{})},
			},
			expectedError: errors.New("resource name can not be empty"),
		},
		{
			name: "createTerraformDataSource fails",
			specV2stub: &specAnalyserStub{
				dataSources: []SpecResource{&specStubResource{
					name: "hello",
					funcGetResourceSchema: func() (*SpecSchemaDefinition, error) {
						return nil, errors.New("createTerraformDataSource failed")
					},
				}},
			},
			expectedError: errors.New("createTerraformDataSource failed"),
		},
	}

	Convey("Given a providerFactory", t, func() {
		for _, tc := range testCases {
			p := providerFactory{
				name:         "provider",
				specAnalyser: tc.specV2stub,
			}
			Convey(fmt.Sprintf("When createTerraformProviderDataSourceMap method is called: %s", tc.name), func() {
				schemaResource, err := p.createTerraformProviderDataSourceMap()
				Convey("Then the result returned should be the expected one", func() {
					So(err, ShouldResemble, tc.expectedError)
					if tc.expectedResourceName != "" {
						So(schemaResource, ShouldContainKey, tc.expectedResourceName)
					}
				})
			})
		}
	})
}

func TestGetTelemetryHandler(t *testing.T) {
	Convey("Given a providerFactory configured with a telemetry provider", t, func() {
		expectedTelemetryProvider := &TelemetryProviderHTTPEndpoint{
			URL: "https://endpoint/v1/metrics",
		}
		expectedResourceData := &schema.ResourceData{}
		expectedProviderName := "provider_name"
		providerFactory := providerFactory{
			name: expectedProviderName,
			serviceConfiguration: &ServiceConfigStub{
				Telemetry: expectedTelemetryProvider,
			},
		}
		Convey("When the newSpecV2Resource method is called", func() {
			telemetryHandler := providerFactory.GetTelemetryHandler(expectedResourceData)
			Convey("Then the handler returned should be configured as expected", func() {
				So(telemetryHandler, ShouldNotBeNil)
				So(telemetryHandler, ShouldHaveSameTypeAs, telemetryHandlerTimeoutSupport{})
				So(telemetryHandler.(telemetryHandlerTimeoutSupport).providerName, ShouldEqual, expectedProviderName)
				So(telemetryHandler.(telemetryHandlerTimeoutSupport).openAPIVersion, ShouldEqual, version.Version)
				So(telemetryHandler.(telemetryHandlerTimeoutSupport).timeout, ShouldEqual, telemetryTimeout)
				So(telemetryHandler.(telemetryHandlerTimeoutSupport).telemetryProvider, ShouldEqual, expectedTelemetryProvider)
				So(telemetryHandler.(telemetryHandlerTimeoutSupport).data, ShouldEqual, expectedResourceData)

			})
		})
	})
}

func TestGetTelemetryHandlerReturnsNilTelemetryProviderDueToTelemetryValidationError(t *testing.T) {
	Convey("Given a providerFactory configured with a telemetry provider witn an empty URL", t, func() {
		expectedResourceData := &schema.ResourceData{}
		expectedProviderName := "provider_name"
		providerFactory := providerFactory{
			name: expectedProviderName,
			serviceConfiguration: &ServiceConfigStub{
				Telemetry: &TelemetryProviderHTTPEndpoint{
					URL: "",
				},
			},
		}
		Convey("When the newSpecV2Resource method is called", func() {
			telemetryHandler := providerFactory.GetTelemetryHandler(expectedResourceData)
			Convey("Then the handler returned should be nil", func() {
				So(telemetryHandler, ShouldBeNil)
			})
		})
	})
}

func assertExpectedMetric(t *testing.T, metricChannel chan string, expectedMetric string) {
	assertExpectedMetricAndLogging(t, metricChannel, expectedMetric, "", "", nil)
}

func assertExpectedMetricAndLogging(t *testing.T, metricChannel chan string, expectedMetric, expectedLogMetricToSubmit, expectedLogMetricSuccess string, logging *bytes.Buffer) {
	select {
	case metricReceived := <-metricChannel:
		assert.Contains(t, metricReceived, expectedMetric)
		if expectedLogMetricToSubmit != "" {
			assert.Contains(t, logging.String(), expectedLogMetricToSubmit)
		}
		if expectedLogMetricSuccess != "" {
			assert.Contains(t, logging.String(), expectedLogMetricSuccess)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("[FAIL] '%s' not reveided within the expected timeframe (timed out)", expectedMetric)
	}
}
