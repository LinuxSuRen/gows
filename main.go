package main

import (
	"os"

	"github.com/linuxsuren/gows/cli"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
