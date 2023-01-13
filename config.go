package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	configFileName string

	Settings struct {
		ListenAddress string `json:"listen_address"`
		ListenPort    uint   `json:"listen_port"`
	} `json:"settings"`
	Servers map[string]ServerConfig `json:"servers"`
}

type ServerConfig struct {
	Name        string            `json:"name"`
	Directory   string            `json:"cwd"`
	Command     string            `json:"cmd"`
	Environment map[string]string `json:"env"`
}

func (config *Config) Read(configFileName string) {
	// Store the config file name
	config.configFileName = configFileName

	// Set up default config struct.
	config.Settings.ListenAddress = "127.0.0.1"
	config.Settings.ListenPort = 6969

	// Init servers map
	if config.Servers == nil {
		config.Servers = make(map[string]ServerConfig)
	}

	if _, err := os.Stat(config.configFileName); err == nil {
		// Read the config file
		fileContent, err := ioutil.ReadFile(config.configFileName)

		if err != nil {
			log.Fatalf("File read error: %v", err)
		} else {
			log.Printf("Parsed config file: %s\n", config.configFileName)
		}

		// Parse config
		json.Unmarshal(fileContent, &config)
	} else {
		log.Printf("Using default values as config will store config in %s if any changes are made\n", config.configFileName)
	}
}

func (config *Config) Save() {
	log.Printf("Writing configuration file at %s\n", config.configFileName)

	encodedFile, _ := json.MarshalIndent(config, "", "    ")
	_ = ioutil.WriteFile(config.configFileName, encodedFile, 0640)
}

func (config *Config) WriteServer(server ServerConfig) error {
	if len(server.Name) == 0 {
		return fmt.Errorf("Server 'name' cannot be empty")
	}

	if len(server.Directory) == 0 {
		return fmt.Errorf("Server 'cwd' cannot be empty")
	}

	if len(server.Command) == 0 {
		return fmt.Errorf("Server 'cmd' cannot be empty")
	}

	// Store the sent server config to the config.
	config.Servers[server.Name] = server

	// Save the config to disk.
	config.Save()

	return nil
}

func (config *Config) GuessFileName(fileName string) string {
	if len(fileName) > 0 {
		return fileName
	}

	if os.Getenv("XDG_CONFIG_DIR") != "" {
		return os.Getenv("XDG_CONFIG_DIR") + "/goprocmgr.json"
	}

	if os.Getenv("HOME") != "" {
		return os.Getenv("HOME") + "/.config/goprocmgr.json"
	}

	return "/tmp/goprocmgr.json"
}
