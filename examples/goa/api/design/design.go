package design

import . "github.com/goadesign/goa/design"
import . "github.com/goadesign/goa/design/apidsl"

var _ = API("cellar", func() {
	Description("The wine review service")
	Host("localhost:9090")
})

var BottlePayload = Type("BottlePayload", func() {
	Description("BottlePayload is the type used to create bottles")

	Attribute("id", String, "Unique bottle ID", func() {
		// This makes the id attribute read-only, this means that the value will be computed by the API and therefore parameter
		// is not expected from the user when invoking the API
		ReadOnly()
	})

	Attribute("name", String, "Name of bottle", func() {
		MinLength(1)
	})
	Attribute("vintage", Integer, "Vintage of bottle", func() {
		Minimum(1900)
	})
	Attribute("rating", Integer, "Rating of bottle", func() {
		Minimum(1)
		Maximum(5)
	})
	Required("name", "vintage", "rating")
})

var BottleMedia = MediaType("application/vnd.gophercon.goa.bottle", func() {
	TypeName("bottle")
	// Reusing BottlePayload reduces the boiler plate having to define attribute properties in one place only
	Reference(BottlePayload)

	Attributes(func() {
		Attribute("id")
		Attribute("name")
		Attribute("vintage")
		Attribute("rating")
		Required("id", "name", "vintage", "rating")
	})

	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("vintage")
		Attribute("rating")
	})
})

var _ = Resource("bottle", func() {
	Description("A wine bottle")
	BasePath("/bottles")

	Action("create", func() {
		Description("creates a bottle")
		Routing(POST("/"))
		Payload(BottlePayload)
		Response(Created, BottleMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError)
	})

	Action("show", func() {
		Description("shows a bottle")
		Routing(GET("/:id"))
		Params(func() {
			Param("id", String)
		})
		Response(OK, BottleMedia)
		Response(NotFound)
	})
})