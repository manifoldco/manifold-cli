package main

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	catalogcache "github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/analytics"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	resClient "github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

var errCannotFindResource = errors.New("Cannot find resource")

func init() {
	updateCmd := cli.Command{
		Name:      "update",
		ArgsUsage: "[label]",
		Usage:     "Allows a user to update a resource in Manifold",
		Action:    update,
		Flags: []cli.Flag{
			nameFlag(),
			appFlag(),
			planFlag(),
			skipFlag(),
		},
	}

	cmds = append(cmds, updateCmd)
}

func update(cliCtx *cli.Context) error {
	ctx := context.Background()
	args := cliCtx.Args()

	dontWait := cliCtx.Bool("no-wait")

	resourceLabel := ""

	if len(args) > 2 {
		return errs.NewUsageExitError(cliCtx, errs.ErrTooManyArgs)
	}

	if len(args) > 0 {
		resourceLabel = args[0]
		l := manifold.Label(resourceLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidResourceName)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load configuration: %s", err), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not retrieve session: %s", err), -1)
	}
	if !s.Authenticated() {
		return errs.ErrNotLoggedIn
	}

	marketplaceClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Marketplace client: %s", err), -1)
	}

	catalogClient, err := clients.NewCatalog(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Catalog client: %s", err), -1)
	}

	provisioningClient, err := clients.NewProvisioning(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Provisioning Client", err), -1)
	}

	res, err := marketplaceClient.Resource.GetResources(
		resClient.NewGetResourcesParamsWithContext(ctx), nil)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	resource, err := pickResourcesByLabel(res.Payload, resourceLabel)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to fetch resource: %s", err), -1)
	}

	newResourceName := cliCtx.String("name")
	if newResourceName != "" {
		n := manifold.Name(newResourceName)
		if err := n.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidResourceName)
		}
	} else {
		newResourceName, err = prompts.ResourceName(string(resource.Body.Name))
		if err != nil {
			cli.NewExitError(fmt.Sprintf("Could not rename the resource: %s", err), -1)
		}
	}

	catalog, err := catalogcache.New(ctx, catalogClient)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to catalog data: %s", err), -1)
	}

	planLabel := cliCtx.String("plan")
	if planLabel != "" {
		l := manifold.Label(planLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidPlanLabel)
		}
	}
	// TODO: Move this+filterRegionsForPlans+fetchUniqueAppNames from create.go into another file/package?
	plans := filterPlansByProductID(catalog.Plans(), resource.Body.ProductID)
	planIdx, _, err := prompts.SelectPlan(plans, planLabel)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select plan")
	}

	regions := filterRegionsForPlan(catalog.Regions(), plans[planIdx].Body.Regions)
	regionIdx, _, err := prompts.SelectRegion(regions)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select region")
	}

	appName := cliCtx.String("app")
	if appName != "" {
		n := manifold.Name(appName)
		if err := n.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidAppName)
		}
	} else {
		apps := fetchUniqueAppNames(res.Payload)
		// TODO: This auto-selects the app and doesn't let the user change it without the -a flag
		_, appName, err = prompts.SelectCreateAppName(apps, string(resource.Body.AppName))
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select apps")
		}
	}

	spin := spinner.New(spinner.CharSets[38], 500*time.Millisecond)
	if !dontWait {
		fmt.Printf("\nWe're starting to update the resource \"%s\". This may take some time, please wait!\n\n",
			resource.Body.Label,
		)
		spin.Start()
	}

	pOp, mrb, err := updateResource(ctx, cfg, s, resource, marketplaceClient, provisioningClient,
		plans[planIdx], regions[regionIdx], appName, newResourceName,
	)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to update resource: %s", err), -1)
	}
	_ = pOp
	_ = mrb

	if !dontWait {
		spin.Stop()
	}

	return nil
}

func pickResourcesByLabel(resources []*mModels.Resource, resourceLabel string) (*mModels.Resource, error) {
	if resourceLabel == "" {
		return nil, errCannotFindResource
	}

	for _, resource := range resources {
		if string(resource.Body.Label) == resourceLabel {
			return resource, nil
		}
	}

	return nil, errCannotFindResource
}

func updateResource(ctx context.Context, cfg *config.Config, s session.Session, resource *mModels.Resource,
	marketplaceClient *mClient.Marketplace, provisioningClient *pClient.Provisioning, plan *cModels.Plan,
	region *cModels.Region, appName, resourceName string,
) (*pModels.Operation, *mModels.ResourceBody, error) {
	a, err := analytics.New(cfg, s)
	if err != nil {
		return nil, nil, err
	}
	_ = a

	if appName != string(resource.Body.AppName) || resourceName != string(resource.Body.Name) {
		label := "" // re-generate
		rename := &mModels.PublicUpdateResource{
			Body: &mModels.PublicUpdateResourceBody{
				AppName: &appName,
				Label:   manifold.Label(label),
				Name:    manifold.Name(resourceName),
			},
		}

		c := resClient.NewPatchResourcesIDParamsWithContext(ctx)
		c.SetBody(rename)
		c.SetID(resource.ID.String())

		res, err := marketplaceClient.Resource.PatchResourcesID(c, nil)
		if err != nil {
			panic(err)
		}

		fmt.Sprintf("%#v", res)
	}

	return nil, nil, nil
}
