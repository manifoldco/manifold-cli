package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var version = "dev"

func init() {
	versionCmd := cli.Command{
		Name:   "version",
		Usage:  "Display version of this tool",
		Action: versionLookup,
	}

	cmds = append(cmds, versionCmd)
}

func versionLookup(ctx *cli.Context) error {
	fmt.Printf("%s\n", version)

	return nil
}
