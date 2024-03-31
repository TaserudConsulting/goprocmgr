package main // import "github.com/TaserudConsulting/goprocmgr"

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Serve struct {
	config *Config
	runner *Runner
}

type ServeMessageResponse struct {
	Message string `json:"message"`
}

type ServeRunnerResponseItem struct {
	Name   string   `json:"name"`
	Stdout []string `json:"stdout"`
	Stderr []string `json:"stderr"`
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

//go:embed "static"
var static embed.FS

func (serve *Serve) newRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Web server endpoints
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, _ := static.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(file)
	})
	router.HandleFunc("/web/style.css", func(w http.ResponseWriter, r *http.Request) {
		file, _ := static.ReadFile("static/style.css")
		w.Header().Set("Content-Type", "text/css")
		w.Write(file)
	})
	router.HandleFunc("/web/script.js", func(w http.ResponseWriter, r *http.Request) {
		file, _ := static.ReadFile("static/script.js")
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(file)
	})
	router.HandleFunc("/web/van-1.5.0.nomodule.min.js", func(w http.ResponseWriter, r *http.Request) {
		file, _ := static.ReadFile("static/van-1.5.0.nomodule.min.js")
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(file)
	})

	// This endpoint is served at GET /api/config and it returns the
	// currently loaded config.
	router.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serve.config)
	}).Methods(http.MethodGet)

	router.HandleFunc("/api/config/server", func(w http.ResponseWriter, r *http.Request) {
		var server ServerConfig
		var resp ServeMessageResponse

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&server)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp.Message = fmt.Sprintf("%s", err)
		} else if err := serve.config.WriteServer(server); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp.Message = fmt.Sprintf("%s", err)
		} else {
			w.WriteHeader(http.StatusCreated)
			resp.Message = "OK"
		}

		// Stop servers on update in case it's running.
		serve.runner.Stop(server.Name)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodPost)

	router.HandleFunc("/api/config/server/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var resp ServeMessageResponse

		err := serve.runner.Stop(vars["name"])

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp.Message = fmt.Sprintf("Failed to stop running server %s: %s", vars["name"], err)
		} else {
			w.WriteHeader(http.StatusOK)
			resp.Message = "OK"
			serve.config.DeleteServer(vars["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/api/runner", func(w http.ResponseWriter, r *http.Request) {
		resp := make(map[string]ServeRunnerResponseItem)

		for key, value := range serve.runner.ActiveProcesses {
			resp[key] = ServeRunnerResponseItem{
				Name:   key,
				Stdout: value.Stdout,
				Stderr: value.Stderr,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodGet)

	router.HandleFunc("/api/runner/{name}", func(w http.ResponseWriter, r *http.Request) {
		var resp ServeMessageResponse
		vars := mux.Vars(r)

		err := serve.runner.Start(vars["name"])

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp.Message = fmt.Sprintf("Failed to start server %s, %s", vars["name"], err)
		} else {
			w.WriteHeader(http.StatusCreated)
			resp.Message = "OK"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodPost)

	router.HandleFunc("/api/runner/{name}", func(w http.ResponseWriter, r *http.Request) {
		var resp ServeMessageResponse
		vars := mux.Vars(r)

		err := serve.runner.Stop(vars["name"])

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp.Message = fmt.Sprintf("Failed to stop server %s, %s", vars["name"], err)
		} else {
			w.WriteHeader(http.StatusOK)
			resp.Message = "OK"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodDelete)

	return router
}
