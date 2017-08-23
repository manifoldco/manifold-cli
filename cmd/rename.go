package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
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
	renameCmd := cli.Command{
		Name:      "rename",
		ArgsUsage: "[name] [new-name]",
		Usage:     "Rename a resource label",
		Flags: []cli.Flag{
			appFlag(),
			nameFlag(),
		},
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs, rename),
	}

	cmds = append(cmds, renameCmd)
}

func rename(cliCtx *cli.Context) error {
	ctx := context.Background()
	args := cliCtx.Args()

	appName, err := validateName(cliCtx, "app")
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return errs.NewErrorExitError("Could not load config: %s", err)
	}

	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to create Maketplace API client: %s", err), -1)
	}

	resourceLabel := ""
	if len(args) > 0 {
		resourceLabel = args[0]
		l := manifold.Label(resourceLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidResourceName)
		}
	}

	newResourceLabel := ""
	if len(args) > 1 {
		newResourceLabel = args[1]
		l := manifold.Label(newResourceLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidResourceName)
		}
	}

	res, err := clients.FetchResources(ctx, marketplaceClient)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch list of provisioned resources: %s", err), -1)
	}

	if len(res) == 0 {
		return cli.NewExitError("No resources found to rename", -1)
	}

	var resource *models.Resource
	autoSelect := false
	if newResourceLabel != "" {
		resource, err = pickResourcesByLabel(res, resourceLabel)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf(
				"Unable to find resource named \"%s\": %s", resourceLabel, err), -1)
		}
		autoSelect = true
	} else {
		res = filterResourcesByAppName(res, appName)
		if appName != "" && len(res) == 0 {
			return cli.NewExitError(fmt.Sprintf("No resources in the app \"%s\"", appName), -1)
		}
		resourceIdx, _, err := prompts.SelectResource(res, resourceLabel)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}

		resource = res[resourceIdx]
	}
	newResourceLabel, err = prompts.ResourceLabel(newResourceLabel, autoSelect)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not get new resource label")
	}

	if _, err := pickResourcesByLabel(res, newResourceLabel); err == nil {
		return cli.NewExitError("A resource with that label already exists", -1)
	}

	if _, err := prompts.Confirm(
		fmt.Sprintf("Are you sure you want to rename \"%s\" to \"%s\"", resource.Body.Label, newResourceLabel),
	); err != nil {
		return cli.NewExitError("Resource not renamed", -1)
	}

	updatedRes, err := relabelResource(ctx, resource, marketplaceClient, newResourceLabel)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to update resource: %s", err), -1)
	}

	fmt.Printf("Your instance \"%s\" has been renamed\n", updatedRes.Body.Label)
	return nil
}

func relabelResource(ctx context.Context, resource *models.Resource,
	marketplaceClient *client.Marketplace, resourceName string,
) (*models.Resource, error) {
	label := strings.Replace(strings.ToLower(resourceName), " ", "-", -1)
	mLabel := manifold.Label(label)
	if err := mLabel.Validate(nil); err != nil {
		return nil, err
	}

	rename := &models.PublicUpdateResource{
		Body: &models.PublicUpdateResourceBody{
			Label: manifold.Label(label),
		},
	}

	c := resClient.NewPatchResourcesIDParamsWithContext(ctx)
	c.SetBody(rename)
	c.SetID(resource.ID.String())

	patchRes, err := marketplaceClient.Resource.PatchResourcesID(c, nil)
	if err != nil {
		if err != nil {
			switch e := err.(type) {
			case *resClient.PatchResourcesIDBadRequest:
				return nil, e.Payload
			case *resClient.PatchResourcesIDUnauthorized:
				return nil, e.Payload
			case *resClient.PatchResourcesIDInternalServerError:
				return nil, errs.ErrSomethingWentHorriblyWrong
			}
		}
	}

	return patchRes.Payload, nil
}
