package main

import (
	"fmt"
	"os"

	"github.com/abushady/tdo/cmd"
)

func main() {
	os.Args = cmd.RewriteIDArgs(os.Args)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
