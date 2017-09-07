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
	accountCmd := cli.Command{
		Name:     "account",
		Usage:    "Create and use a manifold account",
		Category: "ADMINISTRATIVE",
		Subcommands: []cli.Command{
			{
				Name:   "signup",
				Usage:  "Create a new account",
				Action: signup,
			},
			{
				Name:   "login",
				Usage:  "Log in to your account",
				Action: login,
			},
			{
				Name:   "logout",
				Usage:  "Log out of your account",
				Action: logout,
			},
		},
	}

	cmds = append(cmds, accountCmd)
}

func signup(*cli.Context) error {
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

func login(*cli.Context) error {
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

func logout(*cli.Context) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	if !s.Authenticated() {
		return errs.ErrAlreadyLoggedOut
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return cli.NewExitError("Something went horribly wrong: "+err.Error(), -1)
	}

	a.Track(ctx, "Logout", nil)

	err = session.Destroy(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Failed to logout: "+err.Error(), -1)
	}

	fmt.Printf("You are now logged out!\n")
	return nil
}
