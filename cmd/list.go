package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	listCmd := cli.Command{
		Name: "list",
		Usage: "Allows a user to list the status of their provisioned Manifold " +
			"resources.",
		Action: list,
	}

	cmds = append(cmds, listCmd)
}

func list(_ *cli.Context) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}
	if !s.Authenticated() {
		return errNotLoggedIn
	}

	// Get catalog
	catalog, err := GenerateCatalog(ctx, cfg, nil)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Get resources
	resources, err := GenerateResourceCache(ctx, cfg, nil)
	if err != nil {
		return cli.NewExitError("Failed to fetch resource data: "+err.Error(), -1)
	}

	// Write out the resources table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	// TODO: Make table prettier
	fmt.Fprintln(w, "Resource Name\tApp Name\tProduct\tPlan\tRegion")
	for _, resource := range resources.resources {
		appName := string(resource.Body.AppName)
		if appName == "" {
			appName = "None"
		}

		// Get catalog data
		product, err := catalog.GetProduct(resource.Body.ProductID.String())
		if err != nil {
			cli.NewExitError("Product referenced by resource does not exist: "+
				err.Error(), -1)
		}
		plan, err := catalog.GetPlan(resource.Body.PlanID.String())
		if err != nil {
			cli.NewExitError("Plan referenced by resource does not exist: "+
				err.Error(), -1)
		}
		region, err := catalog.GetRegion(resource.Body.RegionID.String())
		if err != nil {
			cli.NewExitError("Region referenced by resource does not exist: "+
				err.Error(), -1)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", resource.Body.Name,
			appName, product.Body.Name, plan.Body.Name, region.Body.Name)
	}
	w.Flush()
	return nil
}
