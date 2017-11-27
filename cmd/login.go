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

var errLoggedInEnv = cli.NewExitError("", 0)

func init() {
	loginCmd := cli.Command{
		Name:     "login",
		Usage:    "Log in to your account",
		Category: "ADMINISTRATIVE",
		Subcommands: []cli.Command{
			{
				Name:   "github",
				Usage:  "Log in using a GitHub account with the OAuth 2 Web Flow",
				Action: loginGitHubWeb,
			},
			{
				Name:      "github-token",
				Usage:     "Log in using GitHub account with a GitHub Personal Access Token",
				ArgsUsage: "[token]",
				Action:    loginGitHubToken,
			},
			{
				Name:   "github-basic",
				Usage:  "Log in using GitHub using BASIC authentication to manage tokens",
				Action: loginGitHubBasic,
			},
		},
		Action: loginWithEmail,
	}

	cmds = append(cmds, loginCmd)
}

func loginWithEmail(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, _, err := loadLogin(ctx)
	switch err {
	case errLoggedInEnv:
		return nil
	case nil:
	default:
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

	s, err := session.Create(ctx, cfg, email, password)
	if err != nil {
		return cli.NewExitError("Are you sure the password and email match? "+err.Error(), -1)
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		cli.NewExitError("A problem ocurred: "+err.Error(), -1)
	}

	a.Track(ctx, "Logged In", nil)

	return nil
}

func loginGitHubWeb(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, a, err := loadLogin(ctx)
	switch err {
	case errLoggedInEnv:
		return nil
	case nil:
	default:
		return err
	}

	err = githubWithCallback(ctx, cfg, a, createOAuthAuth)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	fmt.Println("You are logged in, hooray!")

	return nil
}

func loginGitHubToken(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, a, err := loadLogin(ctx)
	switch err {
	case errLoggedInEnv:
		return nil
	case nil:
	default:
		return err
	}

	if cliCtx.NArg() == 0 || cliCtx.NArg() > 1 {
		return cli.NewExitError("You must provide a token for use with login", -1)
	}

	args := cliCtx.Args()
	token := args[0]

	err = githubWithToken(ctx, cfg, a, token, createOAuthAuth)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to log in: %s", err), -1)
	}

	fmt.Println("You are logged in, hooray!")

	return nil
}

func loginGitHubBasic(cliCtx *cli.Context) error {
	ctx := context.Background()

	cfg, a, err := loadLogin(ctx)
	switch err {
	case errLoggedInEnv:
		return nil
	case nil:
	default:
		return err
	}

	err = githubWithUser(ctx, cfg, a, createOAuthAuth)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	fmt.Println("You are logged in, hooray!")

	return nil
}

func loadLogin(ctx context.Context) (*config.Config, *analytics.Analytics, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return nil, nil, cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	if s.Authenticated() {
		if s.FromEnvVars() {
			fmt.Printf("You are logged in using your Manifold environment " +
				"variables, hooray!\n")
			return nil, nil, errLoggedInEnv
		}
		return nil, nil, errs.ErrAlreadyLoggedIn
	}

	a, err := analytics.New(cfg, s)
	if err != nil {
		return nil, nil, cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	return cfg, a, nil
}
