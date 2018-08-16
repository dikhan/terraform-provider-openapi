package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"strings"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}



func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},

	Route{
		"ContentDeliveryNetworkCreateV1",
		strings.ToUpper("Post"),
		"/v1/cdns",
		ContentDeliveryNetworkCreateV1,
	},

	Route{
		"ContentDeliveryNetworkDeleteV1",
		strings.ToUpper("Delete"),
		"/v1/cdns/{id}",
		ContentDeliveryNetworkDeleteV1,
	},

	Route{
		"ContentDeliveryNetworkGetV1",
		strings.ToUpper("Get"),
		"/v1/cdns/{id}",
		ContentDeliveryNetworkGetV1,
	},

	Route{
		"ContentDeliveryNetworkUpdateV1",
		strings.ToUpper("Put"),
		"/v1/cdns/{id}",
		ContentDeliveryNetworkUpdateV1,
	},

	Route{
		"LBCreateV1",
		strings.ToUpper("Post"),
		"/v1/lbs",
		LBCreateV1,
	},

	Route{
		"LBDeleteV1",
		strings.ToUpper("Delete"),
		"/v1/lbs/{id}",
		LBDeleteV1,
	},

	Route{
		"LBGetV1",
		strings.ToUpper("Get"),
		"/v1/lbs/{id}",
		LBGetV1,
	},

	Route{
		"LBUpdateV1",
		strings.ToUpper("Put"),
		"/v1/lbs/{id}",
		LBUpdateV1,
	},

	Route{
		"ApiDiscovery",
		strings.ToUpper("Get"),
		"/swagger.yaml",
		ApiDiscovery,
	},

	Route{
		"GetVersion",
		strings.ToUpper("Get"),
		"/version",
		GetVersion,
	},
}