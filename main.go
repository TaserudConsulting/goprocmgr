package main

import (
	"flag"
)

func main() {
	var configFile string
	var config Config
	var serveFlag bool

	flag.StringVar(&configFile, "config", config.GuessFileName(""), "Specify config file")
	flag.BoolVar(&serveFlag, "serve", false, "Run the serve command")
	flag.Parse()

	config.Read(configFile)

	switch true {
	case serveFlag:
		var serve Serve

		serve.Run(&config)
	}
}
