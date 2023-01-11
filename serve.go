package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Serve struct{}

func (serve *Serve) Run(config *Config) {
	router := serve.newRouter(config)

	fmt.Fprintln(os.Stderr, "Listening on", fmt.Sprintf("http://%s:%d", config.Settings.ListenAddress, config.Settings.ListenPort))

	// Listen to configured address and port.
	log.Fatal(http.ListenAndServe(
		fmt.Sprintf("%s:%d", config.Settings.ListenAddress, config.Settings.ListenPort),
		router,
	))
}

func (serve *Serve) newRouter(config *Config) *mux.Router {
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

	return router
}
