package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	"github.com/manifoldco/manifold-cli/generated/marketplace/client/credential"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

func init() {
	updateCmd := cli.Command{
		Name:      "alias",
		ArgsUsage: "[resource-name]",
		Usage:     "Rename credential keys to match your needs",
		Category:  "RESOURCES",
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs,
			middleware.LoadTeamPrefs, aliasCredentialCmd),
		Flags: append(teamFlags, []cli.Flag{
			titleFlag(),
			projectFlag(),
		}...),
	}

	cmds = append(cmds, updateCmd)
}

func aliasCredentialCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	name, err := optionalArgName(cliCtx, 0, "resource")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	project, err := validateName(cliCtx, "project")
	if err != nil {
		return err
	}

	client, err := api.New(api.Catalog, api.Marketplace)
	if err != nil {
		return err
	}

	catalog, err := catalog.New(ctx, client.Catalog)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	prompts.SpinStart("Fetching Resources")
	resourceResults, err := clients.FetchResources(ctx, client.Marketplace, teamID, project)
	prompts.SpinStop()
	if err != nil {
		return cli.NewExitError("Could not retrieve resources: "+err.Error(), -1)
	}
	var resources []*models.Resource
	for _, r := range resourceResults {
		if r.Body.Source != nil && *r.Body.Source != "custom" {
			resources = append(resources, r)
		}
	}
	if len(resources) == 0 {
		return cli.NewExitError("No resources found", -1)
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	var selectedResource *models.Resource
	if name != "" {
		var err error
		res, err := pickResourcesByName(resources, projects, name)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to fetch resource: %s", err), -1)
		}
		selectedResource = res
	} else {
		idx, _, err := prompts.SelectResource(resources, projects, name)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		res := resources[idx]
		selectedResource = res
	}
	if selectedResource.Body.Source == nil || *selectedResource.Body.Source == "custom" {
		cli.NewExitError("Custom resources should be managed with `manifold config`", -1)
	}

	product, err := catalog.GetProduct(*selectedResource.Body.ProductID)
	if err != nil {
		cli.NewExitError("Product referenced by resource does not exist: "+
			err.Error(), -1)
	}

	prompts.SpinStart("Fetching Credentials")
	cMap, err := fetchCredentials(ctx, client.Marketplace, []*models.Resource{selectedResource}, false)
	prompts.SpinStop()
	if err != nil {
		return cli.NewExitError("Could not retrieve credentials: "+err.Error(), -1)
	}
	var creds []*models.Credential
	for _, cred := range cMap {
		creds = append(creds, cred...)
	}
	if len(creds) < 1 {
		return cli.NewExitError("No credentials found to be aliased", -1)
	}

	fmt.Printf("Choose from one of %s's credentials\n", product.Body.Name)
	cred, originalName, aliasName, err := prompts.CredentialAlias(creds)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not alias Credential")
	}

	alias := make(map[string]string)
	alias[originalName] = aliasName
	params := credential.NewPatchCredentialsIDParamsWithContext(ctx)
	params.SetID(cred.ID.String())
	params.SetBody(&models.UpdateCredential{
		Body: &models.UpdateCredentialBody{
			CustomNames: alias,
		},
	})
	_, err = client.Marketplace.Credential.PatchCredentialsID(params, nil)
	if err != nil {
		return cli.NewExitError("Could not create alias: "+err.Error(), -1)
	}

	fmt.Println("")
	if originalName == aliasName {
		fmt.Printf("You have cleared the alias for %s's `%s` key.\n", color.Bold(selectedResource.Body.Label), color.Bold(originalName))
	} else {
		fmt.Printf("You have aliased %s's config key `%s` to `%s`.\n", color.Bold(selectedResource.Body.Label), color.Bold(originalName), color.Bold(aliasName))
	}
	return nil
}
