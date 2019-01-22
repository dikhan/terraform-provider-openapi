package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
)

func TestNewProviderConfiguration(t *testing.T) {
	Convey("Given a headers a SpecHeaderParameters, securitySchemaDefinitions and a schema ResourceData", t, func() {
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("headerProperty", "header_property", true, false, "updatedValue")
		headers := SpecHeaderParameters{
			SpecHeaderParam{
				Name: headerProperty.getTerraformCompliantPropertyName(),
			},
		}
		securitySchemaDefinitions := &SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition(stringProperty.getTerraformCompliantPropertyName(), "someHeaderSecDefName"),
			newAPIKeyQuerySecurityDefinition(stringWithPreferredNameProperty.getTerraformCompliantPropertyName(), "someQuerySecDefName"),
		}
		data := newTestSchema(stringProperty, stringWithPreferredNameProperty, headerProperty).getResourceData(t)
		Convey("When newProviderConfiguration method is called", func() {
			providerConfiguration, err := newProviderConfiguration(headers, securitySchemaDefinitions, data)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the provider configuration data should not be nil", func() {
				So(providerConfiguration.data, ShouldNotPointTo, providerConfiguration)
			})
			Convey("Then the providerConfiguration headers should contain the configured header with the right value", func() {
				So(providerConfiguration.Headers, ShouldContainKey, headerProperty.getTerraformCompliantPropertyName())
				So(providerConfiguration.Headers[headerProperty.getTerraformCompliantPropertyName()], ShouldEqual, headerProperty.Default)
			})
			Convey("And the providerConfiguration securitySchemaDefinitions should contain the configured stringProperty security definitions with the right value", func() {
				So(providerConfiguration.SecuritySchemaDefinitions, ShouldContainKey, stringProperty.Name)
				So(providerConfiguration.SecuritySchemaDefinitions[stringProperty.Name].getContext().(apiKey).value, ShouldEqual, stringProperty.Default)
			})
			Convey("And the providerConfiguration securitySchemaDefinitions should contain the configured stringWithPreferredNameProperty security definitions with the right value", func() {
				So(providerConfiguration.SecuritySchemaDefinitions, ShouldContainKey, stringWithPreferredNameProperty.getTerraformCompliantPropertyName())
				So(providerConfiguration.SecuritySchemaDefinitions[stringWithPreferredNameProperty.getTerraformCompliantPropertyName()].getContext().(apiKey).value, ShouldEqual, stringWithPreferredNameProperty.Default)
			})
		})
	})

	Convey("Given securitySchemaDefinitions and a schema ResourceData not containing values for the security definitions", t, func() {
		headers := SpecHeaderParameters{}
		securitySchemaDefinitions := &SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition(stringProperty.getTerraformCompliantPropertyName(), "someHeaderSecDefName"),
		}
		data := newTestSchema().getResourceData(t)
		Convey("When newProviderConfiguration method is called", func() {
			_, err := newProviderConfiguration(headers, securitySchemaDefinitions, data)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Adn the error message returned should be equal to", func() {
				So(err.Error(), ShouldEqual, "security schema definition 'string_property' is missing the value, please make sure this value is provided in the terraform configuration")
			})
		})
	})

	Convey("Given a headers a SpecHeaderParameters and a schema ResourceData not containing values for the security definitions", t, func() {
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("headerProperty", "header_property", true, false, "updatedValue")
		headers := SpecHeaderParameters{
			SpecHeaderParam{
				Name: headerProperty.getTerraformCompliantPropertyName(),
			},
		}
		securitySchemaDefinitions := &SpecSecurityDefinitions{}
		data := newTestSchema().getResourceData(t)
		Convey("When newProviderConfiguration method is called", func() {
			_, err := newProviderConfiguration(headers, securitySchemaDefinitions, data)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Adn the error message returned should be equal to", func() {
				So(err.Error(), ShouldEqual, "header parameter 'header_property' is missing the value, please make sure this value is provided in the terraform configuration")
			})
		})
	})
}

func TestGetAuthenticatorFor(t *testing.T) {
	Convey("Given a providerConfiguration with some security schema definitions", t, func() {
		providerConfiguration := providerConfiguration{
			Headers: map[string]string{},
			SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
				"registered_sec_def_name": createAPIKeyAuthenticator(newAPIKeyHeaderSecurityDefinition("registeredSecDefName", "headerName"), "value"),
			},
		}
		Convey("When getAuthenticatorFor method with an existing sec def", func() {
			apiKeyAuth := providerConfiguration.getAuthenticatorFor(SpecSecurityScheme{"registered_sec_def_name"})
			Convey("Then the apikey name should be headerName", func() {
				So(apiKeyAuth.getContext().(apiKey).name, ShouldEqual, "headerName")
			})
			Convey("Then the apikey value should be value", func() {
				So(apiKeyAuth.getContext().(apiKey).value, ShouldEqual, "value")
			})
		})
		Convey("When getAuthenticatorFor method with a NON existing sec def", func() {
			apiKeyAuth := providerConfiguration.getAuthenticatorFor(SpecSecurityScheme{"nonExistingSecDef"})
			Convey("Then the apiKeyAuth returned should be nil", func() {
				So(apiKeyAuth, ShouldBeNil)
			})
		})
	})
}

func TestGetHeaderValueFor(t *testing.T) {
	Convey("Given a providerConfiguration with some headers", t, func() {
		providerConfiguration := providerConfiguration{
			Headers: map[string]string{
				"header_name": "headerValue",
			},
		}
		Convey("When getHeaderValueFor method with an existing header", func() {
			value := providerConfiguration.getHeaderValueFor(SpecHeaderParam{Name: "headerName"})
			Convey("Then the value returned should be headerValue", func() {
				So(value, ShouldEqual, "headerValue")
			})
		})
		Convey("When getHeaderValueFor method with a NON existing header", func() {
			value := providerConfiguration.getHeaderValueFor(SpecHeaderParam{Name: "nontExistingHeader"})
			Convey("Then the value returned should be empty", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
}

func TestGetRegion(t *testing.T) {
	Convey("Given a providerConfiguration with data that has no values for the region property", t, func() {
		s := newTestSchema()
		providerConfiguration := providerConfiguration{
			data: s.getResourceData(t),
			mutex: &sync.Mutex{},
		}
		Convey("When getRegion() method is called", func() {
			value := providerConfiguration.getRegion()
			Convey("Then the value returned should be empty", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
	Convey("Given a providerConfiguration with data that has a value for the region property", t, func() {
		expectedRegion := "us-west1"
		regionProperty := newStringSchemaDefinitionPropertyWithDefaults(providerPropertyRegion, "", true, false, expectedRegion)
		s := newTestSchema(regionProperty)
		providerConfiguration := providerConfiguration{
			data:  s.getResourceData(t),
			mutex: &sync.Mutex{},
		}
		Convey("When getRegion() method is called", func() {
			value := providerConfiguration.getRegion()
			Convey("Then the value returned should match the value set in the resource data for region field", func() {
				So(value, ShouldEqual, expectedRegion)
			})
		})
	})
}
