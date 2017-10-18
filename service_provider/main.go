package main

import (
	sw "github.com/dikhan/terraform-provider-api/service_provider/api"
	"log"
	"net/http"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()
	
	log.Fatal(http.ListenAndServe(":8080", router))
}
