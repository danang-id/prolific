package web_hook

import (
	"github.com/gorilla/mux"
	"net/http"
)

func New() Route {
	return Route{}
}

type Route struct {}

func (route Route) Initialise(r *mux.Router) {
	r.Path("/github").Methods(http.MethodPost).HandlerFunc(github)
}