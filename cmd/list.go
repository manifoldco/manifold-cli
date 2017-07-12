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
	catalog, err := GenerateCatalogCache(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Get resources
	resources, err := GenerateResourceCache(ctx, cfg, catalog)
	if err != nil {
		return cli.NewExitError("Failed to fetch resource data: "+err.Error(), -1)
	}

	// Write out the resources table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	// TODO: Make table prettier
	fmt.Fprintln(w, "Resource Name\tApp Name\tProduct\tPlan\tRegion")
	for _, resource := range resources.resources {
		appName := string(resource.Resource.AppName)
		if appName == "" {
			appName = "None"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", resource.Resource.Name,
			appName, resource.Product.Product.Name, resource.Plan.Name,
			resource.Region.Name)
	}
	w.Flush()
	return nil
}
