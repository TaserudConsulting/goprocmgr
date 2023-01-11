package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Settings struct {
		ListenAddress string `json:"listen_address"`
		ListenPort    uint   `json:"listen_port"`
	} `json:"settings"`
	Servers []struct {
		Directory   string            `json:"cwd"`
		Command     string            `json:"cmd"`
		Environment map[string]string `json:"environment"`
	} `json:"servers"`
}

func ParseConfig(configFileName string) Config {
	var parsedConfig Config

	// Set up default config struct.
	parsedConfig.Settings.ListenAddress = "127.0.0.1"
	parsedConfig.Settings.ListenPort = 6969

	// Write a default config file if it's missing.
	if _, err := os.Stat(GetConfigFileName(configFileName)); err != nil {
		log.Println("Creating default configuration file at " + GetConfigFileName(configFileName))

		encodedFile, _ := json.MarshalIndent(parsedConfig, "", " ")
		_ = ioutil.WriteFile(GetConfigFileName(configFileName), encodedFile, 0640)
	}

	// Read the config file
	fileContent, err := ioutil.ReadFile(GetConfigFileName(configFileName))
	if err != nil {
		log.Fatalf("File read error: %v", err)
	}

	// Parse config
	json.Unmarshal(fileContent, &parsedConfig)

	return parsedConfig
}

func GetConfigFileName(fileName string) string {
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
