resource "sp1_cdns" "my_cdn" {
  label = "label"
  ips = ["127.0.0.1"]
  hostnames = ["origin.com"]
}

resource "sp2_users" "my_user" {
  username = "dikhan"
  first_name = "Daniel"
  last_name = "Khan"
  email = "info@server.com"
  password = "password1"
  phone = "6049991234"
}