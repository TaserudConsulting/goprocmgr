package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Cli struct {
	config *Config
}

func (cli *Cli) List() {
	var config Config

	// Build URL based on config
	requestUrl := fmt.Sprintf("http://%s:%d/api/config", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	// Do request to running instance of program
	res, err := http.Get(requestUrl)

	if err != nil {
		log.Printf("Failed to connect to running instance of program: %s\n", err)
		os.Exit(1)
	}

	// Validate status code
	if res.StatusCode != 200 {
		log.Printf("Unexpected status code when fetching active config: %d\n", res.StatusCode)
		os.Exit(2)
	}

	// Read the body content
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	// Parse the json
	json.Unmarshal(body, &config)

	output := table.NewWriter()
	output.SetOutputMirror(os.Stdout)
	output.AppendHeader(table.Row{"Name", "Directory", "Command", "Environment"})

	for _, val := range config.Servers {
		output.AppendRow([]interface{}{val.Name, val.Directory, val.Command, val.Environment})
	}

	output.Render()
}

func (cli *Cli) Add(command string) {
	// Build URL based on config to post to
	requestUrl := fmt.Sprintf("http://%s:%d/api/config/server", cli.config.Settings.ListenAddress, cli.config.Settings.ListenPort)

	directory, err := os.Getwd()

	if err != nil {
		log.Printf("Failed to get current working directory: %s\n", err)
		os.Exit(3)
	}

	// Build a new server config
	server := ServerConfig{
		Name:      filepath.Base(directory),
		Command:   command,
		Directory: directory,
		Environment: map[string]string{
			"PATH": os.Getenv("PATH"),
		},
	}

	// Encode the server config as bytes
	body, _ := json.Marshal(server)

	//Pass new buffer for request with URL to post.
	//This will make a post request and will share the JSON data
	res, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(body))

	// An error is returned if something goes wrong
	if err != nil {
		panic(err)
	}

	//Need to close the response stream, once response is read.
	//Hence defer close. It will automatically take care of it.
	defer res.Body.Close()

	//Check response code, if New user is created then read response.
	if res.StatusCode == http.StatusCreated {
		log.Println("Created")
	} else {
		var response map[string]string
		resbody, _ := ioutil.ReadAll(res.Body)

		// Parse the json
		json.Unmarshal(resbody, &response)

		//The status is not Created. print the error.
		log.Printf("Failed to create server with response: %s", resbody)
	}
}
