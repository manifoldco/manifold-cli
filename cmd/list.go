package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/go-manifold"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/middleware"

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
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
			middleware.LoadTeamPrefs, list),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
		}...),
	}

	cmds = append(cmds, listCmd)
}

func list(cliCtx *cli.Context) error {
	ctx := context.Background()

	projectLabel, err := validateLabel(cliCtx, "project")
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	catalogClient, err := loadCatalogClient()
	if err != nil {
		return err
	}

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	provisionClient, err := loadProvisioningClient()
	if err != nil {
		return err
	}

	// Get catalog
	catalog, err := catalog.New(ctx, catalogClient)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Get resources
	res, err := clients.FetchResources(ctx, marketplaceClient, teamID, projectLabel)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of provisioned "+
			"resources: "+err.Error(), -1)
	}

	// Get operations
	oRes, err := clients.FetchOperations(ctx, provisionClient, teamID)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of operations: "+err.Error(), -1)
	}

	resources, statuses := buildResourceList(res, oRes)

	// Sort resources by name
	sort.Sort(resourcesSortByName(resources))

	// Write out the resources table
	tw := ansiterm.NewTabWriter(os.Stdout, 0, 0, 8, ' ', 0)
	tw.SetForeground(ansiterm.Gray)

	fmt.Fprintln(tw, "Label\tType\tStatus")
	tw.SetForeground(ansiterm.Default)

	for _, resource := range resources {
		rType := "Custom"

		if *resource.Body.Source != "custom" {
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

			rType = fmt.Sprintf("%s %s", product.Body.Name, plan.Body.Name)
		}

		status, ok := statuses[resource.ID]
		if !ok {
			status = "Ready"
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\n", resource.Body.Label, rType, status)
	}
	tw.Flush()
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
