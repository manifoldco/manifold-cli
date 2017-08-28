package main

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"

	"github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	mModels "github.com/manifoldco/manifold-cli/generated/marketplace/models"
	pClient "github.com/manifoldco/manifold-cli/generated/provisioning/client"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"
)

func init() {
	deleteCmd := cli.Command{
		Name:      "delete",
		ArgsUsage: "[name]",
		Usage:     "Delete a resource",
		Action:    middleware.Chain(middleware.EnsureSession, deleteCmd),
		Flags: []cli.Flag{
			skipFlag(),
		},
	}

	cmds = append(cmds, deleteCmd)
}

func deleteCmd(cliCtx *cli.Context) error {
	ctx := context.Background()

	dontWait := cliCtx.Bool("no-wait")

	if err := maxOptionalArgsLength(cliCtx, 1); err != nil {
		return err
	}

	resourceLabel, err := optionalArgLabel(cliCtx, 0, "resource")
	if err != nil {
		return err
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
		return cli.NewExitError(fmt.Sprintf("Failed to create Maketplace Client: %s", err), -1)
	}

	provisioningClient, err := clients.NewProvisioning(cfg)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create Provision Client: %s", err), -1)
	}

	res, err := clients.FetchResources(ctx, marketplaceClient)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	var resource *mModels.Resource
	if resourceLabel != "" {
		resource, err = pickResourcesByLabel(res, resourceLabel)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Failed to find resource \"%s\": %s", resourceLabel, err), -1)
		}
	} else {
		resourceIdx, _, err := prompts.SelectResource(res, resourceLabel)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}

		resource = res[resourceIdx]
	}

	if _, err := prompts.Confirm(
		fmt.Sprintf("Are you sure you want to delete \"%s\"", resource.Body.Label)); err != nil {
		return cli.NewExitError("Resource not deleted", -1)
	}

	spin := spinner.New(spinner.CharSets[38], 500*time.Millisecond)
	if !dontWait {
		fmt.Printf("\nWe're starting to delete the resource \"%s\". This may take some time, please wait!\n\n",
			resource.Body.Label,
		)
		spin.Start()
	}

	err = deleteResource(ctx, cfg, s, resource, provisioningClient, dontWait)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to delete resource: %s", err), -1)
	}

	if !dontWait {
		spin.Stop()
	}

	fmt.Printf("Your instance \"%s\" has been deleted\n", resource.Body.Label)

	return nil
}

func deleteResource(ctx context.Context, cfg *config.Config, s session.Session, resource *mModels.Resource,
	provisioningClient *pClient.Provisioning, dontWait bool,
) error {

	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return err
	}

	typeStr := "operation"
	version := int64(1)
	state := "deprovision"
	curTime := strfmt.DateTime(time.Now())
	op := &pModels.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &pModels.Deprovision{
			State: &state,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)
	op.Body.SetUserID(&s.User().ID)
	op.Body.SetResourceID(resource.ID)

	d := operation.NewPutOperationsIDParamsWithContext(ctx)
	d.SetBody(op)
	d.SetID(ID.String())

	opRes, err := provisioningClient.Operation.PutOperationsID(d, nil)
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

	if dontWait {
		return nil
	}

	_, err = waitForOp(ctx, provisioningClient, opRes.Payload)
	return err
}
