package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-openapi/strfmt"
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	provisioning "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

func init() {
	createCmd := cli.Command{
		Name:      "create",
		ArgsUsage: "[product] [name]",
		Usage:     "Allows a user to create a new resource through Manifold.",
		Action:    chain(loadDirPrefs, create),
		Flags: []cli.Flag{
			appFlag(),
			planFlag(),
			skipFlag(),
			// TODO: Support a region flag
		},
	}

	cmds = append(cmds, createCmd)
}

func create(cliCtx *cli.Context) error {
	ctx := context.Background()
	args := cliCtx.Args()

	dontWait := cliCtx.Bool("no-wait")
	appName := cliCtx.String("app")
	if appName != "" {
		name := manifold.Name(appName)
		if err := name.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidAppName)
		}
	}

	planLabel := cliCtx.String("plan")
	if planLabel != "" {
		l := manifold.Label(planLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidPlanLabel)
		}
	}

	productLabel := ""
	resourceName := ""
	if len(args) > 3 {
		return errs.NewUsageExitError(cliCtx, errs.ErrTooManyArgs)
	}

	if len(args) > 0 {
		productLabel = args[0]
		l := manifold.Label(productLabel)
		if err := l.Validate(nil); err != nil {
			return errs.NewUsageExitError(cliCtx, errs.ErrInvalidProductLabel)
		}

		if len(args) == 2 {
			resourceName = args[1]
			l := manifold.Name(resourceName)
			if err := l.Validate(nil); err != nil {
				return errs.NewUsageExitError(cliCtx, errs.ErrInvalidResourceName)
			}
		}
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

	cClient, err := clients.NewCatalog(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Catalog API client: "+
			err.Error(), -1)
	}

	mClient, err := clients.NewMarketplace(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create Marketplace API client: "+
			err.Error(), -1)
	}

	pClient, err := clients.NewProvisioning(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create a Provisioning API client: "+
			err.Error(), -1)
	}

	catalog, err := catalog.New(ctx, cClient)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	products := catalog.Products()
	productIdx, _, err := prompts.SelectProduct(products, productLabel)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select product.")
	}

	plans := filterPlansByProductID(catalog.Plans(), products[productIdx].ID)
	planIdx, _, err := prompts.SelectPlan(plans, planLabel, false)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select plan.")
	}

	regions := filterRegionsForPlan(catalog.Regions(), plans[planIdx].Body.Regions)
	regionIdx, _, err := prompts.SelectRegion(regions)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select region.")
	}

	// Get resources, so we can fetch the list of valid appnames
	res, err := clients.FetchResources(ctx, mClient)
	if err != nil {
		return cli.NewExitError("Failed to fetch resource list: "+err.Error(), -1)
	}

	appNames := fetchUniqueAppNames(res)
	newA, appName, err := prompts.SelectCreateAppName(appNames, appName, false)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select app.")
	}
	if newA == -1 {
		// TODO: create app name that doesn't exist yet
		// https://github.com/manifoldco/engineering/issues/2614
		return cli.NewExitError("Whoops! A new app cannot be created without a resource", -1)
	}

	resourceName, err = prompts.ResourceName(resourceName, false)
	if err != nil {
		return cli.NewExitError("Could not name the resource: "+err.Error(), -1)
	}

	if dontWait {
		op, err := createResource(ctx, cfg, s, pClient, products[productIdx], plans[planIdx],
			regions[regionIdx], appName, resourceName, true)
		if err != nil {
			return cli.NewExitError("Could not create resource: "+err.Error(), -1)
		}

		provision := op.Body.(*pModels.Provision)
		fmt.Printf("\nAn instance of %s named \"%s\" is being created!\n",
			products[productIdx].Body.Name, *provision.Name)
		return nil
	}

	fmt.Printf("\nWe're starting to create an instance of %s."+
		" This may take some time, please wait!\n\n", products[productIdx].Body.Name)

	spin := spinner.New(spinner.CharSets[38], 500*time.Millisecond)
	spin.Start()
	op, err := createResource(ctx, cfg, s, pClient, products[productIdx], plans[planIdx],
		regions[regionIdx], appName, resourceName, dontWait)
	spin.Stop()
	if err != nil {
		return cli.NewExitError("Could not create resource: "+err.Error(), -1)
	}

	provision := op.Body.(*pModels.Provision)
	fmt.Printf("An instance named \"%s\" has been created!\n", *provision.Name)
	return nil
}

func createResource(ctx context.Context, cfg *config.Config, s session.Session,
	pClient *provisioning.Provisioning, product *cModels.Product, plan *cModels.Plan,
	region *cModels.Region, appName, resourceName string, dontWait bool) (*pModels.Operation, error) {

	a, err := analytics.New(cfg, s)
	if err != nil {
		return nil, err
	}

	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return nil, err
	}

	resourceID, err := manifold.NewID(idtype.Resource)
	if err != nil {
		return nil, err
	}

	// TODO: Generate a label from the name if name provided..?
	// TODO: Expose the Operation primitive from the core marketplace code base into
	// go-manifold so we can use it here.
	typeStr := "operation"
	version := int64(1)
	state := "provision"
	empty := ""
	curTime := strfmt.DateTime(time.Now())
	op := &pModels.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &pModels.Provision{
			AppName:   appName,
			Label:     &empty,
			Name:      &resourceName,
			PlanID:    plan.ID,
			ProductID: product.ID,
			RegionID:  region.ID,
			State:     &state,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)
	op.Body.SetResourceID(resourceID)
	op.Body.SetUserID(&s.User().ID)

	p := operation.NewPutOperationsIDParamsWithContext(ctx)
	p.SetBody(op)
	p.SetID(ID.String())

	res, err := pClient.Operation.PutOperationsID(p, nil)
	if err != nil {
		switch e := err.(type) {
		case *operation.PutOperationsIDBadRequest:
			return nil, e.Payload
		case *operation.PutOperationsIDUnauthorized:
			return nil, e.Payload
		case *operation.PutOperationsIDNotFound:
			return nil, e.Payload
		case *operation.PutOperationsIDConflict:
			return nil, e.Payload
		case *operation.PutOperationsIDInternalServerError:
			return nil, errs.ErrSomethingWentHorriblyWrong
		default:
			return nil, err
		}
	}

	params := map[string]string{
		"product": string(product.Body.Label),
		"plan":    string(plan.Body.Label),
		"price":   toPrice(*plan.Body.Cost),
		"region":  string(*region.Body.Location),
	}
	a.Track(ctx, "Provision Operation", &params)
	if dontWait {
		return res.Payload, nil
	}

	return waitForOp(ctx, pClient, res.Payload)
}

func waitForOp(ctx context.Context, pClient *provisioning.Provisioning, op *pModels.Operation) (*pModels.Operation, error) {
	checkOp := func() (*pModels.Operation, error) {
		p := operation.NewGetOperationsIDParams()
		p.SetContext(ctx)
		p.SetID(op.ID.String())

		res, err := pClient.Operation.GetOperationsID(p, nil)
		if err != nil {
			switch e := err.(type) {
			case *operation.GetOperationsIDBadRequest:
				return nil, e.Payload
			case *operation.GetOperationsIDUnauthorized:
				return nil, e.Payload
			case *operation.GetOperationsIDNotFound:
				return nil, e.Payload
			case *operation.GetOperationsIDInternalServerError:
				return nil, e.Payload
			default:
				return nil, err
			}
		}

		return res.Payload, nil
	}

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		<-ticker.C
		op, err := checkOp()
		if err != nil {
			return nil, err
		}

		switch provision := op.Body.(type) {
		case *pModels.Provision:
			switch *provision.State {
			case "done":
				return op, nil
			case "error":
				return nil, fmt.Errorf("Error completing provision")
			default:
				continue
			}
		case *pModels.Resize:
			switch *provision.State {
			case "done":
				return op, nil
			case "error":
				return nil, fmt.Errorf("Error completing resize")
			default:
				continue
			}
		case *pModels.Deprovision:
			switch *provision.State {
			case "done":
				return op, nil
			case "error":
				return nil, fmt.Errorf("Error completing delete")
			default:
				continue
			}
		default:
			return nil, fmt.Errorf("Unknown provision operation")
		}

	}
}

func filterPlansByProductID(plans []*cModels.Plan, productID manifold.ID) []*cModels.Plan {
	out := make([]*cModels.Plan, 0, len(plans))
	for _, p := range plans {
		if p.Body.ProductID == productID {
			out = append(out, p)
		}
	}

	return out
}

func filterRegionsForPlan(regions []*cModels.Region, regionIDs []manifold.ID) []*cModels.Region {
	out := make([]*cModels.Region, 0, len(regionIDs))
	idx := make(map[manifold.ID]bool)
	for _, r := range regionIDs {
		idx[r] = true
	}

	for _, r := range regions {
		_, ok := idx[r.ID]
		if ok {
			out = append(out, r)
		}
	}

	return out
}

func fetchUniqueAppNames(resources []*mModels.Resource) []string {
	out := []string{}

	// TODO: Make this not awful :(
	exists := func(name string) bool {
		for _, n := range out {
			if n == name {
				return true
			}
		}

		return false
	}

	for _, r := range resources {
		name := string(r.Body.AppName)
		if !exists(name) && name != "" {
			out = append(out, name)
		}
	}

	return out
}

func toPrice(cost int64) string {
	s := strconv.Itoa(int(cost))
	if len(s) == 0 {
		return "0.00"
	}

	if len(s) == 1 {
		return "0.0" + s
	}

	if len(s) == 2 {
		return "0." + s
	}

	return s[:len(s)-2] + "." + s[len(s)-2:]
}
