package main

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/plugins"
)

var cmds []cli.Command

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		versionLookup(ctx)
	}

	app := cli.NewApp()
	app.Name = "manifold"
	app.HelpName = "manifold"
	app.Usage = "A tool making it easy to buy, manage, and integrate developer services into an application."
	app.Version = config.Version
	app.Commands = cmds

	app.Action = func(cliCtx *cli.Context) error {
		// Show help if no arguments passed
		if len(os.Args) < 2 {
			cli.ShowAppHelp(cliCtx)
			return nil
		}

		// Execute plugin if installed
		cmd := os.Args[1]
		success, err := plugins.Run(cmd)
		if err != nil {
			return cli.NewExitError("Plugin error: "+err.Error(), -1)
		}
		if success {
			return nil
		}

		// Otherwise display global help
		cli.ShowAppHelp(cliCtx)
		fmt.Println("\nIf you were attempting to use a plugin, it may not be installed.")
		return nil
	}

	app.Run(os.Args)
}

func newSpinner(suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	if suffix != "" {
		s.Suffix = " " + suffix
	}
	return s
}
