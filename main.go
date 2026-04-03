package main

import (
	"fmt"
	"os"

	"github.com/abushady/tdo/cmd"
)

func main() {
	args, err := cmd.RewriteIDArgs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Args = args

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
