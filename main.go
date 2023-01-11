package main

import (
	"flag"
)

func main() {
	var configFile string
	var parsedConfig Config
	var serveFlag bool

	flag.StringVar(&configFile, "config", GetConfigFileName(""), "Specify config file")
	flag.BoolVar(&serveFlag, "serve", false, "Run the serve command")
	flag.Parse()

	parsedConfig = ParseConfig(configFile)

	switch true {
	case serveFlag:
		var serve Serve

		serve.Run(&parsedConfig)
	}
}
