package main

import (
	"flag"
)

func main() {
	var serveFlag bool

	flag.BoolVar(&serveFlag, "serve", false, "Run the serve command")
	flag.Parse()

	switch true {
	case serveFlag:
		var serve Serve

		serve.Run()
	}
}
