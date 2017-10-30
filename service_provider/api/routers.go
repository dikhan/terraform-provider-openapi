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
	b, err := ioutil.ReadFile(pwd + "/resources/swagger.json")
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
		"/swagger.json",
		Discovery,
	},

	Route{
		"CreateUser",
		"POST",
		"/users",
		CreateUser,
	},

	Route{
		"DeleteUser",
		"DELETE",
		"/users/{username}",
		DeleteUser,
	},

	Route{
		"GetUserByName",
		"GET",
		"/users/{username}",
		GetUserByName,
	},

	Route{
		"UpdateUser",
		"PUT",
		"/users/{username}",
		UpdateUser,
	},
}
