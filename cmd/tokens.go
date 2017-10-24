package main

import (
	"context"
	"fmt"
	"os"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/generated/identity/client/authentication"
	"github.com/manifoldco/manifold-cli/generated/identity/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	appCmd := cli.Command{
		Name:     "tokens",
		Usage:    "Manage your API tokens",
		Category: "AUTHENTICATION",
		Flags:    teamFlags,
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new token",
				Flags: teamFlags,
				Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
					middleware.LoadTeamPrefs, createTokenCmd),
			},
			{
				Name:  "list",
				Usage: "List existing tokens",
				Flags: teamFlags,
				Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
					middleware.LoadTeamPrefs, listTokensCmd),
			},
			{
				Name:  "delete",
				Usage: "Delete an existing token",
				Flags: append(teamFlags, yesFlag()),
				Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
					middleware.LoadTeamPrefs, deleteTokenCmd),
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func createTokenCmd(cliCtx *cli.Context) error {
	ctx := context.Background()
	uID, err := loadUserID(ctx)
	if err != nil {
		return err
	}

	tID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	client, err := api.New(api.Identity)
	if err != nil {
		return err
	}

	desc, err := prompts.TokenDescription()
	if err != nil {
		return prompts.HandleSelectError(err, "Failed to describe token")
	}

	role, err := prompts.SelectRole()
	if err != nil {
		return prompts.HandleSelectError(err, "Failed to select role")
	}

	var teamID, userID *manifold.ID
	if tID != nil {
		teamID = tID
	} else {
		userID = uID
	}
	params := authentication.NewPostTokensParamsWithContext(ctx).WithBody(&models.APITokenRequest{
		Description: &desc,
		Role:        models.RoleLabel(role),
		TeamID:      teamID,
		UserID:      userID,
	})
	resp, err := client.Identity.Authentication.PostTokens(params, nil)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not create token: %s", err), -1)
	}
	token := resp.Payload

	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Token"), color.Bold(token.Body.Token)))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Description"), *token.Body.Description))
	w.Flush()

	fmt.Println("")
	fmt.Println("Be sure to save your token in a safe place! We won't be able to show you it again.")
	return nil
}

func deleteTokenCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}
	tokenID, err := optionalArgID(cliCtx, 0, "token")
	if err != nil {
		return err
	}

	// If no token identified, prompt and select one
	if tokenID == nil {
		tokens, err := listTokens(ctx, cliCtx)
		if err != nil {
			return err
		}
		token, err := prompts.SelectAPIToken(tokens)
		if err != nil {
			return prompts.HandleSelectError(err, "Failed to select token")
		}
		tokenID = &token.ID
	}

	if !cliCtx.Bool("yes") {
		_, err = prompts.Confirm("Are you sure you want to revoke this token? It cannot be undone")
		if err != nil {
			return err
		}
	}

	client, err := api.New(api.Identity)
	if err != nil {
		return err
	}

	params := authentication.NewDeleteTokensTokenParamsWithContext(ctx).WithToken(tokenID.String())
	_, err = client.Identity.Authentication.DeleteTokensToken(params, nil)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not delete token: %s", err), -1)
	}

	fmt.Println("Your token has been revoked, you can now discard it.")
	return nil
}

func listTokensCmd(cliCtx *cli.Context) error {
	ctx := context.Background()
	tokens, err := listTokens(ctx, cliCtx)
	if err != nil {
		return err
	}

	fmt.Printf("%d tokens found\n", len(tokens))
	if len(tokens) == 0 {
		fmt.Println("Use `manifold tokens create` to issue a token")
		return nil
	}
	fmt.Println("Use `manifold tokens delete [token-id]` to revoke a token")
	fmt.Println("")
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)
	w.SetStyle(ansiterm.Bold)
	fmt.Fprintln(w, "ID\tToken\tDescription\tRole")
	w.Reset()
	for _, t := range tokens {
		token := fmt.Sprintf("%s****%s", *t.Body.FirstFour, *t.Body.LastFour)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID, token, *t.Body.Description, t.Body.Role)
	}
	w.Flush()

	return nil
}

func listTokens(ctx context.Context, cliCtx *cli.Context) ([]*models.APIToken, error) {
	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return nil, err
	}

	client, err := api.New(api.Identity)
	if err != nil {
		return nil, err
	}

	params := authentication.NewGetTokensParamsWithContext(ctx)
	params.SetType("api")
	if teamID != nil {
		teamIDStr := teamID.String()
		params.TeamID = &teamIDStr
	} else {
		me := true
		params.Me = &me
	}

	resp, err := client.Identity.Authentication.GetTokens(params, nil)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Could not list tokens: %s", err), -1)
	}

	return resp.Payload, nil
}
