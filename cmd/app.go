package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client"
	resClient "github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	appCmd := cli.Command{
		Name:  "apps",
		Usage: "Manages an App in Manifold",
		Subcommands: []cli.Command{
			{
				Name:      "add",
				ArgsUsage: "[label]",
				Usage:     "Add a resource to an app in Manifold.",
				Flags: []cli.Flag{
					appFlag(),
				},
				Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs, appAddCmd),
			},
			{
				Name:      "delete",
				ArgsUsage: "[label]",
				Usage:     "Removes a resource from an app in Manifold.",
				Flags: []cli.Flag{
					appFlag(),
				},
				Action: middleware.Chain(middleware.EnsureSession, deleteAppCmd),
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func appAddCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	resourceLabel, err := optionalArgLabel(cliCtx, 0, "resource")
	if err != nil {
		return err
	}

	appName, err := validateName(cliCtx, "app")
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Marketplace client: %s", err), -1)
	}

	resource, res, err := getResource(ctx, resourceLabel, marketplaceClient, false)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	if appName == "" {
		apps := fetchUniqueAppNames(res)
		_, appName, err = prompts.SelectCreateAppName(apps, appName, false)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select app")
		}
	}

	err = updateResourceApp(ctx, cfg, resource, marketplaceClient, appName)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to add app to resource: %s", err), -1)
	}

	fmt.Printf("Your resource \"%s\" has been added to the app \"%s\"\n", resource.Body.Label, appName)

	return nil
}

func deleteAppCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	resourceLabel, err := optionalArgLabel(cliCtx, 0, "resource")
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Marketplace client: %s", err), -1)
	}

	resource, _, err := getResource(ctx, resourceLabel, marketplaceClient, true)
	if err != nil {
		return cli.NewExitError(err, -1)
	}

	err = updateResourceApp(ctx, cfg, resource, marketplaceClient, "")
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to remove app from resource \"%s\": %s", resource.Body.Label, err), -1)
	}

	fmt.Printf("Your resource \"%s\" has been removed from the app \"%s\"\n", resource.Body.Label,
		resource.Body.AppName)

	return nil
}

func getResource(ctx context.Context, resourceLabel string, marketplaceClient *client.Marketplace,
	withAppsOnly bool,
) (*models.Resource, []*models.Resource, error) {
	res, err := clients.FetchResources(ctx, marketplaceClient)
	if err != nil {
		return nil, nil, cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	if withAppsOnly {
		res, err = filterResourcesWithApp(res)
		if err != nil {
			return nil, nil, err
		}
	}

	resourceIdx, _, err := prompts.SelectResource(res, resourceLabel)
	if err != nil {
		return nil, nil, prompts.HandleSelectError(err, "Could not select Resource")
	}

	return res[resourceIdx], res, nil
}

func updateResourceApp(ctx context.Context, cfg *config.Config, resource *models.Resource,
	marketplaceClient *client.Marketplace, appName string,
) error {
	appAdd := &models.PublicUpdateResource{
		Body: &models.PublicUpdateResourceBody{
			AppName: &appName,
		},
	}

	c := resClient.NewPatchResourcesIDParamsWithContext(ctx)
	c.SetBody(appAdd)
	c.SetID(resource.ID.String())

	_, err := marketplaceClient.Resource.PatchResourcesID(c, nil)
	if err != nil {
		switch e := err.(type) {
		case *resClient.PatchResourcesIDBadRequest:
			return e.Payload
		case *resClient.PatchResourcesIDUnauthorized:
			return e.Payload
		case *resClient.PatchResourcesIDInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		}
	}

	return nil
}

func filterResourcesWithApp(resources []*models.Resource) ([]*models.Resource, error) {
	var rs []*models.Resource

	for _, r := range resources {
		if r.Body.AppName != "" {
			rs = append(rs, r)
		}
	}

	if len(rs) == 0 {
		return nil, errs.ErrNoApps
	}

	return rs, nil
}
