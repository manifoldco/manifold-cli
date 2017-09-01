package main

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/placeholder"
)

var teamFlags = []cli.Flag{teamFlag(), teamIDFlag(), meFlag()}

func formatFlag(defaultValue, description string) cli.Flag {
	return placeholder.New("format, f", "FORMAT", description, defaultValue, "MANIFOLD_FORMAT", false)
}

func nameFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "name, n",
		Usage:  "Specify a name for a resource",
		Value:  "",
		EnvVar: "MANIFOLD_NAME",
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

func appFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "app, a",
		Usage:  "Specify an app for filtering and updating",
		Value:  "",
		EnvVar: "MANIFOLD_APP",
	}
}

func planFlag() cli.Flag {
	return cli.StringFlag{
		Name:   "plan, p",
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

func skipFlag() cli.Flag {
	return cli.BoolFlag{
		Name:   "no-wait, w",
		Usage:  "Do not wait when creating, updating, or deleting a resource",
		EnvVar: "MANIFOLD_DONT_WAIT",
	}
}
