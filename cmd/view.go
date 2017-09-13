package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/rhymond/go-money"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
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

	resourceLabel, err := optionalArgLabel(cliCtx, 0, "resource")
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

	catalogClient, err := loadCatalogClient()
	if err != nil {
		return err
	}

	pClient, err := loadProvisioningClient()
	if err != nil {
		return err
	}

	// Get catalog
	catalog, err := catalog.New(ctx, catalogClient)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Get resources
	resources, err := clients.FetchResources(ctx, marketplaceClient, teamID, project)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}
	if len(resources) == 0 {
		return errs.ErrNoResources
	}

	// Get operations
	oRes, err := clients.FetchOperations(ctx, pClient, teamID)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of operations: "+err.Error(), -1)
	}

	resources, statuses := buildResourceList(resources, oRes)

	var resource *mModels.Resource
	if resourceLabel != "" {
		resource, err = pickResourcesByLabel(resources, resourceLabel)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Failed to find resource \"%s\": %s", resourceLabel, err), -1)
		}
	} else {
		idx, _, err := prompts.SelectResource(resources, resourceLabel)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		resource = resources[idx]
	}

	bold := color.New(color.Bold).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()
	productName := faint("-")
	planName := faint("-")
	regionName := faint("-")
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
			cli.NewExitError("Plan referenced by resource does not exist: "+
				err.Error(), -1)
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
		green := color.New(color.FgGreen).SprintFunc()
		status = green("Ready")
	}

	projectID := resource.Body.ProjectID
	projectLabel := "-"
	if projectID != nil {
		project, err := clients.FetchProject(ctx, marketplaceClient, projectID.String())
		if err != nil {
			cli.NewExitError("Project referenced by resource does not exist: "+
				err.Error(), -1)
		}
		projectLabel = string(project.Body.Label)
	}

	fmt.Println("Use `manifold update [label] --project [project]` to edit your resource")
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Name"), bold(resource.Body.Name)))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Label"), resource.Body.Label))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Project"), projectLabel))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("State"), status))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Custom"), isCustom))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Product"), productName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Plan"), planName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Region"), regionName))
	w.Flush()

	return nil
}
