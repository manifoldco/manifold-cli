package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/config"
	"github.com/urfave/cli"
)

func init() {
	logoutCmd := cli.Command{
		Name:   "logout",
		Usage:  "Allows a user to logout of their Manifold session",
		Action: logout,
	}

	cmds = append(cmds, logoutCmd)
}

func logout(_ *cli.Context) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("You are already logged out: "+err.Error(), -1)
	}

	if !s.Authenticated() {
		return cli.NewExitError("You're already logged out!", -1)
	}

	err = session.Destroy(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Failed to logout: "+err.Error(), -1)
	}

	fmt.Printf("You are now logged out!\n")
	return nil
}
