package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/plugins"
)

func init() {
	pluginsCmd := cli.Command{
		Name:  "plugins",
		Usage: "Manage installed plugins",
		Subcommands: []cli.Command{
			{
				Name:      "install",
				Usage:     "Install a new plugin",
				ArgsUsage: "[repository]",
				Action:    install,
			},
		},
	}

	cmds = append(cmds, pluginsCmd)
}

func install(cliCtx *cli.Context) error {
	pluginsDir, err := plugins.Path()
	if err != nil {
		return cli.NewExitError("Failed to install plugin: "+err.Error(), -1)
	}

	args := cliCtx.Args()
	if len(args) < 1 {
		return errs.NewUsageExitError(cliCtx, cli.NewExitError("Missing repository", -1))
	}

	// Identify the name of the plugin being installed
	name := plugins.DeriveName(args[0])
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}
	fmt.Printf("Installing plugin `%s`.\n", name)

	// Abort if already exists
	destinationDir := path.Join(pluginsDir, name)
	if _, err := os.Stat(destinationDir); !os.IsNotExist(err) {
		return cli.NewExitError("Plugin already installed.", -1)
	}

	// Clone to its destination
	pluginURL := plugins.NormalizeURL(args[0])
	cmd := exec.Command("git", "clone", pluginURL, destinationDir)
	err = cmd.Run()
	if err != nil {
		return cli.NewExitError("Failed to clone plugin: "+err.Error(), -1)
	}

	// Initialize the config file
	newFile, err := os.Create(path.Join(destinationDir, ".config.json"))
	if err != nil {
		return cli.NewExitError("Failed to create config file: "+err.Error(), -1)
	}
	newFile.Close()

	fmt.Println("Done.")
	fmt.Println("")

	// Finally output the Help text
	return plugins.Help(name)
}
