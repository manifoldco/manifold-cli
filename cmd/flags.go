package main

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/placeholder"
)

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

func skipFlag() cli.Flag {
	return cli.BoolFlag{
		Name:   "no-wait, w",
		Usage:  "Do not wait when creating, updating, or deleting a resource",
		EnvVar: "MANIFOLD_DONT_WAIT",
	}
}
