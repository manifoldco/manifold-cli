package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"

	manifold "github.com/manifoldco/go-manifold"
	"github.com/manifoldco/go-manifold/idtype"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/clients"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/middleware"
	"github.com/manifoldco/manifold-cli/prompts"

	iClient "github.com/manifoldco/manifold-cli/generated/identity/client"
	"github.com/manifoldco/manifold-cli/generated/marketplace/models"
	"github.com/manifoldco/manifold-cli/generated/provisioning/client/operation"
	pModels "github.com/manifoldco/manifold-cli/generated/provisioning/models"
)

func init() {
	createCmd := cli.Command{
		Name:      "transfer",
		ArgsUsage: "[owner]",
		Usage:     "Transfer a resource",
		Category:  "RESOURCES",
		Action: middleware.Chain(middleware.LoadDirPrefs, middleware.EnsureSession,
			middleware.LoadTeamPrefs, transfer),
		Flags: append(teamFlags, []cli.Flag{
			resourceFlag(),
			projectFlag(),
			skipFlag(),
			cli.StringFlag{
				Name:  "owner, o",
				Usage: "The new owner for the resource. This can either be a team title you belong to or an admin email from one of the teams you belong to",
			},
		}...),
	}

	cmds = append(cmds, createCmd)
}

func transfer(cliCtx *cli.Context) error {
	ctx := context.Background()

	if err := exactArgsLength(cliCtx, 1); err != nil {
		return err
	}

	newOwner, err := requiredArgName(cliCtx, 0, "owner")
	if err != nil {
		return err
	}

	resourceName, err := validateName(cliCtx, "resource")
	if err != nil {
		return err
	}

	userID, userIDErr := loadUserID(ctx)
	if userIDErr != nil && userIDErr != errUserActionAsTeam {
		return userIDErr
	}

	teamID, err := validateTeamID(cliCtx)
	if err != nil {
		return err
	}

	project, err := validateName(cliCtx, "project")
	if err != nil {
		return err
	}

	client, err := api.New(api.Marketplace, api.Identity, api.Provisioning, api.Analytics)
	if err != nil {
		return err
	}

	var resources []*models.Resource

	resources, err = clients.FetchResources(ctx, client.Marketplace, teamID, project)

	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch the list of provisioned resources: %s", err), -1)
	}

	if len(resources) == 0 {
		return cli.NewExitError("No resources found for transfering", -1)
	}

	projects, err := clients.FetchProjects(ctx, client.Marketplace, teamID)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to fetch list of projects: %s", err), -1)
	}

	var resource *models.Resource
	if resourceName != "" {
		var err error
		resource, err = pickResourcesByName(resources, projects, resourceName)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to fetch resource: %s", err), -1)
		}
	} else {
		idx, _, err := prompts.SelectResource(resources, projects, resourceName)
		if err != nil {
			return prompts.HandleSelectError(err, "Could not select Resource")
		}
		resource = resources[idx]
	}

	newOwnerID, err := newOwnerID(ctx, newOwner, client.Identity)
	if err != nil {
		return err
	}

	dontWait := cliCtx.Bool("no-wait")
	if err := transferResource(ctx, client, userID, teamID, resource.ID, *newOwnerID, dontWait); err != nil {
		return err
	}

	fmt.Printf("Resource \"%s\" has been transfered to \"%s\"\n", resource.Body.Label, newOwner)

	return nil
}

func newOwnerID(ctx context.Context, newOwner string, ic *iClient.Identity) (*manifold.ID, error) {
	teams, err := clients.FetchTeams(ctx, ic)
	if err != nil {
		return nil, err
	}

	isEmail := govalidator.IsEmail(newOwner)
	for _, t := range teams {
		if isEmail {
			members, err := clients.FetchTeamMembers(ctx, t.ID.String(), ic)
			if err != nil {
				return nil, err
			}

			for _, m := range members {
				if string(m.Email) == newOwner {
					return &m.UserID, nil
				}
			}
		} else if string(t.Body.Label) == newOwner {
			return &t.ID, nil
		}
	}

	return nil, errors.New("Could not find team or user with the provided label or email")
}

func transferResource(ctx context.Context, client *api.API,
	uID, tID *manifold.ID,
	resourceID, newOwnerID manifold.ID,
	dontWait bool) error {
	ID, err := manifold.NewID(idtype.Operation)
	if err != nil {
		return err
	}

	typeStr := "operation"
	version := int64(1)
	state := "transfer"
	curTime := strfmt.DateTime(time.Now())
	op := &pModels.Operation{
		ID:      ID,
		Type:    &typeStr,
		Version: &version,
		Body: &pModels.Transfer{
			State:      &state,
			ResourceID: resourceID,
			NewOwnerID: &newOwnerID,
		},
	}

	op.Body.SetCreatedAt(&curTime)
	op.Body.SetUpdatedAt(&curTime)

	if tID == nil {
		op.Body.SetUserID(uID)
	} else {
		op.Body.SetTeamID(tID)
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

	if dontWait {
		return nil
	}

	_, err = waitForOp(ctx, client.Provisioning, res.Payload)
	return err
}
