package main // import "github.com/etu/goprocmgr"

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(
			os.Stderr,
			"Missing required first argument to know what to do\n"+
				"run the following command to get help:\n"+
				"$ "+os.Args[0]+" help",
		)

		os.Exit(1)
	}

	switch os.Args[1] {
	default:
		fmt.Fprintln(os.Stderr, "Argument \""+os.Args[1]+"\" not implemented")
		os.Exit(2)
	}
}
