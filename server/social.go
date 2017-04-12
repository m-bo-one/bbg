package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func serveSocial(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["social"]
	w.Write([]byte("Hello " + name))
}
