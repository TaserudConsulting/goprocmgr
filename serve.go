package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Serve struct {
	config *Config
}

func (serve *Serve) Run() {
	router := serve.newRouter()

	log.Printf("Listening on http://%s:%d\n", serve.config.Settings.ListenAddress, serve.config.Settings.ListenPort)

	// Listen to configured address and port.
	log.Fatal(http.ListenAndServe(
		fmt.Sprintf("%s:%d", serve.config.Settings.ListenAddress, serve.config.Settings.ListenPort),
		router,
	))
}

func (serve *Serve) newRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Web server endpoints
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	router.HandleFunc("/web/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/style.css")
	})
	router.HandleFunc("/web/script.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/script.js")
	})

	// This endpoint is served at GET /api/config and it returns the
	// currently loaded config.
	router.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serve.config)
	}).Methods(http.MethodGet)

	router.HandleFunc("/api/config/server", func(w http.ResponseWriter, r *http.Request) {
		var server ServerConfig
		resp := make(map[string]string)

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&server)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp["message"] = fmt.Sprintf("%s", err)
		} else if err := serve.config.WriteServer(server); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp["message"] = fmt.Sprintf("%s", err)
		} else {
			w.WriteHeader(http.StatusCreated)
			resp["message"] = "OK"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodPost)

	router.HandleFunc("/api/config/server/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		serve.config.DeleteServer(vars["name"])

		resp := make(map[string]string)
		resp["message"] = "OK"

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodDelete)

	return router
}
