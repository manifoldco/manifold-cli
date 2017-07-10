package main

import (
	"os"

	"github.com/urfave/cli"
)

var cmds []cli.Command

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		versionLookup(ctx)
	}

	app := cli.NewApp()
	app.Name = "manifold-cli"
	app.HelpName = "manifold-cli"
	app.Usage = "A tool making it easy to buy, manage, and integrate developer services into an application."
	app.Version = version
	app.Commands = cmds

	app.Run(os.Args)
}
