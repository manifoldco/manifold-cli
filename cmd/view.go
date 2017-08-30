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
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

func init() {
	viewCmd := cli.Command{
		Name:   "view",
		Usage:  "View specific details of the provided resource",
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.LoadTeamPrefs, view),
		Flags: append(teamFlags, []cli.Flag{
			appFlag(),
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

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	_, err = session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not retrieve session: %s", err), -1)
	}

	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Maketplace Client: %s", err), -1)
	}

	catalogClient, err := clients.NewCatalog(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Catalog API client: "+
			err.Error(), -1)
	}

	pClient, err := clients.NewProvisioning(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Provisioning API Client: "+
			err.Error(), -1)
	}

	// Get catalog
	catalog, err := catalog.New(ctx, catalogClient)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Get resources
	res, err := clients.FetchResources(ctx, marketplaceClient, teamID, false)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}
	if len(res) == 0 {
		return errs.ErrNoResources
	}

	var resource *mModels.Resource
	if resourceLabel != "" {
		resource, err = pickResourcesByLabel(res, resourceLabel)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Failed to find resource \"%s\": %s", resourceLabel, err), -1)
		}
	} else {
		resourceIdx, _, err := prompts.SelectResource(res, resourceLabel)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		resource = res[resourceIdx]
	}

	// Get operations
	oRes, err := clients.FetchOperations(ctx, pClient, nil, false)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of operations: "+err.Error(), -1)
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

	_, statuses := buildResourceList(res, oRes)
	status, ok := statuses[resource.ID]
	if !ok {
		green := color.New(color.FgGreen).SprintFunc()
		status = green("Ready")
	}

	fmt.Println("Use `manifold update [label] --project [project]` to edit your resource")
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Name"), bold(resource.Body.Name)))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Label"), resource.Body.Label))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("App"), resource.Body.AppName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("State"), status))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Custom"), isCustom))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Product"), productName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Plan"), planName))
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s", faint("Region"), regionName))
	w.Flush()

	return nil
}
