package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/clients"
	catalogcache "github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	"github.com/manifoldco/manifold-cli/generated/provisioning/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
)

func init() {
	resizeCmd := cli.Command{
		Name:      "resize",
		ArgsUsage: "[label]",
		Usage:     "Resize a resource",
		Category:  "RESOURCES",
		Action: middleware.Chain(middleware.EnsureSession, middleware.LoadDirPrefs, middleware.LoadTeamPrefs,
			resizeResourceCmd),
		Flags: append(teamFlags, []cli.Flag{
			projectFlag(),
			skipFlag(),
			planFlag(),
		}...),
	}

	cmds = append(cmds, resizeCmd)
}

func resizeResourceCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	label, err := optionalArgLabel(cliCtx, 0, "resource")
	if err != nil {
		return err
	}

	userID, err := loadUserID(ctx)
	if err != nil {
		return err
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	dontWait := cliCtx.Bool("no-wait")
	planLabel := cliCtx.String("plan")

	marketplaceClient, err := loadMarketplaceClient()
	if err != nil {
		return err
	}

	provisioningClient, err := loadProvisioningClient()
	if err != nil {
		return err
	}

	catalogClient, err := loadCatalogClient()
	if err != nil {
		return err
	}

	catalog, err := catalogcache.New(ctx, catalogClient)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load client catalog: %s", err), -1)
	}

	res, err := clients.FetchResources(ctx, marketplaceClient, teamID, cliCtx.String("project"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load resources: %s", err), -1)
	}
	if len(res) == 0 {
		return errs.ErrNoResources
	}

	projects, err := clients.FetchProjects(ctx, marketplaceClient, teamID)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load projects: %s", err), -1)
	}
	rIdx, _, err := prompts.SelectResource(res, projects, label)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select resource")
	}
	r := res[rIdx]

	plans := filterPlansByProductID(catalog.Plans(), *r.Body.ProductID)
	if len(plans) == 0 {
		return errs.ErrNoPlans
	}

	pIdx, _, err := prompts.SelectPlan(plans, planLabel, false)
	if err != nil {
		return prompts.HandleSelectError(err, "Could not select Plan")
	}
	p := plans[pIdx]

	spin := prompts.NewSpinner(fmt.Sprintf("Updating resource \"%s\"", r.Body.Label))
	spin.Start()
	defer spin.Stop()

	if err := resizeResource(ctx, r, p, provisioningClient, teamID, userID, dontWait); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not update resource \"%s\": %s", string(r.Body.Label), err), -1)
	}

	fmt.Printf("\nYour resource \"%s\" has been resized to the plan \"%s\"", r.Body.Label, p.Body.Label)

	return nil
}

func resizeResource(ctx context.Context, r *mModels.Resource, p *cModels.Plan,
	pc *client.Provisioning, tid, uid *manifold.ID, dontWait bool,
) error {
	a, err := loadAnalytics()
	if err != nil {
		return err
	}

	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return err
	}

	typeStr := "operation"
	version := int64(1)
	state := "resize"
	curTime := strfmt.DateTime(time.Now())
	op := &models.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &models.Resize{
			PlanID: p.ID,
			State:  &state,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)
	op.Body.SetResourceID(r.ID)

	if tid == nil {
		op.Body.SetUserID(uid)
	} else {
		op.Body.SetTeamID(tid)
	}

	resize := operation.NewPutOperationsIDParamsWithContext(ctx)
	resize.SetBody(op)
	resize.SetID(ID.String())

	res, err := pc.Operation.PutOperationsID(resize, nil)
	if err != nil {
		switch e := err.(type) {
		case *operation.PutOperationsIDBadRequest:
			return e.Payload
		case *operation.PutOperationsIDUnauthorized:
			return e.Payload
		case *operation.PutOperationsIDNotFound:
			return e.Payload
		case *operation.PutOperationsIDConflict:
			return e.Payload
		case *operation.PutOperationsIDInternalServerError:
			return errs.ErrSomethingWentHorriblyWrong
		default:
			return err
		}
	}

	aParams := map[string]string{
		"plan":  string(p.Body.Label),
		"price": toPrice(*p.Body.Cost),
	}

	a.Track(ctx, "Resize Operation", &aParams)

	if dontWait {
		return nil
	}

	_, err = waitForOp(ctx, pc, res.Payload)
	return err
}
