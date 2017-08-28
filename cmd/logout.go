package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	logoutCmd := cli.Command{
		Name:   "logout",
		Usage:  "Log out of your account",
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
