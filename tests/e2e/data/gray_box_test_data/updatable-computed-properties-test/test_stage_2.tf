# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  nested_list_prop {
    string_property = "value2"
  }

  nested_list_prop {
    string_property = "value1"
    computed_property = "value1"
    number_property = 3
    integer_property = 5
    boolean_property = true
  }

  nested_list_prop {
    string_property = "value3"
    computed_property = "value3"
    boolean_property = false
  }
}