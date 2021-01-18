terraform {
  required_providers {
    goa = {
      source  = "terraform.example.com/examplecorp/goa"
      version = ">= 1.0.0"
    }
  }
}

resource "goa_bottles" "my_bottle" {
  name = "Name of bottle"
  rating = 3
  vintage = 2653
}