package main // import "github.com/TaserudConsulting/goprocmgr"

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Serve struct {
	config              *Config
	runner              *Runner
	stateChange         chan bool                       // Channel to signal a state change
	clientSubscriptions map[*websocket.Conn]string      // Map of client connections and their subscriptions
	clientLocks         map[*websocket.Conn]*sync.Mutex // Map of locks for each client connection to not send multiple messages at once
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

func NewServe(config *Config, runner *Runner) *Serve {
	return &Serve{
		config:              config,
		runner:              runner,
		stateChange:         make(chan bool),
		clientSubscriptions: make(map[*websocket.Conn]string),
		clientLocks:         make(map[*websocket.Conn]*sync.Mutex),
	}
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
		serve.runner.Stop(server.Name, serve)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodPost)

	// Method to delete servers.
	router.HandleFunc("/api/config/server/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var resp ServeMessageResponse

		err := serve.runner.Stop(vars["name"], serve)

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

		err := serve.runner.Start(vars["name"], serve)

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

		err := serve.runner.Stop(vars["name"], serve)

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

		serve.clientSubscriptions[conn] = ""    // Initialize with no subscription
		serve.clientLocks[conn] = &sync.Mutex{} // Initialize mutex for this connection

		go func() {
			defer func() {
				conn.Close()
				delete(serve.clientSubscriptions, conn)
				delete(serve.clientLocks, conn) // Remove mutex for this connection
			}()

			// Send initial list state on connect to the client
			listState := serve.GetServerList()
			serve.sendMessage(conn, listState)

			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("ReadMessage:", err)
					break
				}

				// Parse the subscription message
				var subscription map[string]string
				if err := json.Unmarshal(message, &subscription); err != nil {
					log.Println("Unmarshal:", err)
					continue
				}

				if name, ok := subscription["name"]; ok {
					serve.clientSubscriptions[conn] = name

					// Send initial state for the subscribed server
					serverState := serve.GetServerLogs(name)

					// Only send logs if there are any
					if len(serverState.Logs) > 0 {
						serve.sendMessage(conn, serverState)
					}
				}
			}
		}()

		// Store last send time to avoid sending too many messages per second
		var lastSend int64

		for range serve.stateChange {
			// Limit the amount of messages sent per second to send at most every 100ms
			// to avoid flooding the client with messages.
			if ((time.Now().UnixNano() / int64(time.Millisecond)) - 100) < lastSend {
				continue
			}

			// Store the last send time
			lastSend = time.Now().UnixNano() / int64(time.Millisecond)

			for client, name := range serve.clientSubscriptions {
				// Send the list state regardless of subscription
				listState := serve.GetServerList()
				serve.sendMessage(client, listState)

				// Skip clients with no subscription
				if name == "" {
					continue
				}

				// Send state for the subscribed server
				serverState := serve.GetServerLogs(name)

				// Only send logs if there are any
				if len(serverState.Logs) > 0 {
					serve.sendMessage(client, serverState)
				}
			}
		}
	}).Methods(http.MethodGet)

	return router
}

func (serve *Serve) GetServer(name string) (ServerItem, error) {
	var serverItem ServerItem

	// Check if name is a valid entry in serve.config.Servers, if
	// it isn't, return error.
	if _, ok := serve.config.Servers[name]; !ok {
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

	if serverItemWithLogs.ServerItem.IsRunning {
		serverItemWithLogs.Logs = serve.runner.ActiveProcesses[name].Logs
	}

	return serverItemWithLogs
}

// Send a message to a client over a websocket connection
func (serve *Serve) sendMessage(client *websocket.Conn, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Println("Marshal:", err)
		return
	}

	serve.clientLocks[client].Lock()
	defer serve.clientLocks[client].Unlock()

	if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Println("WriteMessage:", err)
		client.Close()
		delete(serve.clientSubscriptions, client)
		delete(serve.clientLocks, client)
	}
}
