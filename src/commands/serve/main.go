package serve

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Serve() {
	router := newRouter()

	// TODO: read config somehow to not have a static port.
	log.Fatal(http.ListenAndServe(":6969", router))
}

func newRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

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
