provider "swaggercodegen" {
  ## the example server is expecting the api key to have 'apiKeyValue' (it's hard coded)
  ## auth testing can be done by tweaking this value to be something else
  apikey_auth = "${var.apikey_auth}"
}

resource "swaggercodegen_cdns_v1" "my_cdn" {
  label = "label" ## This is an immutable property (refer to swagger file)
  ips = ["127.0.0.1"] ## This is a force-new property (refer to swagger file)
  hostnames = ["origin.com"]

  example_int = 12
  example_number = 1.12
  example_boolean = true
}