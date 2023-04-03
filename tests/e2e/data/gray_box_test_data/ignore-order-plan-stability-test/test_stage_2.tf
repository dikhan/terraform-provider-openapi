# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  list_prop {
    string_property  = "string_value2"
    integer_property = 2

    nested_list_prop {
      string_property = "nested_value_2_2"
    }

    nested_list_prop {
      string_property = "nested_value_2_1"
    }
  }

  list_prop {
    string_property   = "string_value1"
    computed_property = "computed_value1"
    integer_property  = 1
  }
}