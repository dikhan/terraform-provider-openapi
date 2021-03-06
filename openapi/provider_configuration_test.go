package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewProviderConfiguration(t *testing.T) {
	Convey("Given a headers a SpecHeaderParameters, securitySchemaDefinitions, a schema ResourceData and a providerConfigurationEndPoints", t, func() {
		headerProperty := newStringSchemaDefinitionPropertyWithDefaults("headerProperty", "header_property", true, false, "updatedValue")

		specAnalyser := &specAnalyserStub{
			headers: SpecHeaderParameters{
				SpecHeaderParam{
					Name: headerProperty.GetTerraformCompliantPropertyName(),
				},
			},
			security: &specSecurityStub{
				securityDefinitions: &SpecSecurityDefinitions{
					newAPIKeyHeaderSecurityDefinition(stringProperty.GetTerraformCompliantPropertyName(), "someHeaderSecDefName"),
					newAPIKeyQuerySecurityDefinition(stringWithPreferredNameProperty.GetTerraformCompliantPropertyName(), "someQuerySecDefName"),
				},
				globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
			},
		}

		providerConfigurationEndPoints := &providerConfigurationEndPoints{}

		data := newTestSchema(stringProperty, stringWithPreferredNameProperty, headerProperty).getResourceData(t)
		Convey("When newProviderConfiguration method is called", func() {
			providerConfiguration, err := newProviderConfiguration(specAnalyser, data, providerConfigurationEndPoints)
			Convey("Then the error providerConfiguretion headers and security definitions should be configured as expected and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(providerConfiguration.Headers, ShouldContainKey, headerProperty.GetTerraformCompliantPropertyName())
				So(providerConfiguration.Headers[headerProperty.GetTerraformCompliantPropertyName()], ShouldEqual, headerProperty.Default)
				So(providerConfiguration.SecuritySchemaDefinitions, ShouldContainKey, stringProperty.Name)
				So(providerConfiguration.SecuritySchemaDefinitions[stringProperty.Name].getContext().(apiKey).value, ShouldEqual, stringProperty.Default)
				So(providerConfiguration.SecuritySchemaDefinitions, ShouldContainKey, stringWithPreferredNameProperty.GetTerraformCompliantPropertyName())
				So(providerConfiguration.SecuritySchemaDefinitions[stringWithPreferredNameProperty.GetTerraformCompliantPropertyName()].getContext().(apiKey).value, ShouldEqual, stringWithPreferredNameProperty.Default)
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
			Convey("Then the apikey name should be headerName and the apikey value should have the expected value", func() {
				So(apiKeyAuth.getContext().(apiKey).name, ShouldEqual, "headerName")
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
		providerConfiguration := providerConfiguration{}
		Convey("When getRegion() method is called", func() {
			value := providerConfiguration.getRegion()
			Convey("Then the value returned should be empty", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
	Convey("Given a providerConfiguration with data that has a value for the region property", t, func() {
		expectedRegion := "us-west1"
		providerConfiguration := providerConfiguration{
			Region: expectedRegion,
		}
		Convey("When getRegion() method is called", func() {
			value := providerConfiguration.getRegion()
			Convey("Then the value returned should match the value set in the resource data for region field", func() {
				So(value, ShouldEqual, expectedRegion)
			})
		})
	})
}

func TestGetEndPoint(t *testing.T) {
	Convey("Given a providerConfiguration configured with some endpoints", t, func() {
		expectedResource := "cdn_v1"
		expectedResourceValue := "www.endpoint.com"
		providerConfiguration := providerConfiguration{
			Endpoints: map[string]string{
				expectedResource: expectedResourceValue,
			},
		}
		Convey("When getEndPoint method is called with an existing resource name", func() {
			value := providerConfiguration.getEndPoint(expectedResource)
			Convey("Then the value returned should be the expected value", func() {
				So(value, ShouldEqual, expectedResourceValue)
			})
		})
		Convey("When getEndPoint method is called with an NON existing resource name", func() {
			value := providerConfiguration.getEndPoint("nonExistingResource")
			Convey("Then the value returned should be empty", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
	Convey("Given a providerConfiguration configured with nil endpoints", t, func() {
		providerConfiguration := providerConfiguration{
			Endpoints: nil,
		}
		Convey("When getEndPoint method is called with an existing resource name", func() {
			value := providerConfiguration.getEndPoint("cdn_v1")
			Convey("Then the value returned should be empty", func() {
				So(value, ShouldBeEmpty)
			})
		})
	})
}
