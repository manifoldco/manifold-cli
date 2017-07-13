package main

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/placeholder"
)

func formatFlag(defaultValue, description string) cli.Flag {
	return placeholder.New("format, f", "FORMAT", description, defaultValue, "MANIFOLD_FORMAT", false)
}
