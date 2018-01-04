package main

import (
	"github.com/urfave/cli"
	"github.com/manifoldco/manifold-cli/middleware"

	"github.com/manifoldco/manifold-cli/cmd/stack"
)


func init() {
	stackCmd := cli.Command{
		Name: "stack",
		Usage: "Organize project resources into Stacks",
		Category: "RESOURCES",
		Subcommands: []cli.Command{
			{
				Name: "init",
				Usage: "Initialize a new stack.yaml",
				Flags: append(teamFlags, []cli.Flag{
					cli.StringFlag{
						Name: "project, p",
						Usage: "Set a project name in your stack",
					},
					cli.BoolFlag{
						Name: "generate, g",
						Usage: "Auto-generate a new stack.yaml based on existing resource",
					},
					yesFlag(),
				}...),
				Action: middleware.Chain(middleware.EnsureSession, middleware.LoadTeamPrefs, stack.Init),
			},
		},
	}

	cmds = append(cmds, stackCmd)
}

