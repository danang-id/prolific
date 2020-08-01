package common

import "github.com/gorilla/mux"

type IRoute interface {
	Initialise(r *mux.Router)
}

