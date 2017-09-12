package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	"github.com/manifoldco/manifold-cli/generated/marketplace/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

func init() {
	updateCmd := cli.Command{
		Name:      "update",
		ArgsUsage: "[label]",
		Usage:     "Update a resource",
		Category:  "RESOURCES",
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs,
			middleware.LoadTeamPrefs, updateResourceCmd),
		Flags: append(teamFlags, []cli.Flag{
			nameFlag(),
			projectFlag(),
		}...),
	}

	cmds = append(cmds, updateCmd)
}

func updateResourceCmd(cliCtx *cli.Context) error {
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

	project, err := validateLabel(cliCtx, "project")
	if err != nil {
		return err
	}

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	var resources []*models.Resource

	resources, err = clients.FetchResources(ctx, marketplaceClient, teamID, project)

	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	if len(resources) == 0 {
		return cli.NewExitError("No resources found for updating", -1)
	}

	var resource *models.Resource
	if label != "" {
		var err error
		resource, err = pickResourcesByLabel(resources, label)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to fetch resource: %s", err), -1)
		}
	} else {
		idx, _, err := prompts.SelectResource(resources, label)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		resource = resources[idx]
	}

	newName := cliCtx.String("name")
	name := string(resource.Body.Name)
	autoSelect := false
	if newName != "" {
		name = newName
		autoSelect = true
	}

	newName, err = prompts.ResourceName(name, autoSelect)
	if err != nil {
		cli.NewExitError(fmt.Sprintf("Could not rename the resource: %s", err), -1)
	}

	prompts.SpinStart(fmt.Sprintf("Updating resource %q", resource.Body.Label))

	mrb, err := updateResource(ctx, resource, marketplaceClient, newName)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to update resource: %s", err), -1)
	}

	prompts.SpinStop()

	fmt.Printf("Your instance \"%s\" has been updated\n", mrb.Body.Name)
	return nil
}

func pickResourcesByLabel(resources []*models.Resource, label string) (*models.Resource, error) {
	if label == "" {
		return nil, errs.ErrResourceNotFound
	}

	for _, resource := range resources {
		if string(resource.Body.Label) == label {
			return resource, nil
		}
	}

	return nil, errs.ErrResourceNotFound
}

func updateResource(ctx context.Context, r *models.Resource,
	marketplaceClient *client.Marketplace, resourceName string) (*models.Resource, error) {
	rename := &models.PublicUpdateResource{
		Body: &models.PublicUpdateResourceBody{
			Name:  manifold.Name(resourceName),
			Label: generateLabel(resourceName),
		},
	}

	c := resource.NewPatchResourcesIDParamsWithContext(ctx)
	c.SetBody(rename)
	c.SetID(r.ID.String())

	patchRes, err := marketplaceClient.Resource.PatchResourcesID(c, nil)
	if err != nil {
		switch e := err.(type) {
		case *resource.PatchResourcesIDBadRequest:
			return nil, e.Payload
		case *resource.PatchResourcesIDUnauthorized:
			return nil, e.Payload
		case *resource.PatchResourcesIDInternalServerError:
			return nil, errs.ErrSomethingWentHorriblyWrong
		}
	}

	return patchRes.Payload, err
}
