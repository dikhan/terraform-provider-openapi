package main

import (
	// WARNING!
	// Change this to a fully-qualified import path
	// once you place this file into your project.
	// For example,
	//
	//    sw "github.com/myname/myrepo/api"
	//
	"log"
	"net/http"

	sw "github.com/dikhan/terraform-provider-api/service_provider_example/api"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()

	log.Fatal(http.ListenAndServe(":80", router))
}
