resource "openapi_cdn_v1" "my_cdn" {
  main_optional_prop = "main_optional_value"
  list_prop {
    sub_optional_prop   = "sub_optional_value_1"
  }
}