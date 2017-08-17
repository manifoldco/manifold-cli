package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client"
	resClient "github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	appCmd := cli.Command{
		Name:  "app",
		Usage: "Manages an App in Manifold",
		Subcommands: []cli.Command{
			{
				Name:      "add",
				ArgsUsage: "[label]",
				Usage:     "Add a resource to an app in Manifold.",
				Flags: []cli.Flag{
					appFlag(),
				},
				Action: chain(ensureSession, loadDirPrefs, appAddCmd),
			},
		},
	}

	cmds = append(cmds, appCmd)
}

func appAddCmd(cliCtx *cli.Context) error {
	ctx := context.Background()
	args := cliCtx.Args()

	resourceLabel := ""
	if len(args) > 1 {
		return errs.NewUsageExitError(cliCtx, errs.ErrTooManyArgs)
	}
	if len(args) > 0 {
		resourceLabel = args[0]
		l := manifold.Label(resourceLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidResourceName)
		}
	}

	appName := cliCtx.String("app")
	if appName != "" {
		n := manifold.Name(appName)
		if err := n.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidAppName)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Marketplace client: %s", err), -1)
	}

	res, err := clients.FetchResources(ctx, marketplaceClient)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	resourceIdx, _, err := prompts.SelectResource(res, resourceLabel)
	resource := res[resourceIdx]

	if appName == "" {
		apps := fetchUniqueAppNames(res)
		_, appName, err = prompts.SelectCreateAppName(apps, appName, false)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select app")
		}
	}

	err = addAppToResource(ctx, cfg, resource, marketplaceClient, appName)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to add app to resource: %s", err), -1)
	}

	fmt.Printf("Your resource \"%s\" has been added to the app \"%s\"\n", resource.Body.Label, appName)

	return nil
}

func addAppToResource(ctx context.Context, cfg *config.Config, resource *models.Resource,
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
