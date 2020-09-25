package openapi

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/hashcode"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEndpointsSchema(t *testing.T) {
	Convey("Given a provider configuration endpoints configured with a spec analyser that has one resource", t, func() {
		resourceName := "resource_name_v1"
		p := providerConfigurationEndPoints{
			resourceNames: []string{resourceName},
		}
		Convey("When endpointsSchema is called", func() {
			resourceName := "resource_name_v1"
			s := p.endpointsSchema()
			Convey("Then the schema returned should be configured as expected", func() {
				So(s, ShouldNotBeNil)
				So(s.Type, ShouldEqual, schema.TypeSet)
				So(s.Optional, ShouldBeTrue)
				So(s.Elem, ShouldHaveSameTypeAs, &schema.Resource{})
				So(s.Elem.(*schema.Resource).Schema, ShouldContainKey, resourceName)
				// the schema Set should not be nil (this defines the unique ID )
				So(s.Set, ShouldNotBeNil)
			})
		})
	})
	Convey("Given a provider configuration endpoints configured with a spec analyser that has no resources", t, func() {
		p := providerConfigurationEndPoints{
			resourceNames: []string{},
		}
		Convey("When endpointsSchema is called", func() {
			s := p.endpointsSchema()
			Convey("Then the schema returned should be nil", func() {
				So(s, ShouldBeNil)
			})
		})
	})
}

func TestEndpointsToHash(t *testing.T) {
	Convey("Given a provider configuration endpoints configured", t, func() {
		p := providerConfigurationEndPoints{
			resourceNames: []string{},
		}
		Convey("When endpointsSchema is called with a list of resources", func() {
			resourceName := "resource_name_v1"
			schemaSetFunction := p.endpointsToHash([]string{resourceName})
			Convey("Then the schema set function returned should NOT be nil and the return int from calling the schemaSetFunction() should be the expected one", func() {
				So(schemaSetFunction, ShouldNotBeNil)
				m := map[string]interface{}{}
				m[resourceName] = "something to get the string representation from"
				var buf bytes.Buffer
				buf.WriteString(fmt.Sprintf("%s-", m[resourceName].(string)))
				So(schemaSetFunction(m), ShouldEqual, String(buf.String()))
			})
		})
	})
}

func TestEndpointsValidateFunc(t *testing.T) {
	Convey("Given a provider configuration endpoints configured", t, func() {
		p := providerConfigurationEndPoints{}
		Convey("When endpointsValidateFunc is invoked with a valid domain host", func() {
			warns, errs := p.endpointsValidateFunc()("www.valid-domain.com", "something")
			Convey("Then the warns should be nil and the errs should be nil", func() {
				So(warns, ShouldBeNil)
				So(errs, ShouldBeNil)
			})
		})
		Convey("When endpointsValidateFunc is invoked with a valid IP host", func() {
			warns, errs := p.endpointsValidateFunc()("127.0.0.1", "something")
			Convey("Then the warns should be nil and the errs should be nil", func() {
				So(warns, ShouldBeNil)
				So(errs, ShouldBeNil)
			})
		})
		Convey("When endpointsValidateFunc is invoked with a valid host using custom ports", func() {
			warns, errs := p.endpointsValidateFunc()("192.168.1.1:8080", "something")
			Convey("Then the warns should be nil and the errs should be nil", func() {
				So(warns, ShouldBeNil)
				So(errs, ShouldBeNil)
			})
		})
		Convey("When endpointsValidateFunc is invoked with a whole URL (only hostnames are valid)", func() {
			warns, errs := p.endpointsValidateFunc()("http://www.valid-domain.com", "something")
			Convey("Then the warns should be nil and the errs should be the expected one", func() {
				So(warns, ShouldBeNil)
				So(errs[0].Error(), ShouldEqual, "property 'something' value 'http://www.valid-domain.com' is not valid, please make sure the value is a valid FQDN or well formed IP (the host may contain non standard ports too followed by a colon - e,g: www.api.com:8080). The protocol used when performing the API call will be populated based on the swagger specification")

			})
		})
	})
}

//func TestGetProviderConfigEndPointsFromData(t *testing.T) {
//	Convey("Given a provider factory", t, func() {
//		expectedResource := "resource_name"
//		p := providerConfigurationEndPoints{
//			specAnalyser: &specAnalyserStub{
//				resources: []SpecResource{
//					newSpecStubResource(expectedResource, "", false, nil),
//				},
//				headers: SpecHeaderParameters{},
//				security: &specSecurityStub{
//					securityDefinitions:   &SpecSecurityDefinitions{},
//					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
//				},
//			},
//		}
//
//		resourceSchema := map[string]*schema.Schema{}
//		resourceDataMap := map[string]interface{}{}
//
//		s, err := p.endpointsSchema()
//		So(err, ShouldBeNil)
//
//		resourceSchema[providerPropertyEndPoints] = s
//
//		expectedValue := "resource_value"
//		e := map[string]interface{}{
//			expectedResource: expectedValue,
//		}
//
//		hash := p.endpointsToHash([]string{expectedResource})(e)
//
//
//		resourceDataMap[providerPropertyEndPoints] = map[string]interface{}{
//			fmt.Sprintf("%d", hash): map[string]interface{}{
//				expectedResource: expectedValue,
//			},
//		}
//		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
//
//		Convey("When configureEndpoints is called with a list of resources", func() {
//			resourceName := "resource_name_v1"
//			endpoints, err := p.configureEndpoints(resourceLocalData)
//			Convey("Then the error returned should not be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("And the endpoints returned should contain the resource with the value expected", func() {
//				So(endpoints, ShouldContainKey, resourceName)
//			})
//		})
//	})
//}
