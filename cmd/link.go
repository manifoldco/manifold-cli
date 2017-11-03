package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/urfave/cli"
)

func init() {
	linkCmd := cli.Command{
		Name:     "link",
		Usage:    "Link your an account to a third-party account",
		Category: "USER",
		Flags:    githubAuthFlags(),
		Action:   middleware.Chain(middleware.EnsureSession, link),
	}

	cmds = append(cmds, linkCmd)
}

func link(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to load config: %s", err), -1)
	}

	a, err := api.New(api.Analytics)
	if err != nil {
		return cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	if cliCtx.NumFlags() < 1 {
		return cli.NewExitError("You must provide a flag to specify an account to link to", -1)
	}

	linked := false
	if cliCtx.Bool("github") {
		err := githubWithCallback(ctx, cfg, a.Analytics, linkOAuthAuth)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Unable to link accounts: %s", err), -1)
		}
		linked = true
	}

	if cliCtx.Bool("github-user") {
		err := githubWithUser(ctx, cfg, a.Analytics, linkOAuthAuth)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Unable to link accounts: %s", err), -1)
		}
		linked = true
	}

	if cliCtx.String("github-token") != "" {
		token := cliCtx.String("github-token")
		err := githubWithToken(ctx, cfg, a.Analytics, token, linkOAuthAuth)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Unable to link accounts: %s", err), -1)
		}
		linked = true
	}

	if linked {
		fmt.Sprintf("Your Manifold account is not linked with %s\n\n", cliCtx.FlagNames())
	} else {
		return cli.NewExitError("No link type given, see --help", -1)
	}

	return nil
}
