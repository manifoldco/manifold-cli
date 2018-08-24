package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/juju/ansiterm"
	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/color"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"

	"github.com/manifoldco/manifold-cli/generated/billing/client/profile"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	provisioning "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

func init() {
	createCmd := cli.Command{
		Name:      "create",
		ArgsUsage: "[resource-name]",
		Usage:     "Create a new resource",
		Category:  "RESOURCES",
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
			middleware.LoadTeamPrefs, create),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
			planFlag(),
			regionFlag(),
			cli.StringFlag{
				Name:  "product",
				Usage: "Create a resource for this product",
			},
			cli.BoolFlag{
				Name:  "custom, c",
				Usage: "Create a custom resource, for holding custom configuration",
			},
			skipFlag(),
		}...),
	}

	cmds = append(cmds, createCmd)
}

func create(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	dontWait := cliCtx.Bool("no-wait")
	projectName, err := validateName(cliCtx, "project")
	if err != nil {
		return err
	}

	planName, err := validateName(cliCtx, "plan")
	if err != nil {
		return err
	}

	productName, err := validateName(cliCtx, "product")
	if err != nil {
		return err
	}

	regionName, err := validateName(cliCtx, "region")
	if err != nil {
		return err
	}

	custom := cliCtx.Bool("custom")
	if custom && (planName != "" || productName != "" || regionName != "") {
		return errs.NewUsageExitError(cliCtx, cli.NewExitError(
			"You cannot specify product options for a custom resource", -1,
		))
	}

	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError("Could not load config: "+err.Error(), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		return cli.NewExitError("Could not retrieve session: "+err.Error(), -1)
	}

	client, err := api.New(api.Catalog, api.Marketplace, api.Provisioning)
	if err != nil {
		return err
	}

	catalog, err := catalog.New(ctx, client.Catalog)
	if err != nil {
		return cli.NewExitError("Failed to fetch catalog data: "+err.Error(), -1)
	}

	var product *cModels.Product
	var plan *cModels.Plan
	var region *cModels.Region

	if !custom {
		products := catalog.Products()
		productIdx, _, err := prompts.SelectProduct(products, productName)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select product.")
		}

		if planName != "" {
			plan, err = catalog.FetchPlanByLabel(ctx, products[productIdx].ID, planName)
			if err != nil {
				return prompts.HandleSelectError(err, "Plan does not exist.")
			}
			_, _, err = prompts.SelectPlan([]*cModels.Plan{plan}, planName)
			if err != nil {
				return prompts.HandleSelectError(err, "Could not select plan.")
			}
		} else {
			selectedPlan, err := selectPlanForCreate(ctx, teamID, catalog.Plans(), products[productIdx].ID, planName)
			if err != nil {
				switch err {
				case errAbortSelectPlan:
					return err
				default:
					return prompts.HandleSelectError(err, "Could not select plan.")
				}
			}
			plan = selectedPlan
		}

		regions := filterRegionsForPlan(catalog.Regions(), plan.Body.Regions)
		regionIdx, _, err := prompts.SelectRegion(regions)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select region.")
		}

		product = products[productIdx]
		region = regions[regionIdx]
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError("Failed to fetch projects list: "+err.Error(), -1)
	}

	var project *mModels.Project

	if len(projects) > 0 {
		pidx, _, err := prompts.SelectProject(projects, projectName, true, true)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select project.")
		}

		if pidx > -1 {
			project = projects[pidx]
		}
	}

	descriptor := "a custom resource"
	if !custom {
		descriptor = "an instance of " + string(product.Body.Name)
	}

	resourceID, err := manifold.NewID(idtype.Resource)
	if err != nil {
		return cli.NewExitError("Could not create resource: "+err.Error(), -1)
	}

	var resourceName string
	if !custom {
		resourceName, err = promptNameForResource(cliCtx, &resourceID, &product.Body.Label)
	} else {
		resourceName, err = promptNameForResource(cliCtx, &resourceID, nil)
	}
	if err != nil {
		return err
	}
	resourceTitle := resourceName

	spin := prompts.NewSpinner(fmt.Sprintf("Creating %s", descriptor))
	if !dontWait {
		spin.Start()
		defer spin.Stop()
	}

	op, err := createResource(ctx, cfg, &resourceID, teamID, s, client.Provisioning,
		custom, product, plan, region, project, resourceName, resourceTitle, dontWait)
	if err != nil {
		return cli.NewExitError("Could not create resource: "+err.Error(), -1)
	}

	provision := op.Body.(*pModels.Provision)
	if !dontWait {
		spin.Stop()
		fmt.Printf("An instance named \"%s\" has been created!\n", *provision.Name)
		return nil
	}

	if custom {
		fmt.Printf("\nA custom resource named \"%s\" is being created!\n", *provision.Name)
	} else {
		fmt.Printf("\nAn instance of %s named \"%s\" is being created!\n",
			product.Body.Name, *provision.Name)
	}

	return nil
}

var errAbortSelectPlan = cli.NewExitError("Resource creation aborted.", -1)

func selectPlanForCreate(ctx context.Context, teamID *manifold.ID, plans []*cModels.Plan, product manifold.ID, planName string) (*cModels.Plan, error) {
	filteredPlans := filterPlansByProductID(plans, product)
	planIdx, _, err := prompts.SelectPlan(filteredPlans, planName)
	if err != nil {
		return nil, prompts.HandleSelectError(err, "Could not select plan.")
	}
	plan := filteredPlans[planIdx]

	// If the plan is not free, check if they have a billing profile
	if (*plan.Body.Cost) > 0 {
		_, err := retrieveBillingProfile(ctx)
		if err != nil {
			switch err.(type) {
			case *profile.GetProfilesIDNotFound:
				fmt.Println("")
				fmt.Println(color.Color(ansiterm.White, "You chose a plan which is not free and have not added your billing details."))
				fmt.Println("What would you like to do?")
				// No billing information on file, ask how they wish to proceed
				action, err := prompts.SelectBillingProfileAction()
				if err != nil {
					return nil, err
				}
				switch action {
				case "Add credit card":
					userID, userIDErr := loadUserID(ctx)
					if userIDErr != nil && userIDErr != errUserActionAsTeam {
						return nil, userIDErr
					}

					// Prompt the user to add their payment information
					err := createNewBillingProfile(ctx, userID, teamID)
					if err != nil {
						return nil, err
					}
				case "Select different plan":
					return selectPlanForCreate(ctx, teamID, plans, product, planName)
				default:
					return nil, errAbortSelectPlan
				}

			default:
				return nil, cli.NewExitError("Failed to retrieve billing profile: "+err.Error(), -1)
			}
		}
	}
	return plan, nil
}

func createResource(ctx context.Context, cfg *config.Config, resourceID, teamID *manifold.ID, s session.Session,
	pClient *provisioning.Provisioning, custom bool, product *cModels.Product, plan *cModels.Plan,
	region *cModels.Region, project *mModels.Project, resourceName, resourceTitle string, dontWait bool) (*pModels.Operation, error) {

	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return nil, err
	}

	var planID, productID, regionID *manifold.ID
	source := "custom"
	if !custom {
		planID = &plan.ID
		productID = &product.ID
		regionID = &region.ID
		source = "catalog"
	}

	var projectID *manifold.ID
	if project != nil {
		projectID = &project.ID
	}

	// TODO: Generate a label from the name if name provided..?
	// TODO: Expose the Operation primitive from the core marketplace code base into
	// go-manifold so we can use it here.
	typeStr := "operation"
	version := int64(1)
	state := "provision"
	curTime := strfmt.DateTime(time.Now())
	op := &pModels.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &pModels.Provision{
			ResourceID: *resourceID,
			Label:      &resourceName,
			Name:       &resourceTitle,
			Source:     &source,
			PlanID:     planID,
			ProductID:  productID,
			RegionID:   regionID,
			State:      &state,
			ProjectID:  projectID,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)
	if teamID == nil {
		if !s.IsUser() {
			return nil, errUserActionAsTeam
		}
		op.Body.SetUserID(&s.User().ID)
	} else {
		op.Body.SetTeamID(teamID)
	}

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
		case *pModels.Move:
			switch *provision.State {
			case "done":
				return op, nil
			case "error":
				return nil, fmt.Errorf("Error completing move")
			}
		case *pModels.ProjectDelete:
			switch *provision.State {
			case "done":
				return op, nil
			case "error":
				return nil, fmt.Errorf("Error completing project delete")
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
