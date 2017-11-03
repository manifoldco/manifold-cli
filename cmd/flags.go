package main

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/placeholder"
)

var teamFlags = []cli.Flag{teamFlag(), teamIDFlag(), meFlag()}

func formatFlag(defaultValue, description string) cli.Flag {
	return placeholder.New("format, f", "FORMAT", description, defaultValue, "MANIFOLD_FORMAT", false)
}

func titleFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "title",
		Usage:  "Specify a title to be used",
		Value:  "",
		EnvVar: "MANIFOLD_TITLE",
	}
}

func descriptionFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "description, d",
		Usage:  "Specify a description",
		Value:  "",
		EnvVar: "MANIFOLD_DESCRIPTION",
	}
}

func meFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  "me, m",
		Usage: "Specify the action should not be done with a team",
	}
}

func teamFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "team, t",
		Usage:  "Specify a team",
		Value:  "",
		EnvVar: "MANIFOLD_TEAM",
	}
}

func teamIDFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "team-id",
		Hidden: true,
		Value:  "",
	}
}

func projectFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "project, p",
		Usage:  "Specify a project for filtering and updating",
		Value:  "",
		EnvVar: "MANIFOLD_PROJECT",
	}
}

func planFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "plan",
		Usage:  "Specify a plan",
		Value:  "",
		EnvVar: "MANIFOLD_PLAN",
	}
}

func regionFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "region",
		Usage:  "Use this region",
		EnvVar: "MANIFOLD_REGION",
	}
}

func resourceFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "resource, r",
		Usage: "Use this resource",
	}
}

func providerFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "provider",
		Usage: "Specify a provider",
	}
}

func productFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "product",
		Usage: "Specify a product",
	}
}

func yesFlag() cli.Flag {
	return cli.BoolFlag{
		Name:   "yes, y",
		Usage:  "Automatically respond y to confirm prompts",
		EnvVar: "MANIFOLD_NO_CONFIRM",
	}
}

func skipFlag() cli.Flag {
	return cli.BoolFlag{
		Name:   "no-wait, w",
		Usage:  "Do not wait when creating, updating, or deleting a resource",
		EnvVar: "MANIFOLD_DONT_WAIT",
	}
}

func openFlag() cli.Flag {
	return cli.BoolFlag{
		Name:   "open, o",
		Usage:  "Opens a browser for a URL instead of printing",
		EnvVar: "MANIFOLD_OPEN_BROWSER",
	}
}

func roleFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "role",
		Usage: "Specify a team role to be used",
	}
}

func githubAuthFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "github",
			Usage: "--github",
		},
		cli.BoolFlag{
			Name:  "github-user",
			Usage: "--github-user",
		},
		cli.StringFlag{
			Name:   "github-token",
			Usage:  "--github-token",
			EnvVar: "MANIFOLD_GITHUB_TOKEN",
		},
	}
}
