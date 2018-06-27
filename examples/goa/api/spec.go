package main

import (
	"github.com/goadesign/goa"
)

// SpecController implements the Spec resource.
type SpecController struct {
	*goa.Controller
}

// NewSpecController creates a Spec controller.
func NewSpecController(service *goa.Service) *SpecController {
	return &SpecController{Controller: service.NewController("SpecController")}
}
