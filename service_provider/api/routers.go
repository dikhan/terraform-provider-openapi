package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
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
		"Index",
		"GET",
		"/v2/",
		Discovery,
	},

	Route{
		"CreateUser",
		"POST",
		"/v2/users",
		CreateUser,
	},

	Route{
		"DeleteUser",
		"DELETE",
		"/v2/users/{username}",
		DeleteUser,
	},

	Route{
		"GetUserByName",
		"GET",
		"/v2/users/{username}",
		GetUserByName,
	},

	Route{
		"UpdateUser",
		"PUT",
		"/v2/users/{username}",
		UpdateUser,
	},
}
