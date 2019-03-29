package openapi

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetResourceNames(t *testing.T) {
	Convey("Given a provider configuration endpoints configured with a spec analyser that has one resource", t, func() {
		expectedResourceName := "resource_name_v1"
		resource := newSpecStubResource(expectedResourceName, "", false, nil)
		p := providerConfigurationEndPoints{
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{resource},
				headers:   SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
		}
		Convey("When getResourceNames is called with a map containing some resources ", func() {
			resources, err := p.getResourceNames()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the list of resource should not be empty", func() {
				So(resources, ShouldNotBeEmpty)
			})
			Convey("And should match the expected number of resources", func() {
				So(len(resources), ShouldEqual, 1)
			})
			Convey("And the list should contain the expected resources", func() {
				So(resources, ShouldContain, expectedResourceName)
			})
		})
	})

	Convey("Given a provider configuration endpoints configured with a spec analyser that has NO resources", t, func() {
		p := providerConfigurationEndPoints{
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{},
				headers:   SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
		}
		Convey("When getResourceNames is called with an empty map", func() {
			resources, err := p.getResourceNames()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the list of resource should be empty", func() {
				So(resources, ShouldBeEmpty)
			})
		})
	})

}

func TestEndpointsSchema(t *testing.T) {
	Convey("Given a provider configuration endpoints configured with a spec analyser that has one resource", t, func() {
		resourceName := "resource_name_v1"
		resource := newSpecStubResource(resourceName, "", false, nil)
		p := providerConfigurationEndPoints{
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{resource},
				headers:   SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
		}
		Convey("When endpointsSchema is called", func() {
			resourceName := "resource_name_v1"
			s, err := p.endpointsSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the schema returned should NOT be nil", func() {
				So(s, ShouldNotBeNil)
			})
			Convey("And the schema type should be a set", func() {
				So(s.Type, ShouldEqual, schema.TypeSet)
			})
			Convey("And the schema should be optional", func() {
				So(s.Optional, ShouldBeTrue)
			})
			Convey("And the schema element should be of type schema resource", func() {
				So(s.Elem, ShouldHaveSameTypeAs, &schema.Resource{})
			})
			Convey("And the schema element resource schema should contain the expected resource", func() {
				So(s.Elem.(*schema.Resource).Schema, ShouldContainKey, resourceName)
			})
			Convey("And the schema Set should not be nil (this defines the unique ID )", func() {
				So(s.Set, ShouldNotBeNil)
			})
		})
	})
	Convey("Given a provider configuration endpoints configured withh a spec analyser that has no resources", t, func() {
		p := providerConfigurationEndPoints{
			specAnalyser: &specAnalyserStub{
				resources: []SpecResource{},
				headers:   SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
		}
		Convey("When endpointsSchema is called", func() {
			s, err := p.endpointsSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the schema returned should be nil", func() {
				So(s, ShouldBeNil)
			})
		})
	})
}

func TestEndpointsToHash(t *testing.T) {
	Convey("Given a provider configuration endpoints configured", t, func() {
		p := providerConfigurationEndPoints{
			specAnalyser: &specAnalyserStub{
				headers: SpecHeaderParameters{},
				security: &specSecurityStub{
					securityDefinitions:   &SpecSecurityDefinitions{},
					globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
				},
			},
		}
		Convey("When endpointsSchema is called with a list of resources", func() {
			resourceName := "resource_name_v1"
			schemaSetFunction := p.endpointsToHash([]string{resourceName})
			Convey("Then the schema set function returned should NOT be nil", func() {
				So(schemaSetFunction, ShouldNotBeNil)
			})
			Convey("And the return int from calling the schemaSetFunction() shuold be the expected one", func() {
				m := map[string]interface{}{}
				m[resourceName] = "something to get the string represention from"
				var buf bytes.Buffer
				buf.WriteString(fmt.Sprintf("%s-", m[resourceName].(string)))
				So(schemaSetFunction(m), ShouldEqual, hashcode.String(buf.String()))
			})
		})
	})
}

func TestEndpointsValidateFunc(t *testing.T) {
	Convey("Given a provider configuration endpoints configured", t, func() {
		p := providerConfigurationEndPoints{}
		Convey("When endpointsValidateFunc is invoked with a valid domain host", func() {
			warns, errs := p.endpointsValidateFunc()("www.valid-domain.com", "something")
			Convey("Then the warns should be nil", func() {
				So(warns, ShouldBeNil)
			})
			Convey("Then the errs should be nil", func() {
				So(errs, ShouldBeNil)
			})
		})
		Convey("When endpointsValidateFunc is invoked with a valid IP host", func() {
			warns, errs := p.endpointsValidateFunc()("127.0.0.1", "something")
			Convey("Then the warns should be nil", func() {
				So(warns, ShouldBeNil)
			})
			Convey("Then the errs should be nil", func() {
				So(errs, ShouldBeNil)
			})
		})
		Convey("When endpointsValidateFunc is invoked with a valid host using custom ports", func() {
			warns, errs := p.endpointsValidateFunc()("192.168.1.1:8080", "something")
			Convey("Then the warns should be nil", func() {
				So(warns, ShouldBeNil)
			})
			Convey("Then the errs should be nil", func() {
				So(errs, ShouldBeNil)
			})
		})
		Convey("When endpointsValidateFunc is invoked with a whole URL (only hostnames are valid)", func() {
			warns, errs := p.endpointsValidateFunc()("http://www.valid-domain.com", "something")
			Convey("Then the warns should be nil", func() {
				So(warns, ShouldBeNil)
			})
			Convey("Then the errs should be nil", func() {
				So(errs, ShouldNotBeNil)
			})
			Convey("And the error message should be the expected one", func() {
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
