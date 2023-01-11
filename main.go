package main // import "github.com/etu/goprocmgr"

import (
	"flag"

	"github.com/etu/goprocmgr/src/commands/serve"
)

func main() {
	var serveFlag bool

	flag.BoolVar(&serveFlag, "serve", false, "Run the serve command")
	flag.Parse()

	switch true {
	case serveFlag:
		serve.Serve()
	}
}
