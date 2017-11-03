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
		Name:     "login",
		Usage:    "Log in to your account",
		Category: "ADMINISTRATIVE",
		Flags:    githubAuthFlags(),
		Action:   login,
	}

	cmds = append(cmds, loginCmd)
}

func login(cliCtx *cli.Context) error {
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

	a, err := analytics.New(cfg, s)
	if err != nil {
		return cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	oAuth := false
	if cliCtx.NumFlags() > 0 {
		if cliCtx.Bool("github") {
			err := githubWithCallback(ctx, cfg, a, createOAuthAuth)
			if err != nil {
				return cli.NewExitError(err, -1)
			}

			oAuth = true
		}

		if cliCtx.Bool("github-user") {
			err := githubWithUser(ctx, cfg, a, createOAuthAuth)
			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Unable to log in: %s", err), -1)
			}

			oAuth = true
		}

		if cliCtx.IsSet("github-token") {
			token := cliCtx.String("github-token")
			err := githubWithToken(ctx, cfg, a, token, createOAuthAuth)
			if err != nil {
				return cli.NewExitError(fmt.Sprintf("Unable to log in: %s", err), -1)
			}

			oAuth = true
		}
	}

	if oAuth {
		fmt.Println("You are logged in, hooray!")
	} else {
		return loginWithEmail(ctx, cfg, s)
	}

	return nil
}

func loginWithEmail(ctx context.Context, cfg *config.Config, s session.Session) error {
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

	return nil
}
