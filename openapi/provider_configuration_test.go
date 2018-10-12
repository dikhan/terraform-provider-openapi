package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
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
			providerConfiguration := newProviderConfiguration(headers, securitySchemaDefinitions, data)
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
