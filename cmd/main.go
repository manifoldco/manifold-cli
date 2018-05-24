package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/names"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/plugins"
	"github.com/manifoldco/manifold-cli/prompts"
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
	app.Commands = append(cmds, helpCommand)
	app.Flags = append(app.Flags, cli.HelpFlag)
	app.EnableBashCompletion = true

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

// copied from urfave/cli so we can set the category
var helpCommand = cli.Command{
	Name:      "help",
	Usage:     "Shows a list of commands or help for one command",
	Category:  "UTILITY",
	ArgsUsage: "[command]",
	Action: func(c *cli.Context) error {
		args := c.Args()
		if args.Present() {
			return cli.ShowCommandHelp(c, args.First())
		}

		cli.ShowAppHelp(c)
		return nil
	},
}

// promptName encapsulates the logic for accepting a name as the first
// positional argument (optionally), and returning the appropriate value
func promptNameForResource(ctx *cli.Context, optionalID *manifold.ID,
	product manifold.Label) (string, error) {
	// The user may supply a name value as the first positional arg
	argName, err := optionalArgName(ctx, 0, "name")
	if err != nil {
		return "", err
	}
	shouldAccept := true
	if optionalID != nil {
		name := names.ForResource(product, *optionalID)
		if argName == "" {
			argName = string(name)
			shouldAccept = false
		}
	}

	// If no name value is supplied, prompt for it
	// otherwise validate the given value
	shouldAcceptName := shouldAccept && argName != ""
	return prompts.Name("resource", argName, shouldAcceptName, true)
}

// promptName encapsulates the logic for accepting a name as the first
// positional argument (optionally), and returning the appropriate value
func promptName(ctx *cli.Context, optionalID *manifold.ID, objectName string, allowEmpty bool) (string, error) {
	// The user may supply a name value as the first positional arg
	argName, err := optionalArgName(ctx, 0, "name")
	if err != nil {
		return "", err
	}
	shouldAccept := true
	if optionalID != nil {
		_, name := names.New(*optionalID)
		if argName == "" {
			argName = string(name)
			shouldAccept = false
		}
	}

	// If no name value is supplied, prompt for it
	// otherwise validate the given value
	shouldAcceptName := shouldAccept && argName != ""
	return prompts.Name(objectName, argName, shouldAcceptName, allowEmpty)
}
