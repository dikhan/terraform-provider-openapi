provider "swaggercodegen" {
  ## the example server is expecting the api key to have 'apiKeyValue' (it's hard coded)
  ## auth testing can be done by tweaking this value to be something else
  apikey_auth  = var.apikey_auth
  x_request_id = "request header value for POST /v1/cdns"
}

resource "swaggercodegen_cdn_v1" "my_cdn" {
  label     = "label"       ## This is an immutable property (refer to swagger file)
  ips       = ["127.0.0.1"] ## This is a force-new property (refer to swagger file)
  hostnames = ["origin.com"]

  example_int                      = 25
  better_example_number_field_name = 15.78
  example_boolean                  = true

  optional_property              = "optional_property value"
  optional_computed              = "optional_computed value"
  optional_computed_with_default = "optional_computed_with_default value"

  object_property = {
    message          = "some message news2"
    detailed_message = "some message news with details"
    example_int      = 11
    example_number   = 12.23
    example_boolean  = true
  }

  array_of_objects_example {
    protocol    = "http"
    origin_port = 81
  }

  array_of_objects_example {
    protocol    = "https"
    origin_port = 443
  }
}

# This is an example on how to use interpolation for 'object' types like the object_property and be able to pass
# along to other resources property values from objects
resource "swaggercodegen_cdn_v1" "my_cdn2" {
  label     = "label"       ## This is an immutable property (refer to swagger file)
  ips       = ["127.0.0.2"] ## This is a force-new property (refer to swagger file)
  hostnames = ["origin.com"]

  example_int                      = swaggercodegen_cdn_v1.my_cdn.object_property.example_int
  better_example_number_field_name = swaggercodegen_cdn_v1.my_cdn.object_property.example_number
  example_boolean                  = swaggercodegen_cdn_v1.my_cdn.object_property.example_boolean

  object_property = {
    message          = "some message news2"
    detailed_message = "some message news with details"
    example_int      = swaggercodegen_cdn_v1.my_cdn.object_property.example_int
    example_number   = swaggercodegen_cdn_v1.my_cdn.object_property.example_number
    example_boolean  = swaggercodegen_cdn_v1.my_cdn.object_property.example_boolean
  }

  dynamic "array_of_objects_example" {
    for_each = [swaggercodegen_cdn_v1.my_cdn.array_of_objects_example[0]]
    content {
      # TF-UPGRADE-TODO: The automatic upgrade tool can't predict
      # which keys might be set in maps assigned here, so it has
      # produced a comprehensive set here. Consider simplifying
      # this after confirming which keys can be set in practice.

      origin_port = lookup(array_of_objects_example.value, "origin_port", null)
      protocol    = lookup(array_of_objects_example.value, "protocol", null)
    }
  }
  dynamic "array_of_objects_example" {
    for_each = [swaggercodegen_cdn_v1.my_cdn.array_of_objects_example[1]]
    content {
      # TF-UPGRADE-TODO: The automatic upgrade tool can't predict
      # which keys might be set in maps assigned here, so it has
      # produced a comprehensive set here. Consider simplifying
      # this after confirming which keys can be set in practice.

      origin_port = lookup(array_of_objects_example.value, "origin_port", null)
      protocol    = lookup(array_of_objects_example.value, "protocol", null)
    }
  }
}

resource "swaggercodegen_lbs_v1" "my_lb" {
  name             = "some_name"
  backends         = ["backend.com"]
  time_to_process  = 1     # the operation (post,update,delete) will take 15s in the API to complete
  simulate_failure = false # no failures wished now ;) (post,update,delete)
}

