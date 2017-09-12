package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

type resourcesSortByName []*models.Resource

func (r resourcesSortByName) Len() int {
	return len(r)
}
func (r resourcesSortByName) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r resourcesSortByName) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(fmt.Sprintf("%s", r[i].Body.Name)),
		fmt.Sprintf("%s", r[j].Body.Name)) > 0
}

func init() {
	listCmd := cli.Command{
		Name:     "list",
		Usage:    "List the status of your provisioned resources",
		Category: "RESOURCES",
		Action:   middleware.Chain(middleware.LoadDirPrefs, middleware.LoadTeamPrefs, list),
		Flags: append(teamFlags, []cli.Flag{
			appFlag(),
		}...),
	}

	cmds = append(cmds, listCmd)
}

func list(cliCtx *cli.Context) error {
	ctx := context.Background()

	appName, err := validateName(cliCtx, "app")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}
	if !s.Authenticated() {
		return errs.ErrNotLoggedIn
	}

	catalogClient, err := clients.NewCatalog(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Catalog API client: "+
			err.Error(), -1)
	}

	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Marketplace API client: "+
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
	res, err := clients.FetchResources(ctx, marketplaceClient, teamID, "")
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of provisioned "+
			"resources: "+err.Error(), -1)
	}

	// Get operations
	oRes, err := clients.FetchOperations(ctx, pClient, teamID)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of operations: "+err.Error(), -1)
	}

	resources, statuses := buildResourceList(res, oRes)

	// Sort resources by name and filter by given app name
	resources = filterResourcesByAppName(resources, appName)
	sort.Sort(resourcesSortByName(resources))

	// Write out the resources table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, "RESOURCE NAME\tAPP NAME\tSTATUS\tPRODUCT\tPLAN\tREGION\tCUSTOM")
	fmt.Fprintln(w, " \t \t \t \t \t \t")
	for _, resource := range resources {
		appName := string(resource.Body.AppName)

		productName := ""
		planName := ""
		regionName := ""
		isCustom := 'Y'

		if *resource.Body.Source != "custom" {
			isCustom = 'N'
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
			planName = string(plan.Body.Name)
			regionName = string(region.Body.Name)
		}

		status, ok := statuses[resource.ID]
		if !ok {
			status = "Ready"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%c\n", resource.Body.Name,
			appName, status, productName, planName, regionName, isCustom)
	}
	w.Flush()
	return nil
}

func buildResourceList(resources []*models.Resource, operations []*pModels.Operation) (
	[]*models.Resource, map[manifold.ID]string) {
	out := []*models.Resource{}
	statuses := make(map[manifold.ID]string)

	for _, op := range operations {
		switch op.Body.Type() {
		case "provision":
			body := op.Body.(*pModels.Provision)
			if body.State == nil {
				panic("State value was nil")
			}

			// if its a terminal state, then we can just ignore the op
			if *body.State == "done" || *body.State == "error" {
				continue
			}

			statuses[op.Body.ResourceID()] = "Creating"
			out = append(out, &models.Resource{
				ID: op.Body.ResourceID(),
				Body: &models.ResourceBody{
					AppName:   manifold.Name(body.AppName),
					CreatedAt: op.Body.CreatedAt(),
					UpdatedAt: op.Body.UpdatedAt(),
					Label:     manifold.Label(*body.Label),
					Name:      manifold.Name(*body.Name),
					Source:    body.Source,
					PlanID:    body.PlanID,
					ProductID: body.ProductID,
					RegionID:  body.RegionID,
					UserID:    op.Body.UserID(),
				},
			})
		case "resize":
			body := op.Body.(*pModels.Resize)
			if body.State == nil {
				panic("State value was nil")
			}

			if *body.State == "done" || *body.State == "error" {
				continue
			}

			statuses[op.Body.ResourceID()] = "Resizing"
		case "deprovision":
			body := op.Body.(*pModels.Deprovision)
			if body.State == nil {
				panic("State value was nil")
			}

			if *body.State == "done" || *body.State == "error" {
				continue
			}

			statuses[op.Body.ResourceID()] = "Deleting"
		}
	}

	for _, r := range resources {
		out = append(out, r)
	}

	return out, statuses
}
