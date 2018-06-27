package main

import (
	"github.com/dikhan/terraform-provider-openapi/examples/goa/api/app"
	"github.com/goadesign/goa"
	"github.com/hashicorp/go-uuid"
)

var db = map[string]*app.Bottle{}

// BottleController implements the bottle resource.
type BottleController struct {
	*goa.Controller
}

// NewBottleController creates a bottle controller.
func NewBottleController(service *goa.Service) *BottleController {
	return &BottleController{Controller: service.NewController("BottleController")}
}

// Create runs the create action.
func (c *BottleController) Create(ctx *app.CreateBottleContext) error {
	// BottleController_Create: start_implement
	// Put your logic here
	id, _ := uuid.GenerateUUID()
	response := &app.Bottle{
		ID: id,
		Rating: ctx.Payload.Rating,
		Name: ctx.Payload.Name,
	}
	db[id] = response
	return ctx.Created(response)
	// BottleController_Create: end_implement
}

// Show runs the show action.
func (c *BottleController) Show(ctx *app.ShowBottleContext) error {
	// BottleController_Show: start_implement
	// Put your logic here
	if db[ctx.ID] == nil {
		return ctx.NotFound()
	}
	return ctx.OK(db[ctx.ID])
	// BottleController_Show: end_implement
}
