package main

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/config"
)

func init() {
	versionCmd := cli.Command{
		Name:   "version",
		Usage:  "Display version of this tool",
		Action: versionLookup,
	}

	cmds = append(cmds, versionCmd)
}

func versionLookup(ctx *cli.Context) error {
	fmt.Printf("%s\n", config.Version)

	return nil
}
