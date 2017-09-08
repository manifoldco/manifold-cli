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
	signupCmd := cli.Command{
		Name:     "signup",
		Usage:    "Create a new account",
		Category: "ADMINISTRATIVE",
		Action:   signup,
	}

	cmds = append(cmds, signupCmd)
}

func signup(_ *cli.Context) error {
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
		return errs.ErrAlreadyLoggedIn
	}

	name, err := prompts.FullName("")
	if err != nil {
		return err
	}

	email, err := prompts.Email("")
	if err != nil {
		return err
	}

	password, err := prompts.Password()
	if err != nil {
		return err
	}

	_, err = session.Signup(ctx, cfg, name, email, password)
	if err != nil {
		return cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	s, err = session.Create(ctx, cfg, email, password)
	if err != nil {
		return cli.NewExitError("Failed to login after signup. "+err.Error(), -1)
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	a.Track(ctx, "Logged In", nil)

	fmt.Println("Account created. You are logged in.")
	return nil
}
