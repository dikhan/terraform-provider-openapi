package main

import (
	"log"
	"net/http"

	sw "github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen/api/api"
)

func main() {
	errs := make(chan error)
	router := sw.NewRouter()
	// Starting HTTP server
	go func() {
		addr := ":80"
		log.Printf("Staring HTTP service on %s ...", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			errs <- err
		}
	}()

	// Starting HTTPS server
	go func() {
		sslAddr := ":443"
		log.Printf("Staring HTTPS service on %s ...", sslAddr)
		if err := http.ListenAndServeTLS(sslAddr, "ssl/certificate.crt", "ssl/privateKey.key", router); err != nil {
			errs <- err
		}
	}()

	// This will run forever until channel receives error
	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}
}
