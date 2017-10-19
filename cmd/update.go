package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/api"
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
		ArgsUsage: "[resource-name]",
		Usage:     "Update a resource",
		Category:  "RESOURCES",
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs,
			middleware.LoadTeamPrefs, updateResourceCmd),
		Flags: append(teamFlags, []cli.Flag{
			titleFlag(),
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

	client, err := api.New(api.Marketplace)
	if err != nil {
		return err
	}

	var resources []*models.Resource

	resources, err = clients.FetchResources(ctx, client.Marketplace, teamID, project)

	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	if len(resources) == 0 {
		return cli.NewExitError("No resources found for updating", -1)
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	var resource *models.Resource
	if name != "" {
		var err error
		resource, err = pickResourcesByName(resources, name)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to fetch resource: %s", err), -1)
		}
	} else {
		idx, _, err := prompts.SelectResource(resources, projects, name)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		resource = resources[idx]
	}

	newTitle := cliCtx.String("title")
	title := string(resource.Body.Name)
	autoSelect := false
	if newTitle != "" {
		title = newTitle
		autoSelect = true
	}

	newTitle, err = prompts.ResourceName(title, autoSelect)
	if err != nil {
		cli.NewExitError(fmt.Sprintf("Could not rename the resource: %s", err), -1)
	}

	prompts.SpinStart(fmt.Sprintf("Updating resource %q", resource.Body.Label))

	mrb, err := updateResource(ctx, resource, client.Marketplace, newTitle)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to update resource: %s", err), -1)
	}

	prompts.SpinStop()

	fmt.Printf("Your instance \"%s\" has been updated\n", mrb.Body.Name)
	return nil
}

func pickResourcesByName(resources []*models.Resource, name string) (*models.Resource, error) {
	if name == "" {
		return nil, errs.ErrResourceNotFound
	}

	for _, resource := range resources {
		if string(resource.Body.Label) == name {
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
			Label: generateName(resourceName),
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
