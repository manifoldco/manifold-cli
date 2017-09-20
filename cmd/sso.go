package main

import (
	"context"
	"fmt"

	openProc "github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/connector/client"
	"github.com/manifoldco/manifold-cli/generated/connector/client/o_auth"
	conModels "github.com/manifoldco/manifold-cli/generated/connector/models"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	ssoCmd := cli.Command{
		Name:      "sso",
		ArgsUsage: "[label]",
		Usage:     "Get an SSO link for a resource",
		Category:  "RESOURCES",
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs, middleware.LoadTeamPrefs,
			ssoCmd),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
			openFlag(),
		}...),
	}

	cmds = append(cmds, ssoCmd)
}

func ssoCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	label, err := optionalArgLabel(cliCtx, 0, "resource")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	open := cliCtx.Bool("open")
	project, err := validateLabel(cliCtx, "project")
	if err != nil {
		return err
	}

	client, err := api.New(api.Marketplace, api.Connector)
	if err != nil {
		return err
	}

	res, err := clients.FetchResources(ctx, client.Marketplace, teamID, project)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load resources: %s", err), -1)
	}
	if len(res) == 0 {
		return errs.ErrNoResources
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load teams: %s", err), -1)
	}

	rIdx, _, err := prompts.SelectResource(res, projects, label)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select project")
	}
	r := res[rIdx]

	sso, err := getSSOLink(ctx, r, client.Connector)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get SSO link: %s", err), -1)
	}

	if open {
		if err := openProc.Start(*sso.RedirectURI); err != nil {
			cli.NewExitError(fmt.Sprintf("Could not load SSO in browser: %s", err), -1)
		}
	} else {
		fmt.Printf("Your SSO link for %s is %s\n", r.Body.Label, *sso.RedirectURI)
	}

	return nil
}

func getSSOLink(ctx context.Context, r *models.Resource, connector *client.Connector) (*conModels.AuthorizationCodeBody, error) {
	sso := &conModels.AuthCodeRequest{
		Body: &conModels.AuthCodeRequestBody{
			ResourceID: r.ID,
		},
	}

	c := o_auth.NewPostSsoParamsWithContext(ctx)
	c.SetBody(sso)

	ssoRes, err := connector.OAuth.PostSso(c, nil)
	if err != nil {
		return nil, err
	}

	return ssoRes.Payload.Body, nil
}
