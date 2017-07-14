package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/session"

	marketplaceModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
)

type resourcesSortByName []*marketplaceModels.Resource

func (r resourcesSortByName) Len() int {
	return len(r)
}
func (r resourcesSortByName) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r resourcesSortByName) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(fmt.Sprintf("%s", r[i].Body.Name)),
		fmt.Sprintf("%s", r[j].Body.Name)) < 0
}

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

	// Init catalog client
	catalogClient, err := clients.NewCatalog(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create the Catalog API client: "+
			err.Error(), -1)
	}

	// Get catalog
	catalog, err := catalog.New(ctx, catalogClient)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	// Init marketplace client
	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create the Marketplace API client: "+
			err.Error(), -1)
	}

	// Get resources
	res, err := marketplaceClient.Resource.GetResources(nil, nil)
	if err != nil {
		return cli.NewExitError("Failed to fetch the list of provisioned "+
			"resources: "+err.Error(), -1)
	}
	// Sort resources by name
	sort.Sort(resourcesSortByName(res.Payload))

	// Write out the resources table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, "Resource Name\tApp Name\tProduct\tPlan\tRegion")
	for _, resource := range res.Payload {
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
