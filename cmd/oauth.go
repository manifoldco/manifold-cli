package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/urfave/cli"
)

var (
	linkedOK = "Your Manifold account is now linked"
)

func init() {
	oauthCmd := cli.Command{
		Name:     "oauth",
		Usage:    "Change your account to authenticate with an OAuth provider",
		Category: "AUTHENTICATION",
		Subcommands: []cli.Command{
			{
				Name:   "github",
				Usage:  "Link to a GitHub account using the OAuth 2 Web Flow",
				Action: linkGitHubWeb,
			},
			{
				Name:      "github-token",
				Usage:     "Link to a GitHub account using a GitHub Personal Access Token",
				ArgsUsage: "[token]",
				Action:    linkGitHubToken,
			},
			{
				Name:   "github-basic",
				Usage:  "Link GitHub using BASIC authentication to manage tokens",
				Action: linkGitHubBasic,
			},
		},
		Action: middleware.Chain(middleware.EnsureSession, oauth),
	}

	cmds = append(cmds, oauthCmd)
}

func oauth(cliCtx *cli.Context) error {
	return cli.NewExitError("subcommands must be provided", -1)
}

func linkGitHubWeb(cliCctx *cli.Context) error {
	ctx := context.Background()
	cfg, a, err := loadConfigAndAnaltics()
	if err != nil {
		return err
	}

	err = githubWithCallback(ctx, cfg, a, linkOAuthAuth)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to link accounts: %s", err), -1)
	}

	fmt.Println(linkedOK)

	return nil
}

func linkGitHubToken(cliCtx *cli.Context) error {
	ctx := context.Background()
	cfg, a, err := loadConfigAndAnaltics()
	if err != nil {
		return err
	}

	if cliCtx.NArg() == 0 || cliCtx.NArg() > 1 {
		return cli.NewExitError("You must provide a token for use with login", -1)
	}

	args := cliCtx.Args()
	token := args[0]

	err = githubWithToken(ctx, cfg, a, token, linkOAuthAuth)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to link accounts: %s", err), -1)
	}

	fmt.Println(linkedOK)

	return nil
}

func linkGitHubBasic(cliCctx *cli.Context) error {
	ctx := context.Background()
	cfg, a, err := loadConfigAndAnaltics()
	if err != nil {
		return err
	}

	err = githubWithUser(ctx, cfg, a, linkOAuthAuth)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to link accounts: %s", err), -1)
	}

	fmt.Println(linkedOK)

	return nil
}

func loadConfigAndAnaltics() (*config.Config, *analytics.Analytics, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, cli.NewExitError(fmt.Sprintf("Unable to load config: %s", err), -1)
	}

	a, err := api.New(api.Analytics)
	if err != nil {
		return nil, nil, cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	return cfg, a.Analytics, nil
}
