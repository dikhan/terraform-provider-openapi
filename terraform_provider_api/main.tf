resource "sp_cdns" "my_cdn" {
  label = "label"
  ips = ["127.0.0.1"]
  hostnames = ["origin.com"]
}