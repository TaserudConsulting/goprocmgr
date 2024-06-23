package main // import "github.com/TaserudConsulting/goprocmgr"

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gorilla/websocket"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Cli struct {
	config *Config
}

func (cli *Cli) List(format string) {
	var state map[string]ServerConfig
	var runningState ServerItemList

	// Build URL based on config
	requestUrl := fmt.Sprintf("http://%s:%d/api/config/server", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	// Do request to running instance of program
	res, err := http.Get(requestUrl)

	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	defer res.Body.Close()

	// Validate status code
	if res.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code when fetching active config: %d\n", res.StatusCode)
		os.Exit(2)
	}

	// Read the body content
	body, _ := io.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(body, &state)

	// Build URL based on config
	requestUrl = fmt.Sprintf("http://%s:%d/api/state", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	// Get the state
	res, err = http.Get(requestUrl)

	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	defer res.Body.Close()

	// Validate status code
	if res.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code when fetching active config: %d\n", res.StatusCode)
		os.Exit(2)
	}

	// Read the body content
	body, _ = io.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(body, &runningState)

	// Extract keys and sort them
	var keys []string
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	switch format {
	case "table":
		output := table.NewWriter()
		output.SetOutputMirror(os.Stdout)
		output.AppendHeader(table.Row{"Name", "Running", "Directory", "Command"})

		for _, key := range keys {
			val := state[key]
			isRunning := false

			if _, ok := runningState.Servers[val.Name]; ok {
				isRunning = runningState.Servers[val.Name].IsRunning
			}

			output.AppendRow([]interface{}{val.Name, isRunning, val.Directory, val.Command})
		}

		output.Render()

	case "csv":
		output := csv.NewWriter(os.Stdout)
		defer output.Flush()

		output.Write([]string{"Name", "Running", "Directory", "Command"})

		for _, key := range keys {
			val := state[key]
			isRunning := false

			if _, ok := runningState.Servers[val.Name]; ok {
				isRunning = runningState.Servers[val.Name].IsRunning
			}

			output.Write([]string{val.Name, fmt.Sprintf("%t", isRunning), val.Directory, val.Command})
		}
	}
}

func (cli *Cli) Add(command string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/config/server", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	directory, err := os.Getwd()

	if err != nil {
		log.Printf("Failed to get current working directory: %s\n", err)
		os.Exit(3)
	}

	// Check if DIRENV_FILE is set or not
	useDirenv := false
	if _, ok := os.LookupEnv("DIRENV_FILE"); ok {
		useDirenv = true
	}

	// Build a new server config
	server := ServerConfig{
		Name:      filepath.Base(directory),
		Command:   command,
		Directory: directory,
		UseDirenv: useDirenv,
		Environment: map[string]string{
			"PATH": os.Getenv("PATH"),
		},
	}

	// Encode the server config as bytes
	body, _ := json.Marshal(server)

	// Pass new buffer for request with URL to post.
	// This will make a post request and will share the JSON data
	res, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(body))

	// An error is returned if something goes wrong
	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	// Need to close the response stream, once response is read.
	// Hence defer close. It will automatically take care of it.
	defer res.Body.Close()

	// Check response code, if New user is created then read response.
	if res.StatusCode == http.StatusCreated {
		log.Println("Created")
	} else {
		var response map[string]string
		resbody, _ := io.ReadAll(res.Body)

		// Parse the json
		json.Unmarshal(resbody, &response)

		// The status is not Created. print the error.
		log.Printf("Failed to create server with response: %s", resbody)
	}
}

func (cli *Cli) Remove(name string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/config/server/%s", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort, name)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodDelete, requestUrl, nil)

	if err != nil {
		log.Printf("Failed to create request: %s\n", err)
		os.Exit(3)
	}

	// Perform request
	res, err := client.Do(req)

	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	if res.StatusCode == http.StatusOK {
		log.Println("OK")
		os.Exit(0)
	}

	// Handle error
	var response map[string]string
	resbody, _ := io.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(resbody, &response)

	// The status is not Created. print the error.
	log.Printf("Failed to create server with response: %s", resbody)
	os.Exit(4)
}

func (cli *Cli) Start(name string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/runner/%s", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort, name)

	// Pass new buffer for request with URL to post.
	// This will make a post request and will share the JSON data
	res, err := http.Post(requestUrl, "application/json", nil)

	// An error is returned if something goes wrong
	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	// Check response code, if New user is created then read response.
	if res.StatusCode == http.StatusCreated {
		log.Println("Started")
	} else {
		var response map[string]string
		resbody, _ := io.ReadAll(res.Body)

		// Parse the json
		json.Unmarshal(resbody, &response)

		// The status is not Created. print the error.
		log.Printf("Failed to start server with response: %s", resbody)
	}
}

func (cli *Cli) Stop(name string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/runner/%s", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort, name)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodDelete, requestUrl, nil)

	if err != nil {
		log.Printf("Failed to create request: %s\n", err)
		os.Exit(3)
	}

	// Perform request
	res, err := client.Do(req)

	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		log.Println("OK")
		os.Exit(0)
	}

	// Handle error
	var response map[string]string
	resbody, _ := io.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(resbody, &response)

	// The status is not Created. print the error.
	log.Printf("Failed to stop server with response: %s", resbody)
	os.Exit(4)
}

func (cli *Cli) Logs(name string) {
	var logsMaxIndex int = -1

	// Build URL to establish websocket connection
	wsUrl := fmt.Sprintf("ws://%s:%d/api/ws", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	// Create a new websocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Printf("Failed to establish websocket connection: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Send subscription message for the specific server
	subscription := map[string]string{"name": name}
	subMsg, err := json.Marshal(subscription)
	if err != nil {
		log.Printf("Failed to marshal subscription message: %s\n", err)
		os.Exit(2)
	}

	err = conn.WriteMessage(websocket.TextMessage, subMsg)
	if err != nil {
		log.Printf("Failed to send subscription message: %s\n", err)
		os.Exit(3)
	}

	// Start loop to receive and process incoming websocket messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Failed to read websocket message: %s\n", err)
			os.Exit(4)
		}

		// Parse the JSON message
		var serverLogs ServerItemWithLogs
		err = json.Unmarshal(message, &serverLogs)
		if err != nil {
			log.Printf("Failed to unmarshal websocket message: %s\n", err)
			continue
		}

		// Check if the message is for the subscribed server
		if serverLogs.ServerItem.Name != name {
			continue
		}

		// Process the logs
		for key, val := range serverLogs.Logs {
			if key > logsMaxIndex {
				if val.Output == "stdout" {
					fmt.Println(val.Output, val.Timestamp.Format("15:04:05"), "|", val.Message)
				} else {
					fmt.Fprintln(os.Stderr, val.Output, val.Timestamp.Format("15:04:05"), "|", val.Message)
				}

				logsMaxIndex = key
			}
		}
	}
}
