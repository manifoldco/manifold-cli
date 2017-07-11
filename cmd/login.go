package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/config"
	"github.com/urfave/cli"
)

func init() {
	loginCmd := cli.Command{
		Name:   "login",
		Usage:  "Allows a user to login to their account",
		Action: login,
	}

	cmds = append(cmds, loginCmd)
}

func login(ctx *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	i, err := session.Retrieve(context.Background(), cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	fmt.Printf("Identity: %+v\n", i)
	return nil
}
