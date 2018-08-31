provider "swaggercodegen" {
  ## the example server is expecting the api key to have 'apiKeyValue' (it's hard coded)
  ## auth testing can be done by tweaking this value to be something else
  apikey_auth = "${var.apikey_auth}"
  x_request_id = "request header value for POST /v1/cdns"
}

resource "swaggercodegen_cdns_v1" "my_cdn" {
  label = "label" ## This is an immutable property (refer to swagger file)
  ips = ["127.0.0.1"] ## This is a force-new property (refer to swagger file)
  hostnames = ["origin.com"]

  example_int = 12
  better_example_number_field_name = 1.12
  example_boolean = true
}

resource "swaggercodegen_lbs_v1" "my_lb" {
  name = "some_name"
  backends = ["backend1.com"]
  time_to_process = 30 # the operation (post,update,delete) will take 30s in the API to complete
  simulate_failure = false # no failures wished now ;) (post,update,delete)
}