package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	catalogcache "github.com/manifoldco/manifold-cli/data/catalog"
	"github.com/manifoldco/manifold-cli/errs"
	cModels "github.com/manifoldco/manifold-cli/generated/catalog/models"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
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

	client, err := api.New(api.Catalog, api.Marketplace, api.Provisioning)
	if err != nil {
		return err
	}

	catalog, err := catalogcache.New(ctx, client.Catalog)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load client catalog: %s", err), -1)
	}

	res, err := clients.FetchResources(ctx, client.Marketplace, teamID, cliCtx.String("project"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not load resources: %s", err), -1)
	}
	if len(res) == 0 {
		return errs.ErrNoResources
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
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

	if err := resizeResource(ctx, r, p, client, teamID, userID, dontWait); err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not update resource \"%s\": %s", string(r.Body.Label), err), -1)
	}

	fmt.Printf("\nYour resource %q has been resized to the plan %q\n", r.Body.Label, p.Body.Label)

	return nil
}

func resizeResource(ctx context.Context, r *mModels.Resource, p *cModels.Plan,
	client *api.API, tid, uid *manifold.ID, dontWait bool,
) error {
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
			PlanID:     p.ID,
			ResourceID: r.ID,
			State:      &state,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)

	if tid == nil {
		op.Body.SetUserID(uid)
	} else {
		op.Body.SetTeamID(tid)
	}

	resize := operation.NewPutOperationsIDParamsWithContext(ctx)
	resize.SetBody(op)
	resize.SetID(ID.String())

	res, err := client.Provisioning.Operation.PutOperationsID(resize, nil)
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

	client.Analytics.Track(ctx, "Resize Operation", &aParams)

	if dontWait {
		return nil
	}

	_, err = waitForOp(ctx, client.Provisioning, res.Payload)
	return err
}
