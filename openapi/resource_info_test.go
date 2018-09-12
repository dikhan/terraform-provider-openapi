package openapi

import (
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetResourceURL(t *testing.T) {

	Convey("Given resource info is configured with https scheme and basePath='/', path='/v1/resource', host='www.host.com'", t, func() {
		expectedBasePath := "/"
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should be https://www.host.com/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s", expectedScheme, expectedHost, expectedPath))
			})
		})
	})

	Convey("Given resource info is configured with http scheme and basePath='/', path='/v1/resource', host='www.host.com'", t, func() {
		expectedBasePath := "/"
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "http"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should be http://www.host.com/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s", expectedScheme, expectedHost, expectedPath))
			})
		})
	})

	Convey("Given resource info is not configured with any scheme and basePath='/', path='/v1/resource', host='www.host.com'", t, func() {
		expectedBasePath := "/"
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "http"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should use the default scheme which is http://www.host.com/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s", expectedScheme, expectedHost, expectedPath))
			})
		})
	})

	Convey("Given resource info is configured with https scheme, basePath is not empty nor / and path is '/v1/resource''", t, func() {
		expectedBasePath := "/api"
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should use the default scheme which is https://www.host.com/api/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedScheme, expectedHost, expectedBasePath, expectedPath))
			})
		})
	})

	Convey("Given resource info is configured with https scheme, basePath is empty and path is '/v1/resource''", t, func() {
		expectedBasePath := ""
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should use the default scheme which is https://www.host.com/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s", expectedScheme, expectedHost, expectedPath))
			})
		})
	})

	Convey("Given resource info is configured with https scheme, basePath is / and path is '/v1/resource''", t, func() {
		expectedBasePath := "/"
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should use the default scheme which is https://www.host.com/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s", expectedScheme, expectedHost, expectedPath))
			})
		})
	})

	Convey("Given resource info is configured with https scheme, basePath does not start with / and path is '/v1/resource''", t, func() {
		expectedBasePath := "api"
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should use the default scheme which is https://www.host.com/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s/%s%s", expectedScheme, expectedHost, expectedBasePath, expectedPath))
			})
		})
	})

	Convey("Given resource info is configured with a path that does not start with /", t, func() {
		expectedBasePath := "/"
		expectedPath := "v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			resourceURL, err := r.getResourceURL()
			Convey("Then the value returned should use the default scheme which is https://www.host.com/v1/resource and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s/%s", expectedScheme, expectedHost, expectedPath))
			})
		})
	})

	Convey("Given resource info is missing the path", t, func() {
		expectedBasePath := ""
		expectedPath := "/v1/resource"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        "",
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			_, err := r.getResourceURL()
			Convey("Then there should be returned error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given resource info is missing the host", t, func() {
		expectedBasePath := ""
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        "",
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceURL method is called'", func() {
			_, err := r.getResourceURL()
			Convey("Then there should be returned error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetResourceIDURL(t *testing.T) {
	Convey("Given resource info is configured with 'https' scheme and basePath='/', path='/v1/resource', host='www.host.com'", t, func() {
		expectedBasePath := "/"
		expectedPath := "/v1/resource"
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceIDURL method is called with ID=1234", func() {
			id := "1234"
			resourceIDURL, err := r.getResourceIDURL(id)
			Convey("Then the value returned should be https://www.host.com/v1/resource/1234 and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceIDURL, ShouldEqual, fmt.Sprintf("%s://%s%s/%s", expectedScheme, expectedHost, expectedPath, id))
			})
		})
	})

	Convey("Given resource info is missing the host", t, func() {
		expectedBasePath := ""
		expectedHost := "www.host.com"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        "",
			host:        expectedHost,
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceIDURL method is called with ID=1234", func() {
			_, err := r.getResourceIDURL("1234")
			Convey("Then there should be returned error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given resource info is missing the path", t, func() {
		expectedBasePath := ""
		expectedPath := "/v1/resource"
		expectedScheme := "https"

		r := resourceInfo{
			basePath:    expectedBasePath,
			path:        expectedPath,
			host:        "",
			httpSchemes: []string{expectedScheme},
		}
		Convey("When getResourceIDURL method is called with ID=1234", func() {
			_, err := r.getResourceIDURL("1234")
			Convey("Then there should be returned error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetImmutableProperties(t *testing.T) {
	Convey("Given resource info is configured with schemaDefinition that contains a property 'immutable_property' that is immutable", t, func() {
		extensions := spec.Extensions{}
		extensions.Add("x-terraform-immutable", true)
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"id": {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
						},
						"immutable_property": {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
						},
					},
				},
			},
		}
		Convey("When getImmutableProperties method is called", func() {
			immutableProperties := r.getImmutableProperties()
			Convey("Then the array returned should contain 'immutable_property'", func() {
				So(immutableProperties, ShouldContain, "immutable_property")
			})
			Convey("And the 'id' property should be ignored even if it's marked as immutable (should never happen though, edge case)", func() {
				So(immutableProperties, ShouldNotContain, "id")
			})
		})
	})

	Convey("Given resource info is configured with schemaDefinition that DOES NOT contain immutable properties", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"id": {
							VendorExtensible: spec.VendorExtensible{},
						},
						"mutable_property": {
							VendorExtensible: spec.VendorExtensible{Extensions: spec.Extensions{}},
						},
					},
				},
			},
		}
		Convey("When getImmutableProperties method is called", func() {
			immutableProperties := r.getImmutableProperties()
			Convey("Then the array returned should be empty", func() {
				So(immutableProperties, ShouldBeEmpty)
			})
		})
	})

}

func TestCreateTerraformPropertyBasicSchema(t *testing.T) {
	Convey("Given a swagger schema definition that has a property of type 'string'", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_prop": propSchema,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("string_prop", propSchema)
			Convey("Then the resulted terraform property schema should be of type string too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'integer'", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"integer"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"int_prop": propSchema,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("int_prop", propSchema)
			Convey("Then the resulted terraform property schema should be of type int too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeInt)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'number'", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"number"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"number_prop": propSchema,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("number_prop", propSchema)
			Convey("Then the resulted terraform property schema should be of type float too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeFloat)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'boolean'", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"boolean_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"boolean"},
							},
						},
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("boolean_prop", r.schemaDefinition.Properties["boolean_prop"])
			Convey("Then the resulted terraform property schema should be of type int too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeBool)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array'", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"array_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"array"},
							},
						},
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("array_prop", r.schemaDefinition.Properties["array_prop"])
			Convey("Then the resulted terraform property schema should be of type array too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
			})
			Convey("And the array elements are of the default type string (only supported type for now)", func() {
				So(reflect.TypeOf(tfPropSchema.Elem).Elem(), ShouldEqual, reflect.TypeOf(schema.Schema{}))
				So(tfPropSchema.Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property 'string_prop' which is required", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
					Required: []string{"string_prop"}, // This array contains the list of properties that are required
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("string_prop", r.schemaDefinition.Properties["string_prop"])
			Convey("Then the returned value should be true", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Required, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger property schema with 'x-terraform-force-new' metadata", t, func() {
		extensions := spec.Extensions{}
		extensions.Add("x-terraform-force-new", true)
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{Extensions: extensions},
			SchemaProps: spec.SchemaProps{
				Type: []string{"boolean"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"boolean_prop": propSchema,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("boolean_prop", propSchema)
			Convey("Then the resulted terraform property schema should be of type int too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.ForceNew, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger property schema configured with readOnly (computed)", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"boolean"},
			},
			SwaggerSchemaProps: spec.SwaggerSchemaProps{ReadOnly: true},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"boolean_prop": propSchema,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("boolean_prop", propSchema)
			Convey("Then the resulted terraform property schema should be configured as computed", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Computed, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger property schema configured with 'x-terraform-force-new' and 'x-terraform-sensitive' metadata", t, func() {
		extensions := spec.Extensions{}
		extensions.Add("x-terraform-force-new", true)
		extensions.Add("x-terraform-sensitive", true)
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{Extensions: extensions},
			SchemaProps: spec.SchemaProps{
				Type: []string{"boolean"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"boolean_prop": propSchema,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("boolean_prop", propSchema)
			Convey("Then the resulted terraform property schema should be configured as forceNew and sensitive", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.ForceNew, ShouldBeTrue)
				So(tfPropSchema.Sensitive, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger property schema configured with default value", t, func() {
		expectedDefaultValue := "defaultValue"
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type:    []string{"boolean"},
				Default: expectedDefaultValue,
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"boolean_prop": propSchema,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertyBasicSchema("boolean_prop", propSchema)
			Convey("Then the resulted terraform property schema should be configured with the according default value, ", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Default, ShouldEqual, expectedDefaultValue)
			})
		})
	})
}

func TestIsArrayProperty(t *testing.T) {
	Convey("Given a swagger property schema of type 'array'", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"array_prop": propSchema,
					},
				},
			},
		}
		Convey("When isArrayProperty method is called", func() {
			isArray := r.isArrayProperty(propSchema)
			Convey("Then the returned value should be true", func() {
				So(isArray, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger property schema of type different than 'array'", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_prop": propSchema,
					},
				},
			},
		}
		Convey("When isArrayProperty method is called", func() {
			isArray := r.isArrayProperty(propSchema)
			Convey("Then the returned value should be false", func() {
				So(isArray, ShouldBeFalse)
			})
		})
	})
}

func TestIsRequired(t *testing.T) {
	Convey("Given a swagger schema definition that has a property 'string_prop' that is required", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_prop": propSchema,
					},
					Required: []string{"string_prop"}, // This array contains the list of properties that are required
				},
			},
		}
		Convey("When isRequired method is called", func() {
			isRequired := r.isRequired("string_prop", r.schemaDefinition.Required)
			Convey("Then the returned value should be true", func() {
				So(isRequired, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property 'string_prop' that is not required", t, func() {
		propSchema := spec.Schema{
			VendorExtensible: spec.VendorExtensible{},
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_prop": propSchema,
					},
				},
			},
		}
		Convey("When isRequired method is called", func() {
			isRequired := r.isRequired("string_prop", r.schemaDefinition.Required)
			Convey("Then the returned value should be false", func() {
				So(isRequired, ShouldBeFalse)
			})
		})
	})
}

func TestCreateTerraformPropertySchema(t *testing.T) {
	Convey("Given a swagger schema definition that has a property 'string_prop' of type string, required, sensitive and has a default value 'defaultValue'", t, func() {
		expectedDefaultValue := "defaultValue"
		extensions := spec.Extensions{}
		extensions.Add("x-terraform-sensitive", true)
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_prop": {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
							SchemaProps: spec.SchemaProps{
								Type:    []string{"string"},
								Default: expectedDefaultValue,
							},
						},
					},
					Required: []string{"string_prop"}, // This array contains the list of properties that are required
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertySchema("string_prop", r.schemaDefinition.Properties["string_prop"])
			Convey("Then the returned tf tfPropSchema should be of type string", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeString)
			})
			Convey("And a validateFunc should be configured", func() {
				So(tfPropSchema.ValidateFunc, ShouldNotBeNil)
			})
			Convey("And be configured as required, sensitive and the default value should match 'defaultValue'", func() {
				So(tfPropSchema.Required, ShouldBeTrue)
			})
			Convey("And be configured as sensitive", func() {
				So(tfPropSchema.Sensitive, ShouldBeTrue)
			})
			Convey("And the default value should match 'defaultValue'", func() {
				So(tfPropSchema.Default, ShouldEqual, expectedDefaultValue)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property 'array_prop' of type array", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"array_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"array"},
							},
						},
					},
					Required: []string{"array_prop"}, // This array contains the list of properties that are required
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := r.createTerraformPropertySchema("array_prop", r.schemaDefinition.Properties["array_prop"])
			Convey("Then the returned tf tfPropSchema should be of type array", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
			})
			Convey("And there should not be any validation function attached to it", func() {
				So(tfPropSchema.ValidateFunc, ShouldBeNil)
			})
		})
	})
}

func TestValidateFunc(t *testing.T) {
	Convey("Given a swagger schema definition that has one property", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"array_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"array"},
							},
						},
					},
				},
			},
		}
		Convey("When validateFunc method is called", func() {
			validateFunc := r.validateFunc("array_prop", r.schemaDefinition.Properties["array_prop"])
			Convey("Then the returned validateFunc should not be nil", func() {
				So(validateFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property which is supposed to be computed but has a default value set", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"array_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type:    []string{"array"},
								Default: "defaultValue",
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{ReadOnly: true},
						},
					},
				},
			},
		}
		Convey("When validateFunc method is called", func() {
			validateFunc := r.validateFunc("array_prop", r.schemaDefinition.Properties["array_prop"])
			Convey("Then the returned validateFunc should not be nil", func() {
				So(validateFunc, ShouldNotBeNil)
			})
			Convey("And when the function is executed it should return an error as computed properties can not have default values", func() {
				_, errs := validateFunc("", "")
				So(errs, ShouldNotBeEmpty)
			})
		})
	})
}

func TestCreateTerraformResourceSchema(t *testing.T) {
	Convey("Given a swagger schema definition that has multiple properties - 'string_prop', 'int_prop', 'number_prop', 'bool_prop' and 'array_prop'", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
						"intProp": { // This prop does not have a terraform compliant name; however an automatic translation is performed behind the scenes to make it compliant
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"integer"},
							},
						},
						"number_prop": {
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									"x-terraform-field-name": "numberProp", // making use of specific extension to override field name; but the new field name is not terrafrom name compliant - hence an automatic translation is performed behind the scenes to make it compliant
								},
							},
							SchemaProps: spec.SchemaProps{
								Type: []string{"number"},
							},
						},
						"bool_prop": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"boolean"},
							},
						},
						"arrayProp": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"array"},
							},
						},
					},
				},
			},
		}
		Convey("When createTerraformResourceSchema method is called", func() {
			resourceSchema, err := r.createTerraformResourceSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the tf resource schema returned should match the swagger props - 'string_prop', 'int_prop', 'number_prop' and 'bool_prop' and 'array_prop', ", func() {
				So(resourceSchema, ShouldNotBeNil)
				So(resourceSchema, ShouldContainKey, "string_prop")
				So(resourceSchema, ShouldContainKey, "int_prop")
				So(resourceSchema, ShouldContainKey, "number_prop")
				So(resourceSchema, ShouldContainKey, "bool_prop")
				So(resourceSchema, ShouldContainKey, "array_prop")
			})
		})
	})
}

func TestConvertToTerraformCompliantFieldName(t *testing.T) {
	Convey("Given a property with a name that is terraform field name compliant", t, func() {
		propertyName := "some_prop_name_that_is_terraform_field_name_compliant"
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						propertyName: {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When convertToTerraformCompliantFieldName method is called", func() {
			fieldName := r.convertToTerraformCompliantFieldName(propertyName, r.schemaDefinition.Properties[propertyName])
			Convey("And string return is terraform field name compliant, ", func() {
				So(fieldName, ShouldEqual, propertyName)
			})
		})
	})

	Convey("Given a property with a name that is NOT terraform field name compliant", t, func() {
		propertyName := "thisPropIsNotTerraformField_Compliant"
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						propertyName: {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When convertToTerraformCompliantFieldName method is called", func() {
			fieldName := r.convertToTerraformCompliantFieldName(propertyName, r.schemaDefinition.Properties[propertyName])
			Convey("And string return is terraform field name compliant, ", func() {
				So(fieldName, ShouldEqual, "this_prop_is_not_terraform_field_compliant")
			})
		})
	})

	Convey("Given a property with a name that is NOT terraform field name compliant but has an extension that overrides it", t, func() {
		propertyName := "thisPropIsNotTerraformField_Compliant"
		expectedPropertyName := "this_property_is_now_terraform_field_compliant"
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						propertyName: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfFieldName: expectedPropertyName,
								},
							},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When convertToTerraformCompliantFieldName method is called", func() {
			fieldName := r.convertToTerraformCompliantFieldName(propertyName, r.schemaDefinition.Properties[propertyName])
			Convey("And string return is terraform field name compliant, ", func() {
				So(fieldName, ShouldEqual, expectedPropertyName)
			})
		})
	})

	Convey("Given a property with a name that is NOT terraform field name compliant but has an extension that overrides it which in turn is also not terraform name compliant", t, func() {
		propertyName := "thisPropIsNotTerraformField_Compliant"
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						propertyName: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfFieldName: "thisPropIsAlsoNotTerraformField_Compliant",
								},
							},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When convertToTerraformCompliantFieldName method is called", func() {
			fieldName := r.convertToTerraformCompliantFieldName(propertyName, r.schemaDefinition.Properties[propertyName])
			Convey("And string return is terraform field name compliant, ", func() {
				So(fieldName, ShouldEqual, "this_prop_is_also_not_terraform_field_compliant")
			})
		})
	})
}

func TestGetResourceIdentifier(t *testing.T) {
	Convey("Given a swagger schema definition that has an id property", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						idDefaultPropertyName: {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			id, err := r.getResourceIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be 'id'", func() {
				So(id, ShouldEqual, idDefaultPropertyName)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'id' property but has a property configured with x-terraform-id set to TRUE", t, func() {
		extensions := spec.Extensions{}
		extensions.Add(extTfID, true)
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"some-other-id": {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			id, err := r.getResourceIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be 'some-other-id'", func() {
				So(id, ShouldEqual, "some-other-id")
			})
		})
	})

	Convey("Given a swagger schema definition that HAS BOTH an 'id' property AND ALSO a property configured with x-terraform-id set to true", t, func() {
		extensions := spec.Extensions{}
		extensions.Add(extTfID, true)
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"id": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
						"some-other-id": {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			id, err := r.getResourceIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be 'some-other-id' as it takes preference over the default 'id' property", func() {
				So(id, ShouldEqual, "some-other-id")
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'id' property but has a property configured with x-terraform-id set to FALSE", t, func() {
		extensions := spec.Extensions{}
		extensions.Add(extTfID, false)
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"some-other-id": {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			_, err := r.getResourceIdentifier()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition that NEITHER HAS an 'id' property NOR a property configured with x-terraform-id set to true", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"prop-that-is-not-id": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
						"prop-that-is-not-id-and-does-not-have-id-metadata": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			_, err := r.getResourceIdentifier()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetStatusIdentifier(t *testing.T) {
	Convey("Given a swagger schema definition that has an status property", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						statusDefaultPropertyName: {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := r.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be 'status'", func() {
				So(status, ShouldEqual, statusDefaultPropertyName)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with x-terraform-field-status set to TRUE", t, func() {
		extensions := spec.Extensions{}
		extensions.Add(extTfFieldStatus, true)
		expectedStatusProperty := "some-other-property-holding-status"
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						expectedStatusProperty: {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			id, err := r.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be 'some-other-property-holding-status'", func() {
				So(id, ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that HAS BOTH an 'status' property AND ALSO a property configured with 'x-terraform-field-status' set to true", t, func() {
		extensions := spec.Extensions{}
		extensions.Add(extTfFieldStatus, true)
		expectedStatusProperty := "some-other-property-holding-status"
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"status": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
						expectedStatusProperty: {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			id, err := r.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be 'some-other-property-holding-status' as it takes preference over the default 'status' property", func() {
				So(id, ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with 'x-terraform-field-status' set to FALSE", t, func() {
		extensions := spec.Extensions{}
		extensions.Add(extTfFieldStatus, false)
		expectedStatusProperty := "some-other-property-holding-status"
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						expectedStatusProperty: {
							VendorExtensible: spec.VendorExtensible{Extensions: extensions},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := r.getStatusIdentifier()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition that NEITHER HAS an 'status' property NOR a property configured with 'x-terraform-field-status' set to true", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"prop-that-is-not-status": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
						"prop-that-is-not-status-and-does-not-have-status-metadata-either": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := r.getStatusIdentifier()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition with a property configured with 'x-terraform-field-status' set to true but is not readonly", t, func() {
		r := resourceInfo{
			schemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"prop-that-is-not-status": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: false,
							},
						},
						"prop-that-is-not-status-and-does-not-have-status-metadata-either": {
							VendorExtensible: spec.VendorExtensible{},
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := r.getStatusIdentifier()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestIsIDProperty(t *testing.T) {
	Convey("Given a swagger schema definition", t, func() {
		r := resourceInfo{}
		Convey("When isIDProperty method is called with property named 'id'", func() {
			isIDProperty := r.isIDProperty("id")
			Convey("Then the error returned should be nil", func() {
				So(isIDProperty, ShouldBeTrue)
			})
		})
		Convey("When isIDProperty method is called with property NOT named 'id'", func() {
			isIDProperty := r.isIDProperty("something_not_id")
			Convey("Then the error returned should be nil", func() {
				So(isIDProperty, ShouldBeFalse)
			})
		})
	})
}

func TestIsStatusProperty(t *testing.T) {
	Convey("Given a swagger schema definition", t, func() {
		r := resourceInfo{}
		Convey("When isStatusProperty method is called with property named 'status'", func() {
			isStatusProperty := r.isStatusProperty("status")
			Convey("Then the error returned should be nil", func() {
				So(isStatusProperty, ShouldBeTrue)
			})
		})
		Convey("When isStatusProperty method is called with property NOT named 'status'", func() {
			isStatusProperty := r.isStatusProperty("something_not_status")
			Convey("Then the error returned should be nil", func() {
				So(isStatusProperty, ShouldBeFalse)
			})
		})
	})
}

func TestPropertyNameMatchesDefaultName(t *testing.T) {
	Convey("Given a swagger schema definition", t, func() {
		r := resourceInfo{}
		Convey("When propertyNameMatchesDefaultName method is called with property named 'status' and an expected name matching the property property name", func() {
			propertyNameMatchesDefaultName := r.propertyNameMatchesDefaultName("status", "status")
			Convey("Then the error returned should be nil", func() {
				So(propertyNameMatchesDefaultName, ShouldBeTrue)
			})
		})
		Convey("When propertyNameMatchesDefaultName method is called with property named 'ID' which is not terraform compliant name and an expected property name", func() {
			propertyNameMatchesDefaultName := r.propertyNameMatchesDefaultName("ID", "id")
			Convey("Then the error returned should be nil", func() {
				So(propertyNameMatchesDefaultName, ShouldBeTrue)
			})
		})
		Convey("When propertyNameMatchesDefaultName method is called with property NOT matching the expected property name", func() {
			propertyNameMatchesDefaultName := r.propertyNameMatchesDefaultName("something_not_status", "")
			Convey("Then the error returned should be nil", func() {
				So(propertyNameMatchesDefaultName, ShouldBeFalse)
			})
		})
	})
}

func TestIsResourcePollingEnabled(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey("When isResourcePollingEnabled method is called with a list of responses where one of the reponses matches the response status received and has the 'x-terraform-resource-poll-enabled' extension set to true", func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollEnabled, true)
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			isResourcePollingEnabled, _ := r.isResourcePollingEnabled(responses, http.StatusAccepted)
			Convey("Then the bool returned should be true", func() {
				So(isResourcePollingEnabled, ShouldBeTrue)
			})
		})
		Convey("When isResourcePollingEnabled method is called with a list of responses where one of the reponses matches the response status received and has the 'x-terraform-resource-poll-enabled' extension set to false", func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollEnabled, false)
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			isResourcePollingEnabled, _ := r.isResourcePollingEnabled(responses, http.StatusAccepted)
			Convey("Then the bool returned should be false", func() {
				So(isResourcePollingEnabled, ShouldBeFalse)
			})
		})
		Convey("When isResourcePollingEnabled method is called with list of responses where non of the codes match the given response http code", func() {
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusOK: {},
					},
				},
			}
			isResourcePollingEnabled, _ := r.isResourcePollingEnabled(responses, http.StatusAccepted)
			Convey("Then bool returned should be false", func() {
				So(isResourcePollingEnabled, ShouldBeFalse)
			})
		})
	})
}

func TestGetResourcePollTargetStatuses(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey("When getResourcePollTargetStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses'", func() {
			expectedTarget := "deployed"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTarget)
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses, err := r.getResourcePollTargetStatuses(responses.StatusCodeResponses[http.StatusAccepted])
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the status returned should contain", func() {
				So(statuses, ShouldContain, expectedTarget)
			})
		})
	})
}

func TestGetResourcePollPendingStatuses(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey("When getResourcePollPendingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-pending-statuses'", func() {
			expectedStatus := "deploy_pending"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollPendingStatuses, expectedStatus)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses, err := r.getResourcePollPendingStatuses(responses.StatusCodeResponses[http.StatusAccepted])
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the status returned should contain", func() {
				So(statuses, ShouldContain, expectedStatus)
			})
		})
	})
}

func TestGetPollingStatuses(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey("When getPollingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses'", func() {
			expectedTarget := "deployed"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTarget)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses, err := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the statuses returned should contain", func() {
				So(statuses, ShouldContain, expectedTarget)
			})
		})

		Convey("When getPollingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses' containing multiple targets (comma separated with spaces)", func() {
			expectedTargets := "deployed, completed, done"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTargets)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses, err := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the statuses returned should contain expected targets", func() {
				// the expected Targets are a list of targets with no spaces whatsoever, hence why the removal of spaces
				for _, expectedTarget := range strings.Split(strings.Replace(expectedTargets, " ", "", -1), ",") {
					So(statuses, ShouldContain, expectedTarget)
				}
			})
		})

		Convey("When getPollingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses' containing multiple targets (comma separated with no spaces)", func() {
			expectedTargets := "deployed,completed,done"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTargets)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses, err := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the statuses returned should contain expected targets", func() {
				for _, expectedTarget := range strings.Split(expectedTargets, ",") {
					So(statuses, ShouldContain, expectedTarget)
				}
			})
		})

		Convey("When getPollingStatuses method is called with a response that has does not have a given extension 'x-terraform-resource-poll-completed-statuses'", func() {
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{},
							},
						},
					},
				},
			}
			_, err := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetResourceTimeout(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey(fmt.Sprintf("When getResourceTimeout method is called with an operation that has the extension '%s'", extTfResourceTimeout), func() {
			expectedTimeout := "30s"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			post := &spec.Operation{
				VendorExtensible: spec.VendorExtensible{
					Extensions: extensions,
				},
			}
			duration, err := r.getResourceTimeout(post)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
				So(*duration, ShouldEqual, time.Duration(30*time.Second))
			})
		})
	})
}

func TestGetTimeDuration(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in seconds", extTfResourceTimeout), func() {
			expectedTimeout := "30s"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
				So(*duration, ShouldEqual, time.Duration(30*time.Second))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in minutes (using fractions)", extTfResourceTimeout), func() {
			expectedTimeout := "20.5m"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
				So(*duration, ShouldEqual, time.Duration((20*time.Minute)+(30*time.Second)))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in hours", extTfResourceTimeout), func() {
			expectedTimeout := "1h"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
				So(*duration, ShouldEqual, time.Duration(1*time.Hour))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES NOT contain the extension passed in '%s'", extTfResourceTimeout), func() {
			expectedTimeout := "30s"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, "nonExistingExtension")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is an empty string", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "invalid duration value: ''. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is a negative duration", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "-1.5h")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "invalid duration value: '-1.5h'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is NOT supported (distinct than s,m and h)", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "300ms")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "invalid duration value: '300ms'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
			})
		})
	})
}

func TestGetDuration(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey("When getDuration method is called a valid formatted time'", func() {
			duration, err := r.getDuration("30s")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the statuses returned should contain", func() {
				fmt.Println(duration)
				So(*duration, ShouldEqual, time.Duration(30*time.Second))
			})
		})
		Convey("When getDuration method is called a invalid formatted time'", func() {
			_, err := r.getDuration("some invalid formatted time")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestShouldIgnoreResource(t *testing.T) {
	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-exclude-resource with value true", t, func() {
		r := resourceInfo{
			createPathInfo: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-terraform-exclude-resource": true,
							},
						},
					},
				},
			},
		}
		Convey("When shouldIgnoreResource method is called", func() {
			shouldIgnoreResource := r.shouldIgnoreResource()
			Convey("Then the value returned should be true", func() {
				So(shouldIgnoreResource, ShouldBeTrue)
			})
		})
	})
	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-exclude-resource with value false", t, func() {
		r := resourceInfo{
			createPathInfo: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-terraform-exclude-resource": false,
							},
						},
					},
				},
			},
		}
		Convey("When shouldIgnoreResource method is called", func() {
			shouldIgnoreResource := r.shouldIgnoreResource()
			Convey("Then the value returned should be true", func() {
				So(shouldIgnoreResource, ShouldBeFalse)
			})
		})
	})
	Convey("Given a terraform compliant resource that has a POST operation that DOES NOT contain the x-terraform-exclude-resource extension", t, func() {
		r := resourceInfo{
			createPathInfo: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{},
						},
					},
				},
			},
		}
		Convey("When shouldIgnoreResource method is called", func() {
			shouldIgnoreResource := r.shouldIgnoreResource()
			Convey("Then the value returned should be true", func() {
				So(shouldIgnoreResource, ShouldBeFalse)
			})
		})
	})
}

func TestGetResourceOverrideHost(t *testing.T) {
	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-resource-host with a non parametrized host containing the host to use", t, func() {
		expectedHost := "some.api.domain.com"
		r := resourceInfo{
			createPathInfo: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: expectedHost,
							},
						},
					},
				},
			},
		}
		Convey("When getResourceOverrideHost method is called", func() {
			host := r.getResourceOverrideHost()
			Convey("Then the value returned should be the host value", func() {
				So(host, ShouldEqual, expectedHost)
			})
		})
	})

	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-resource-host with a parametrized host containing the host to use", t, func() {
		expectedHost := "some.api.${serviceProviderName}.domain.com"
		r := resourceInfo{
			createPathInfo: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: expectedHost,
							},
						},
					},
				},
			},
		}
		Convey("When getResourceOverrideHost method is called", func() {
			host := r.getResourceOverrideHost()
			Convey("Then the value returned should be the host value", func() {
				So(host, ShouldEqual, expectedHost)
			})
		})
	})

	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-resource-host with an empty string value", t, func() {
		expectedHost := ""
		r := resourceInfo{
			createPathInfo: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: expectedHost,
							},
						},
					},
				},
			},
		}
		Convey("When getResourceOverrideHost method is called", func() {
			host := r.getResourceOverrideHost()
			Convey("Then the value returned should be the host value", func() {
				So(host, ShouldEqual, expectedHost)
			})
		})
	})
}

func TestIsMultiRegionHost(t *testing.T) {
	Convey("Given a resourceInfo", t, func() {
		r := resourceInfo{}
		Convey("When isMultiRegionHost method is called with a non multi region host", func() {
			expectedHost := "some.api.domain.com"
			isMultiRegion, _ := r.isMultiRegionHost(expectedHost)
			Convey("Then the value returned should be false", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
		})
		Convey("When isMultiRegionHost method is called with a multi region host", func() {
			expectedHost := "some.api.${%s}.domain.com"
			isMultiRegion, _ := r.isMultiRegionHost(expectedHost)
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
		})
		Convey("When isMultiRegionHost method is called with a multi region host that has region at the beginning", func() {
			expectedHost := "${%s}.domain.com"
			isMultiRegion, _ := r.isMultiRegionHost(expectedHost)
			Convey("Then the value returned should be false", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
		})
	})
}

func TestIsMultiRegionResource(t *testing.T) {
	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-resource-host with a parametrized host containing region variable", t, func() {
		serviceProviderName := "serviceProviderName"
		r := resourceInfo{
			createPathInfo: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: fmt.Sprintf("some.api.${%s}.domain.com", serviceProviderName),
							},
						},
					},
				},
			},
		}
		Convey("When isMultiRegionResource method is called with a set of extensions where one matches the region for which the above 's-terraform-resource-host' extension is for", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, serviceProviderName), "uswest,useast")
			isMultiRegion, regions := r.isMultiRegionResource(rootLevelExtensions)
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("And the map returned should contain uswest", func() {
				So(regions, ShouldContainKey, "uswest")
				So(regions["uswest"], ShouldEqual, "some.api.uswest.domain.com")
			})
			Convey("And the map returned should contain useast", func() {
				So(regions, ShouldContainKey, "useast")
				So(regions["useast"], ShouldEqual, "some.api.useast.domain.com")
			})
		})

		Convey("When isMultiRegionResource method is called with a set of extensions where NONE matches the region for which the above 's-terraform-resource-host' extension is for", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, "someOtherServiceProvider"), "rst, dub")
			isMultiRegion, regions := r.isMultiRegionResource(rootLevelExtensions)
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
			Convey("And the regions map returned should be empty", func() {
				So(regions, ShouldBeEmpty)
			})
		})

		Convey("When isMultiRegionResource method is called with a set of extensions where one matches the region for which the above 's-terraform-resource-host' extension is for BUT the values are not comma separated", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, serviceProviderName), "uswest useast")
			isMultiRegion, regions := r.isMultiRegionResource(rootLevelExtensions)
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("And the map returned should contain uswest", func() {
				So(regions, ShouldContainKey, "uswestuseast")
				So(regions["uswestuseast"], ShouldEqual, "some.api.uswestuseast.domain.com")
			})
		})

		Convey("When isMultiRegionResource method is called with a set of extensions where one matches the region for which the above 's-terraform-resource-host' extension is for BUT the values are comma separated with spaces", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, serviceProviderName), "uswest, useast")
			isMultiRegion, regions := r.isMultiRegionResource(rootLevelExtensions)
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("And the map returned should contain uswest", func() {
				So(regions, ShouldContainKey, "uswest")
				So(regions["uswest"], ShouldEqual, "some.api.uswest.domain.com")
			})
			Convey("And the map returned should contain useast", func() {
				So(regions, ShouldContainKey, "useast")
				So(regions["useast"], ShouldEqual, "some.api.useast.domain.com")
			})
		})
	})
}
