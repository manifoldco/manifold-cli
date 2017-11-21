package main

import (
	"fmt"
	"os"
	"strings"

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

// generateTitle makes a name capitalized and replace dashes with spaaces
func generateTitle(name string) manifold.Name {
	title := strings.Title(strings.Replace(name, "-", " ", -1))
	return manifold.Name(title)
}

// promptNameAndTitle encapsulates the logic for accepting a name as the first
// positional argument (optionally), and a title flag (optionally)
// returning generated values accepted by the user
func promptNameAndTitle(ctx *cli.Context, optionalID *manifold.ID, objectName string, shouldInferTitle, allowEmpty bool) (string, string, error) {
	// The user may supply a name value as the first positional arg
	argName, err := optionalArgName(ctx, 0, "name")
	if err != nil {
		return "", "", err
	}
	shouldAccept := true
	if optionalID != nil {
		_, name := names.New(*optionalID)
		if argName == "" {
			argName = string(name)
		}
		shouldAccept = false
	}

	// The user may supply a title value from a flag, validate it
	flagTitle := ctx.String("title")

	// Create the title based on the name argument
	return createNameAndTitle(ctx, objectName, argName, flagTitle, shouldInferTitle, shouldAccept, allowEmpty)
}

func createNameAndTitle(ctx *cli.Context, objectName, argName, flagTitle string, shouldInferTitle, shouldAccept, allowEmpty bool) (string, string, error) {
	var name, title string

	// If no name value is supplied, prompt for it
	// otherwise validate the given value
	shouldAcceptName := shouldAccept && argName != ""
	nameValue, err := prompts.Name(objectName, argName, shouldAcceptName, allowEmpty)
	if err != nil {
		return name, title, err
	}
	name = nameValue

	// We will automatically validate/accept a title given as flag
	shouldAcceptTitle := shouldAccept && flagTitle != ""
	// If we shouldn't infer the title, do not automatically accept a title value
	if !shouldInferTitle {
		shouldAcceptTitle = false
	}
	defaultTitle := flagTitle
	if flagTitle == "" && shouldInferTitle {
		// If no flag is present, we will infer the title from the validated name
		defaultTitle = string(generateTitle(name))
	}
	titleValue, err := prompts.Title(objectName, defaultTitle, shouldAcceptTitle, allowEmpty)
	if err != nil {
		return name, title, err
	}
	title = titleValue

	return name, title, nil
}
