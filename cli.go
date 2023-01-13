package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
