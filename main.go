package main // import "github.com/TaserudConsulting/goprocmgr"

import (
	"flag"
	"fmt"
)

func main() {
	var config Config
	var configFile string
	var addFlag string
	var listFlag bool
	var listFormat string
	var versionFlag bool
	var serveFlag bool
	var removeFlag string
	var startFlag string
	var stopFlag string
	var logsFlag string

	flag.StringVar(&configFile, "config", config.GuessFileName(""), "Specify config file")
	flag.BoolVar(&serveFlag, "serve", true, "Run the serve command (start the web server)")
	flag.BoolVar(&listFlag, "list", false, "List the stored servers")
	flag.StringVar(&listFormat, "list-format", "table", "List format (table, csv) when using the list command")
	flag.BoolVar(&versionFlag, "version", false, "Print the version")
	flag.StringVar(&addFlag, "add", "", "Add a new server, will capture the current directory and environment and then takes the command as an argument")
	flag.StringVar(&removeFlag, "remove", "", "Remove an existing server by it's name")
	flag.StringVar(&startFlag, "start", "", "Start an existing server by it's name")
	flag.StringVar(&stopFlag, "stop", "", "Stop an existing server by it's name")
	flag.StringVar(&logsFlag, "logs", "", "Tail the logs from an existing server by it's name")
	flag.Parse()

	if versionFlag {
		fmt.Println("goprocmgr version %undefined-version%")
		return
	}

	runner := Runner{config: &config}
	serve := NewServe(&config, &runner)
	cli := Cli{config: &config}

	config.Read(configFile)

	switch true {
	case listFlag:
		cli.List(listFormat)

	case len(addFlag) > 0:
		cli.Add(addFlag)

	case len(removeFlag) > 0:
		cli.Remove(removeFlag)

	case len(startFlag) > 0:
		cli.Start(startFlag)

	case len(stopFlag) > 0:
		cli.Stop(stopFlag)

	case len(logsFlag) > 0:
		cli.Logs(logsFlag)

	case serveFlag:
		serve.Run()
	}
}
