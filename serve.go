package main // import "github.com/TaserudConsulting/goprocmgr"

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Serve struct {
	config *Config
	runner *Runner
}

type ServeFullState struct {
	Config      *Config                            `json:"configs"`
	RunnerState map[string]ServeRunnerResponseItem `json:"runners"`
}

type ServerItem struct {
	Name        string `json:"name"`
	IsRunning   bool   `json:"is_running"`
	Port        uint   `json:"port"`
	StdoutCount uint   `json:"stdout_count"`
	StderrCount uint   `json:"stderr_count"`
}

type ServerItemWithLogs struct {
	ServerItem ServerItem `json:"server"`
	Logs       []LogEntry `json:"logs"`
}

type ServerItemList struct {
	Servers map[string]ServerItem `json:"servers"`
}

type ServeMessageResponse struct {
	Message string `json:"message"`
}

type ServeRunnerResponseItem struct {
	Name string     `json:"name"`
	Port uint       `json:"port"`
	Logs []LogEntry `json:"logs"`
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

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (serve *Serve) newRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	serveFile := func(fileName string, contentType string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Look for the environment variable GOPROCMGR_ALLOW_EXTERNAL_RESOURCES
			// to allow serving files from the filesystem.
			allowExternalResources := os.Getenv("GOPROCMGR_ALLOW_EXTERNAL_RESOURCES")

			if allowExternalResources == "1" {
				if _, err := os.Stat(fileName); err == nil {
					http.ServeFile(w, r, fileName)
					return
				}
			}

			file, _ := static.ReadFile(fileName)
			w.Header().Set("Content-Type", contentType)
			w.Write(file)
		}
	}

	//
	// Web server endpoints
	//
	router.HandleFunc("/", serveFile("static/index.html", "text/html"))
	router.HandleFunc("/web/favicon.png", serveFile("static/favicon.png", "image/png"))
	router.HandleFunc("/web/style.css", serveFile("static/style.css", "text/css"))
	router.HandleFunc("/web/script.js", serveFile("static/script.js", "application/javascript"))
	router.HandleFunc("/web/alpinejs-3.14.0.min.js", serveFile("static/alpinejs-3.14.0.min.js", "application/javascript"))

	//
	// Endpoints to manage the server configuration.
	//

	// Method to create new servers.
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

	// Method to delete servers.
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

	// Method to fetch all servers configurations
	router.HandleFunc("/api/config/server", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serve.config.Servers)
	}).Methods(http.MethodGet)

	//
	// Endpoints to manage running state of servers
	//

	// Endpoint to start a server
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

	// Endpoint to stop a server
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

	//
	// Endpoint to fetch an overview of the state of all servers
	//

	// Fetch state without logs for all servers
	router.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serve.GetServerList())
	}).Methods(http.MethodGet)

	// Fetch state and logs for a single server
	router.HandleFunc("/api/state/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serve.GetServerLogs(vars["name"]))
	}).Methods(http.MethodGet)

	//
	// Websocket endpoint to stream the state of the runner
	//
	router.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade:", err)
			return
		}
		defer conn.Close()

		var lastState []byte

		// Return runner state over the websocket
		for {
			var state = ServeFullState{
				Config:      serve.config,
				RunnerState: make(map[string]ServeRunnerResponseItem),
			}

			for key, value := range serve.runner.ActiveProcesses {
				state.RunnerState[key] = ServeRunnerResponseItem{
					Name: key,
					Port: value.Port,
					Logs: value.Logs,
				}
			}

			stateJson, err := json.Marshal(state)

			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}

			if string(stateJson) == string(lastState) {
				// Sleep a bit to then try again
				time.Sleep(time.Millisecond * 100)

				// Continue to next iteration
				continue
			}

			// Send the updated state
			conn.WriteMessage(websocket.TextMessage, stateJson)

			// Update last state
			lastState = stateJson
		}
	})

	return router
}

func (serve *Serve) GetServer(name string) (ServerItem, error) {
	var serverItem ServerItem

	// Check if name is a valid entry in serve.config.Servers, if
	// it isn't, return error.
	if _, ok := serve.config.Servers[name]; ok == false {
		return serverItem, errors.New("Undefined server requested '" + name + "'")
	}

	serverItem.Name = name
	serverItem.IsRunning = false

	if serve.runner.ActiveProcesses[name] != nil {
		serverItem.IsRunning = true
		serverItem.Port = serve.runner.ActiveProcesses[name].Port

		// Count the logs for each server by output
		for _, logEntry := range serve.runner.ActiveProcesses[name].Logs {
			if logEntry.Output == "stdout" {
				serverItem.StdoutCount++
			} else if logEntry.Output == "stderr" {
				serverItem.StderrCount++
			}
		}
	}

	return serverItem, nil
}

func (serve *Serve) GetServerList() ServerItemList {
	servers := ServerItemList{
		Servers: make(map[string]ServerItem),
	}

	// Go through all configured servers
	for serverName := range serve.config.Servers {
		server, err := serve.GetServer(serverName)

		if err != nil {
			log.Println(err)
			continue
		}

		// Add the server to the list
		servers.Servers[serverName] = server
	}

	return servers
}

func (serve *Serve) GetServerLogs(name string) ServerItemWithLogs {
	var serverItemWithLogs ServerItemWithLogs

	serverItemWithLogs.ServerItem, _ = serve.GetServer(name)

	if serverItemWithLogs.ServerItem.IsRunning == true {
		serverItemWithLogs.Logs = serve.runner.ActiveProcesses[name].Logs
	}

	return serverItemWithLogs
}
