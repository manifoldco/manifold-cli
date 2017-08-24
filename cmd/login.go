package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	loginCmd := cli.Command{
		Name:   "login",
		Usage:  "Allow a user to login to their account",
		Action: login,
	}

	cmds = append(cmds, loginCmd)
}

func login(_ *cli.Context) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	if s.Authenticated() {
		if s.FromEnvVars() {
			fmt.Printf("You are logged in using your Manifold environment " +
				"variables, hooray!\n")
			return nil
		}
		return errs.ErrAlreadyLoggedIn
	}

	email, err := prompts.Email("")
	if err != nil {
		return err
	}

	password, err := prompts.Password()
	if err != nil {
		return err
	}

	s, err = session.Create(ctx, cfg, email, password)
	if err != nil {
		return cli.NewExitError("Are you sure the password and email match? "+err.Error(), -1)
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	a.Track(ctx, "Logged In", nil)

	fmt.Printf("You are logged in, hooray!\n")
	return nil
}
