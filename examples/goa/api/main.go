//go:generate goagen bootstrap -d github.com/dikhan/terraform-provider-openapi/examples/goa/api/design

package main

import (
	"github.com/dikhan/terraform-provider-openapi/examples/goa/api/app"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
)

func main() {
	// Create service
	service := goa.New("cellar")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "bottle" controller
	c := NewBottleController(service)
	app.MountBottleController(service, c)

	// Start service
	if err := service.ListenAndServe(":9090"); err != nil {
		service.LogError("startup", "err", err)
	}

}
