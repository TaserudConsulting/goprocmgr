package main

import (
	"flag"
)

func main() {
	var config Config
	var configFile string
	var addFlag string
	var listFlag bool
	var serveFlag bool
	var removeFlag string
	var startFlag string

	runner := Runner{config: &config}
	serve := Serve{config: &config, runner: &runner}
	cli := Cli{config: &config}

	flag.StringVar(&configFile, "config", config.GuessFileName(""), "Specify config file")
	flag.BoolVar(&serveFlag, "serve", false, "Run the serve command")
	flag.BoolVar(&listFlag, "list", false, "List the stored servers")
	flag.StringVar(&addFlag, "add", "", "Add a new server, will capture the current directory and environment and then takes the command as an argument")
	flag.StringVar(&removeFlag, "remove", "", "Remove an existing server by it's name")
	flag.StringVar(&startFlag, "start", "", "Start an existing server by it's name")
	flag.Parse()

	config.Read(configFile)

	switch true {
	case serveFlag:
		serve.Run()

	case listFlag:
		cli.List()

	case len(addFlag) > 0:
		cli.Add(addFlag)

	case len(removeFlag) > 0:
		cli.Remove(removeFlag)

	case len(startFlag) > 0:
		cli.Start(startFlag)
	}
}
