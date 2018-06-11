package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

func Discovery(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	b, err := ioutil.ReadFile(pwd + "/resources/swagger.yaml")
	if err != nil {
		fmt.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

var routes = Routes{
	Route{
		"ApiDiscovery",
		"GET",
		"/swagger.yaml",
		Discovery,
	},

	Route{
		"ContentDeliveryNetworkCreateV1",
		"POST",
		"/v1/cdns",
		ContentDeliveryNetworkCreateV1,
	},

	Route{
		"ContentDeliveryNetworkDeleteV1",
		"DELETE",
		"/v1/cdns/{id}",
		ContentDeliveryNetworkDeleteV1,
	},

	Route{
		"ContentDeliveryNetworkGetV1",
		"GET",
		"/v1/cdns/{id}",
		ContentDeliveryNetworkGetV1,
	},

	Route{
		"ContentDeliveryNetworkUpdateV1",
		"PUT",
		"/v1/cdns/{id}",
		ContentDeliveryNetworkUpdateV1,
	},

	Route{
		"GetVersion",
		"GET",
		"/version",
		GetVersion,
	},
}
