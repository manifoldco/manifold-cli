package main

import (
	"context"
	"fmt"
	"os"

	"github.com/juju/ansiterm"
	"github.com/rhymond/go-money"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

func init() {
	viewCmd := cli.Command{
		Name:     "view",
		Usage:    "View specific details of the provided resource",
		Category: "RESOURCES",
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
			middleware.LoadTeamPrefs, view),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
		}...),
	}

	cmds = append(cmds, viewCmd)
}

func view(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	resourceName, err := optionalArgName(cliCtx, 0, "resource")
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

	client, err := api.New(api.Catalog, api.Marketplace, api.Provisioning)
	if err != nil {
		return err
	}

	// Get catalog
	catalog, err := catalog.New(ctx, client.Catalog)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Get resources
	resources, err := clients.FetchResources(ctx, client.Marketplace, teamID, project)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}
	if len(resources) == 0 {
		return errs.ErrNoResources
	}

	// Get operations
	oRes, err := clients.FetchOperations(ctx, client.Provisioning, teamID)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of operations: "+err.Error(), -1)
	}

	resources, statuses := buildResourceList(resources, oRes)

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	var resource *models.Resource
	if resourceName != "" {
		resource, err = pickResourcesByName(resources, resourceName)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Failed to find resource \"%s\": %s", resourceName, err), -1)
		}
	} else {
		idx, _, err := prompts.SelectResource(resources, projects, resourceName)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		resource = resources[idx]
	}

	productName := color.Faint("-")
	planName := color.Faint("-")
	regionName := color.Faint("-")
	isCustom := "Yes"

	if *resource.Body.Source != "custom" {
		isCustom = "No"

		// Get catalog data
		product, err := catalog.GetProduct(*resource.Body.ProductID)
		if err != nil {
			cli.NewExitError("Product referenced by resource does not exist: "+
				err.Error(), -1)
		}
		plan, err := catalog.GetPlan(*resource.Body.PlanID)
		if err != nil {
			// Try and get unlisted plan not in local cache
			plan, err = catalog.FetchPlanById(ctx, *resource.Body.PlanID)
			if err != nil {
				return cli.NewExitError("Plan referenced by resource does not exist: "+
					err.Error(), -1)
			}
		}
		region, err := catalog.GetRegion(*resource.Body.RegionID)
		if err != nil {
			cli.NewExitError("Region referenced by resource does not exist: "+
				err.Error(), -1)
		}

		productName = string(product.Body.Name)
		cost := money.New(*plan.Body.Cost, "USD").Display()
		planName = fmt.Sprintf("%s (%s/%s)", string(plan.Body.Name), cost, "month")
		regionName = string(region.Body.Name)
	}

	status, ok := statuses[resource.ID]
	if !ok {
		status = color.Color(ansiterm.Green, "Ready")
	}

	projectID := resource.Body.ProjectID
	projectName := "-"
	if projectID != nil {
		var project *models.Project
		for _, p := range projects {
			if p.ID == *projectID {
				project = p
			}
		}
		if project == nil {
			cli.NewExitError("Project referenced by resource does not exist: "+
				err.Error(), -1)
		}
		projectName = string(project.Body.Label)
	}

	fmt.Println("Use `manifold update [resource-name] --project [project]` to edit your resource")
	fmt.Println("")
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Name"), color.Bold(resource.Body.Label)))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Title"), color.Faint(resource.Body.Name)))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Project"), projectName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("State"), status))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Custom"), isCustom))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Product"), productName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Plan"), planName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", color.Faint("Region"), regionName))
	w.Flush()

	return nil
}
