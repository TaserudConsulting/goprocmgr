package main

import (
	"flag"
)

func main() {
	var config Config
	var configFile string
	var serveFlag bool

	serve := Serve{config: &config}

	flag.StringVar(&configFile, "config", config.GuessFileName(""), "Specify config file")
	flag.BoolVar(&serveFlag, "serve", false, "Run the serve command")
	flag.Parse()

	config.Read(configFile)

	switch true {
	case serveFlag:
		serve.Run()
	}
}
