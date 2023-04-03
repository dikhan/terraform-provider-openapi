# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  label             = "some label"
  list_prop         = ["value1", "value3", "value2"]
  integer_list_prop = [1, 3, 2]

  nested_list_prop {
    some_property        = "value1"
    other_property_str   = "otherValue1"
    other_property_int   = 5
    other_property_float = 3.14
    other_property_bool  = true
    other_property_list  = ["someValue1"]
    other_property_object {
      deeply_nested_property = "someDeeplyNestedValue1"
    }
  }

  nested_list_prop {
    some_property = "value3"
  }

  nested_list_prop {
    some_property        = "value2"
    other_property_str   = "otherValue2"
    other_property_int   = 10
    other_property_float = 1.23
    other_property_bool  = false
    other_property_list  = ["someValue2"]
    other_property_object {
      deeply_nested_property = "someDeeplyNestedValue2"
    }
  }
}