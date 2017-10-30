package main

import (
	"log"
	"net/http"

	sw "github.com/dikhan/terraform-provider-api/service_provider/api"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()

	log.Fatal(http.ListenAndServe(":80", router))
}
