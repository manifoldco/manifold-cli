package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	catalogcache "github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/analytics"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	mClient "github.com/manifoldco/manifold-cli/generated/marketplace/client"
	resClient "github.com/manifoldco/manifold-cli/generated/marketplace/client/resource"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

func init() {
	updateCmd := cli.Command{
		Name:      "update",
		ArgsUsage: "[label]",
		Usage:     "Allows a user to update a resource in Manifold",
		Action:    middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs, updateResourceCmd),
		Flags: []cli.Flag{
			nameFlag(),
			appFlag(),
			planFlag(),
			skipFlag(),
		},
	}

	cmds = append(cmds, updateCmd)
}

func updateResourceCmd(cliCtx *cli.Context) error {
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
		return cli.NewExitError(fmt.Sprintf("Failed to create Provisioning Client: %s", err), -1)
	}

	res, err := clients.FetchResources(ctx, marketplaceClient)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	if len(res) == 0 {
		return cli.NewExitError("No resources found for updating", -1)
	}

	var resource *mModels.Resource
	if resourceLabel != "" {
		var err error
		resource, err = pickResourcesByLabel(res, resourceLabel)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to fetch resource: %s", err), -1)
		}
	} else {
		resourceIdx, _, err := prompts.SelectResource(res, resourceLabel)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		resource = res[resourceIdx]
	}

	newResourceName := cliCtx.String("name")
	resourceName := string(resource.Body.Name)
	autoSelect := false
	if newResourceName != "" {
		resourceName = newResourceName
		autoSelect = true
	}

	newResourceName, err = prompts.ResourceName(resourceName, autoSelect)
	if err != nil {
		cli.NewExitError(fmt.Sprintf("Could not rename the resource: %s", err), -1)
	}

	catalog, err := catalogcache.New(ctx, catalogClient)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to catalog data: %s", err), -1)
	}

	// TODO: Move this+fetchUniqueAppNames from create.go into another file/package?
	plans := filterPlansByProductID(catalog.Plans(), resource.Body.ProductID)
	planLabel := cliCtx.String("plan")
	if planLabel != "" {
		l := manifold.Label(planLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidPlanLabel)
		}
	} else {
		plan, err := pickPlanByID(plans, resource.Body.PlanID)
		if err != nil {
			return cli.NewExitError("Could not find provided plan", -1)
		}
		planLabel = string(plan.Body.Label)
	}

	planIdx, _, err := prompts.SelectPlan(plans, planLabel, true)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select plan")
	}

	appName := cliCtx.String("app")
	if appName != "" {
		n := manifold.Name(appName)
		if err := n.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidAppName)
		}
	} else {
		appName = string(resource.Body.AppName)
	}

	apps := fetchUniqueAppNames(res)
	// TODO: This auto-selects the app and doesn't let the user change it without the -a flag
	_, appName, err = prompts.SelectCreateAppName(apps, appName, true)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select apps")
	}

	spin := spinner.New(spinner.CharSets[38], 500*time.Millisecond)
	if !dontWait {
		fmt.Printf("\nWe're starting to update the resource \"%s\". This may take some time, please wait!\n\n",
			resource.Body.Label,
		)
		spin.Start()
	}

	_, mrb, err := updateResource(ctx, cfg, s, resource, marketplaceClient, provisioningClient,
		plans[planIdx], appName, newResourceName, dontWait,
	)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to update resource: %s", err), -1)
	}

	if !dontWait {
		spin.Stop()
	}

	fmt.Printf("Your instance \"%s\" has been updated\n", mrb.Body.Name)

	return nil
}

func pickResourcesByLabel(resources []*mModels.Resource, resourceLabel string) (*mModels.Resource, error) {
	if resourceLabel == "" {
		return nil, errs.ErrResourceNotFound
	}

	for _, resource := range resources {
		if string(resource.Body.Label) == resourceLabel {
			return resource, nil
		}
	}

	return nil, errs.ErrResourceNotFound
}

func pickPlanByID(plans []*cModels.Plan, id manifold.ID) (*cModels.Plan, error) {
	for _, p := range plans {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, errs.ErrPlanNotFound
}

func updateResource(ctx context.Context, cfg *config.Config, s session.Session, resource *mModels.Resource,
	marketplaceClient *mClient.Marketplace, provisioningClient *pClient.Provisioning, plan *cModels.Plan,
	appName, resourceName string, dontWait bool,
) (*pModels.Operation, *mModels.Resource, error) {
	a, err := analytics.New(cfg, s)
	if err != nil {
		return nil, nil, err
	}

	label := strings.Replace(strings.ToLower(resourceName), " ", "-", -1)
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

	patchRes, err := marketplaceClient.Resource.PatchResourcesID(c, nil)
	if err != nil {
		switch e := err.(type) {
		case *resClient.PatchResourcesIDBadRequest:
			return nil, nil, e.Payload
		case *resClient.PatchResourcesIDUnauthorized:
			return nil, nil, e.Payload
		case *resClient.PatchResourcesIDInternalServerError:
			return nil, nil, errs.ErrSomethingWentHorriblyWrong
		}
	}

	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return nil, nil, err
	}

	typeStr := "operation"
	version := int64(1)
	state := "resize"
	curTime := strfmt.DateTime(time.Now())
	op := &pModels.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &pModels.Resize{
			PlanID: plan.ID,
			State:  &state,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)
	op.Body.SetResourceID(resource.ID)
	op.Body.SetUserID(&s.User().ID)

	p := operation.NewPutOperationsIDParamsWithContext(ctx)
	p.SetBody(op)
	p.SetID(ID.String())

	opRes, err := provisioningClient.Operation.PutOperationsID(p, nil)
	if err != nil {
		switch e := err.(type) {
		case *operation.PutOperationsIDBadRequest:
			return nil, nil, e.Payload
		case *operation.PutOperationsIDUnauthorized:
			return nil, nil, e.Payload
		case *operation.PutOperationsIDNotFound:
			return nil, nil, e.Payload
		case *operation.PutOperationsIDConflict:
			return nil, nil, e.Payload
		case *operation.PutOperationsIDInternalServerError:
			return nil, nil, errs.ErrSomethingWentHorriblyWrong
		default:
			return nil, nil, err
		}
	}

	params := map[string]string{
		"plan":  string(plan.Body.Label),
		"price": toPrice(*plan.Body.Cost),
	}
	a.Track(ctx, "Update Resource", &params)

	if dontWait {
		return opRes.Payload, patchRes.Payload, err
	}

	opModel, err := waitForOp(ctx, provisioningClient, opRes.Payload)
	return opModel, patchRes.Payload, err
}
