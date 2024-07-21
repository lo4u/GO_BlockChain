package main

import (
	"GOPreject/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	pCLI := cli.CommandLine{}
	pCLI.Run()
}
