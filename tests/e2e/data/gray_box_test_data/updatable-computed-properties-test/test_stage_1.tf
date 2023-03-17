# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
  nested_list_prop {
    some_property = "value1"
    computed_property = "value1"
  }

  nested_list_prop {
    some_property = "value2"
  }
}